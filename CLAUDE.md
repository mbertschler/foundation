# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Foundation is an opinionated Go-based web application template using modern tools. It provides a full-stack foundation for building web applications with Go on the backend and a hybrid SSR+SPA approach using Turbo and Stimulus on the frontend.

## Development Setup

Install [Mise](https://mise.jdx.dev/getting-started.html), then run:
```bash
mise setup
mise dev
```

This starts both the Go server and the frontend build watcher. The app will be available at http://localhost:3000.

## Common Commands

**Development:**
- `mise dev` - Start the development server (runs both server and browser watchers in parallel)
- `mise server` - Run the Go server only
- `mise server-watch` - Run the Go server with file watching
- `mise browser-watch` - Watch and rebuild frontend assets

**Building:**
- `mise browser-build` - Build frontend assets (CSS and JS)
- `go build ./cmd/foundation-demo` - Build the Go binary

**Testing & Quality:**
- `mise test` or `go test ./...` - Run Go tests
- `mise lint` or `golangci-lint run --timeout=5m` - Run linter

**Other:**
- `mise tidy` - Clean and tidy Go modules
- `mise cloc` - Count lines of code

## Architecture Overview

### Backend Architecture

**Core Request Flow:**
1. HTTP requests hit `server/server.go` which uses `httprouter` for routing
2. Requests are wrapped in `foundation.Request` which includes context, session, user, and HTTP primitives
3. Routes are rendered via `renderPage()` or `renderFrame()` methods that handle authentication, CSRF protection, and session management
4. Page/frame handler methods (on `pages.Handler`) have access to DB via `h.DB` and return HTML using the `github.com/mbertschler/html` builder

**Key Components:**
- `foundation.Context` - Application-level context containing Config and Broadcast (no DB!)
- `foundation.Request` - Per-request context with Session, User, and HTTP primitives (no DB!)
- `server.Server` - HTTP server with handlers that own the DB reference
- `pages.Handler` - Page handler struct with `DB *db.DB` and `Auth *auth.Handler` fields
- `auth.Handler` - Auth handler struct with `DB *db.DB` field for session/user operations
- `db.DB` - Database struct with concrete repository implementations (Users, Sessions, Links, Visits)

**Database Layer:**
- Uses SQLite via `uptrace/bun` ORM with SQL migrations in `db/migrations/`
- Model structs (User, Session, Link, LinkVisit) are defined in `foundation.go` with `bun` tags
- DB struct and repository implementations are in `db/` package
- DB is initialized in `service.RunApp()` and passed to `server.RunServer()`
- Server creates handlers (pages.Handler, auth.Handler) that own DB references
- Page functions are methods on `pages.Handler`: `func (h *Handler) LinksPage(req)`
- Access DB via handler: `h.DB.Users.ByID(...)` - fully typed, no assertions!
- Migrations are automatically applied on startup via `db/migrations/migrations.go`
- Connection uses WAL mode for better concurrency

**Authentication:**
- Session-based authentication with CSRF protection
- Passwords hashed using Argon2id (see `auth/auth.go`)
- Sessions have automatic rotation via `RotateSessionIfNeeded()`
- CSRF tokens validated via `X-CSRF-TOKEN` header for state-changing requests
- Session middleware in `auth/sessions.go` handles cookie management
- Rate limiting available in `auth/ratelimit.go`

**Real-time Updates:**
- Custom broadcast system in `server/broadcast/broadcast.go` for SSE (Server-Sent Events)
- Listeners can subscribe to named channels
- Used for real-time updates to the UI (e.g., links list updates)

### Frontend Architecture

**Technology Stack:**
- Turbo Drive for page navigation without full reloads
- Stimulus for interactive JavaScript controllers
- Tailwind CSS + basecoat-css for styling
- esbuild for JS bundling, Tailwind CLI for CSS
- PurgeCSS removes unused styles from production builds

**Build Process:**
- `browser/main.js` - Entry point that imports Turbo and Stimulus controllers
- `browser/main.css` - Entry point for Tailwind CSS
- Build outputs to `browser/dist/` which is served via Go's `http.FileServer`
- In dev mode (`-dev` flag), assets are served from disk; in production, they're embedded in the binary

**HTML Rendering:**
- Server-side HTML generation using `github.com/mbertschler/html` package
- Type-safe HTML construction with builder pattern
- Pages defined as `PageFunc` with `Page` struct containing Title, Sidebar, Header, Body
- Frames defined as `FrameFunc` that return `html.Block` for partial updates

### Configuration

Configuration is loaded from `foundation_config.json` with these fields:
- `HostPort` - Server listen address (default: "localhost:3000")
- `DBPath` - SQLite database path (default: "./_data/foundation.db")
- `LitestreamYml` - Path to Litestream config for database replication

The `-dev` flag enables development mode which serves frontend assets from disk instead of embedded files.

## Code Patterns

**Adding a new route:**
1. Define method on `pages.Handler`: `func (h *Handler) MyPage(req *foundation.Request) (*Page, error)`
2. Access DB via `h.DB.Users.ByID(...)` - fully typed!
3. Register route in `server/server.go` `setupPageRoutes()` using `s.pages.MyPage`
4. Use `RequireLogin()` option if authentication is needed
5. For real-time updates, use `renderSSEStreamOnChannel()`

**Adding a new database table:**
1. Create up/down migration SQL files in `db/migrations/`
2. Define model struct in `foundation.go` with `bun` tags
3. Create new file in `db/` package (e.g., `foobars.go`) with repository struct (e.g., `fooBarsDB`)
4. Implement CRUD methods on the repository struct
5. Add repository field to `db.DB` struct in `db/db.go`
6. Initialize repository in `StartDB()` function in `db/db.go`
7. Access in page handlers via `h.DB.FooBars.Method(...)`

**CSRF Protection:**
All POST/PATCH/DELETE/PUT requests require CSRF token in `X-CSRF-TOKEN` header. The token is available in the page via meta tag and can be accessed via `req.CSRFToken()`.

## Project Structure

- `cmd/foundation-demo/` - Main application entry point
- `server/` - HTTP server, routing, and rendering logic
- `pages/` - Page and frame definitions (HTML templates)
- `auth/` - Authentication, sessions, rate limiting
- `db/` - Database implementations and migrations
- `browser/` - Frontend assets (CSS, JS, Stimulus controllers)
- `service/` - Application service layer (startup/shutdown)

## Testing Notes

When writing tests:
- Use standard Go testing (`*testing.T`)
- Database tests should use in-memory SQLite (`:memory:`)
- Auth tests are in `auth/auth_test.go`
