package auth

import (
	"database/sql"
	"net/http"

	"github.com/mbertschler/foundation"
)

const (
	sessionCookieName = "foundation_session"
)

func (h *Handler) GetOrCreateSession(r *foundation.Request) (*foundation.Session, error) {
	session, err := h.getSessionFromRequest(r)
	if err != nil && err != http.ErrNoCookie && err != sql.ErrNoRows {
		return nil, err
	}
	if session != nil {
		// For user sessions, rotate if needed
		if session.UserID.Valid {
			rotatedSession, err := h.DB.Sessions.RotateSessionIfNeeded(r.Context, session.ID)
			if err != nil {
				return nil, err
			}
			// If session was rotated, update the cookie
			if rotatedSession.ID != session.ID {
				r.PreviousSession = session // keep previous session for CSRF checks
				setSessionCookie(r.Writer, rotatedSession)
			}
			return rotatedSession, nil
		}
		return session, nil
	}

	// Create a new session if none exists
	session, err = h.DB.Sessions.InsertAnonymousSession(r.Context)
	if err != nil {
		return nil, err
	}

	setSessionCookie(r.Writer, session)
	return session, nil
}

func setSessionCookie(w http.ResponseWriter, session *foundation.Session) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	}
	http.SetCookie(w, cookie)
}

func (h *Handler) getSessionFromRequest(r *foundation.Request) (*foundation.Session, error) {
	cookie, err := r.Request.Cookie(sessionCookieName)
	if err != nil {
		return nil, err
	}

	return h.DB.Sessions.ByID(r.Context, cookie.Value)
}
