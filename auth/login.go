package auth

import (
	"errors"

	"github.com/mbertschler/foundation"
)

var (
	UsernameFormKey = "username"
	PasswordFormKey = "password"
)

const (
	// somehash is the password "hehehe"
	someHash = "$argon2id$t=3,m=32768,p=4$32J16ZbXQegxU2CU3nOu/lfkno/g+Sv4pZti9LIgfX0$H6lc+0VxTFkPy9yc7z14tHq0bSYknIqmlj66ST67F+"
)

func (h *Handler) Login(r *foundation.Request) error {
	err := r.Request.ParseForm()
	if err != nil {
		return err
	}

	username := r.Request.Form.Get(UsernameFormKey)
	password := r.Request.Form.Get(PasswordFormKey)

	if username == "" || password == "" {
		return errors.New("missing username or password")
	}

	if len(username) > 255 || len(password) > 1024 {
		return errors.New("too long credentials")
	}

	// Check rate limiting BEFORE doing any expensive operations
	if globalRateLimiter.IsBlocked(r, username) {
		return errors.New("too many failed attempts, please try again later")
	}

	user, userErr := h.DB.Users.ByUsername(r.Context, username)

	// We always verify the password, even if the user does not exist
	// to avoid timing attacks, later we check userErr.
	hashedPassword := someHash
	if user != nil {
		hashedPassword = user.HashedPassword
	}

	ok, err := verifyPassword(password, hashedPassword)
	if userErr != nil {
		return userErr
	}
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("invalid password")
	}

	session, err := h.getSessionFromRequest(r)
	if err != nil {
		return err
	}
	if session != nil {
		err = h.DB.Sessions.Delete(r.Context, session.ID)
		if err != nil {
			return err
		}
	}

	session, err = h.DB.Sessions.InsertUserSession(r.Context, user.ID)
	if err != nil {
		return err
	}

	setSessionCookie(r.Writer, session)
	r.Session = session
	r.User = user
	return nil
}

func (h *Handler) Logout(r *foundation.Request) error {
	session, err := h.getSessionFromRequest(r)
	if err != nil {
		return err
	}
	if session != nil {
		err = h.DB.Sessions.Delete(r.Context, session.ID)
		if err != nil {
			return err
		}
	}

	session, err = h.DB.Sessions.InsertAnonymousSession(r.Context)
	if err != nil {
		return err
	}
	r.Session = session
	r.User = nil

	setSessionCookie(r.Writer, session)
	return nil
}
