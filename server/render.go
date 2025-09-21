package server

import (
	"crypto/subtle"
	"errors"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mbertschler/foundation"
	"github.com/mbertschler/foundation/auth"
	"github.com/mbertschler/foundation/pages"
	"github.com/mbertschler/html"
)

var ErrStopRendering = errors.New("stop rendering")

type RenderOption struct {
	BeforeRender func(req *foundation.Request) error
}

func RequireLogin() RenderOption {
	return RenderOption{
		BeforeRender: func(req *foundation.Request) error {
			if req.User == nil {
				http.Redirect(req.Writer, req.Request, "/admin/login", http.StatusFound)
				return ErrStopRendering
			}
			return nil
		},
	}
}

func (s *Server) renderPage(ctx *foundation.Context, fn pages.PageFunc, opts ...RenderOption) httprouter.Handle {
	return s.renderFrame(ctx, func(req *foundation.Request) (html.Block, error) {
		page, err := fn(req)
		if err != nil {
			return nil, err
		}
		if page == nil {
			// some requests like redirects might not return a page
			return nil, nil
		}
		return page.RenderHTML(req), nil
	}, opts...)
}

func (s *Server) renderFrame(ctx *foundation.Context, fn pages.FrameFunc, opts ...RenderOption) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		req := &foundation.Request{
			Context: ctx,
			Writer:  w,
			Request: r,
			Params:  params,
		}

		sess, err := auth.GetOrCreateSession(req)
		if err != nil {
			log.Println("GetOrCreateSession error:", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		req.Session = sess

		// Verify CSRF token for state-changing requests
		if requiresCSRFProtection(r.Method) {
			if err := verifyCSRFToken(req); err != nil {
				log.Println("CSRF verification failed:", err)
				http.Error(w, "CSRF token verification failed", http.StatusForbidden)
				return
			}
		}

		if sess.UserID.Valid {
			user, err := req.DB.Users.ByID(req.Context, sess.UserID.Int64)
			if err != nil {
				log.Println("Users.ByID error:", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			req.User = user
		}

		// in case token was rotated, also make it available to frames and streams
		w.Header().Set("X-CSRF-Token", req.CSRFToken())
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		for _, opt := range opts {
			if opt.BeforeRender != nil {
				err := opt.BeforeRender(req)
				if errors.Is(err, ErrStopRendering) {
					return
				}
				if err != nil {
					log.Println("BeforeRender error:", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		block, err := fn(req)
		if err != nil {
			log.Println("Render error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = html.Render(w, block)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// requiresCSRFProtection returns true if the HTTP method requires CSRF protection
func requiresCSRFProtection(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

// verifyCSRFToken verifies the CSRF token from the X-CSRF-TOKEN header
func verifyCSRFToken(req *foundation.Request) error {
	token := req.Request.Header.Get("X-CSRF-TOKEN")
	if token == "" {
		return errors.New("missing CSRF token header")
	}

	expectedToken := req.CSRFToken()
	if expectedToken == "" {
		return errors.New("no session")
	}

	previousToken := req.PreviousCSRFToken()
	if previousToken != "" {
		expectedToken = previousToken
	}

	if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
		return errors.New("invalid CSRF token")
	}

	return nil
}
