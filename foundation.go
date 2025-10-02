package foundation

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mbertschler/foundation/server/broadcast"
	"github.com/uptrace/bun"
)

type Context struct {
	context.Context
	Config    *Config
	DB        *DB
	Broadcast *broadcast.Broadcaster
}

type Request struct {
	*Context
	Writer  http.ResponseWriter
	Request *http.Request
	Params  httprouter.Params

	Session         *Session
	PreviousSession *Session
	User            *User
}

// CSRFToken returns the CSRF token for this request's current session.
func (r *Request) CSRFToken() string {
	if r.Session == nil {
		return ""
	}
	return r.Session.CSRFToken
}

// PreviousCSRFToken returns the CSRF token for the potentially rotated out session.
func (r *Request) PreviousCSRFToken() string {
	if r.PreviousSession == nil {
		return ""
	}
	return r.PreviousSession.CSRFToken
}

type DB struct {
	Users    UserDB
	Sessions SessionDB
	Links    LinkDB
	Visits   VisitDB

	sqlDB *sql.DB
}

func (db *DB) SetSQLDB(sqlDB *sql.DB) {
	db.sqlDB = sqlDB
}

func (db *DB) Close() error {
	return db.sqlDB.Close()
}

type UserDB interface {
	ByID(ctx context.Context, id int64) (*User, error)
	ByUsername(ctx context.Context, username string) (*User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	Insert(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, userID int64) error
	All(ctx context.Context) ([]*User, error)
}

type SessionDB interface {
	InsertUserSession(ctx context.Context, userID int64) (*Session, error)
	InsertAnonymousSession(ctx context.Context) (*Session, error)
	ByID(ctx context.Context, sessionID string) (*Session, error)
	Delete(ctx context.Context, sessionID string) error
	RotateSessionIfNeeded(ctx context.Context, sessionID string) (*Session, error)
}

type LinkDB interface {
	Insert(ctx context.Context, link *Link) error
	Update(ctx context.Context, link *Link) error
	ByShortLink(ctx context.Context, shortLink string) (*Link, error)
	All(ctx context.Context) ([]*Link, error)
	AllWithVisitCounts(ctx context.Context) ([]*Link, error)
	Delete(ctx context.Context, shortLink string) error
}

type VisitDB interface {
	Insert(ctx context.Context, visit *LinkVisit) error
	CountByLink(ctx context.Context, shortLink string) (int64, error)
}

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID             int64     `bun:"id,pk,autoincrement"`
	DisplayName    string    `bun:"display_name,notnull"`
	UserName       string    `bun:"user_name,unique"`
	HashedPassword string    `bun:"hashed_password,notnull"`
	CreatedAt      time.Time `bun:"created_at,nullzero,notnull"`
	UpdatedAt      time.Time `bun:"updated_at,nullzero,notnull"`
}

type Session struct {
	bun.BaseModel `bun:"table:sessions,alias:s"`

	ID        string        `bun:"id,pk"`
	UserID    sql.NullInt64 `bun:"user_id"`
	CreatedAt time.Time     `bun:"created_at,nullzero,notnull"`
	ExpiresAt time.Time     `bun:"expires_at,nullzero,notnull"`
	CSRFToken string        `bun:"csrf_token,nullzero,notnull"`
}

type Link struct {
	bun.BaseModel `bun:"table:links,alias:l"`

	ShortLink   string       `bun:"short_link,pk"`
	FullURL     string       `bun:"full_url,notnull"`
	UserID      int64        `bun:"user_id,notnull"`
	CreatedAt   time.Time    `bun:"created_at,nullzero,notnull"`
	UpdatedAt   time.Time    `bun:"updated_at,nullzero,notnull"`
	User        *User        `bun:"rel:has-one,join:user_id=id"`
	Visits      []*LinkVisit `bun:"rel:has-many,join:short_link=short_link"`
	VisitsCount int64        `bun:"visits_count,scanonly"`
}

type LinkVisit struct {
	bun.BaseModel `bun:"table:link_visits,alias:lv"`

	ID        int64         `bun:"id,pk,autoincrement"`
	ShortLink string        `bun:"short_link,notnull"`
	UserID    sql.NullInt64 `bun:"user_id"`
	VisitedAt time.Time     `bun:"visited_at,nullzero,notnull"`
}
