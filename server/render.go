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

func (s *Server) renderPage(ctx *foundation.Context, fn pages.PageFunc) httprouter.Handle {
	return s.renderFrame(ctx, func(req *foundation.Request) (html.Block, error) {
		page, err := fn(req)
		if err != nil {
			return nil, err
		}
		return page.RenderHTML(req), nil
	})
}

func (s *Server) renderFrame(ctx *foundation.Context, fn pages.FrameFunc) httprouter.Handle {
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

		block, err := fn(req)
		if err != nil {
			log.Println("Render error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// in case token was rotated, also make it available to frames and streams
		w.Header().Set("X-CSRF-Token", req.CSRFToken())

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
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
