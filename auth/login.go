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

func Login(r *foundation.Request) (*foundation.Session, error) {
	err := r.Request.ParseForm()
	if err != nil {
		return nil, err
	}

	username := r.Request.Form.Get(UsernameFormKey)
	password := r.Request.Form.Get(PasswordFormKey)

	if username == "" || password == "" {
		return nil, errors.New("missing username or password")
	}

	if len(username) > 255 || len(password) > 1024 {
		return nil, errors.New("too long credentials")
	}

	// Check rate limiting BEFORE doing any expensive operations
	if globalRateLimiter.IsBlocked(r, username) {
		return nil, errors.New("too many failed attempts, please try again later")
	}

	user, userErr := r.DB.Users.ByUsername(r.Context, username)
	hashedPassword := someHash
	if user != nil {
		hashedPassword = user.HashedPassword
	}

	ok, err := verifyPassword(password, hashedPassword)
	if userErr != nil {
		return nil, userErr
	}
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("invalid password")
	}

	session, err := getSessionFromRequest(r)
	if err != nil {
		return nil, err
	}
	if session != nil {
		err = r.DB.Sessions.Delete(r.Context, session.ID)
		if err != nil {
			return nil, err
		}
	}

	session, err = r.DB.Sessions.InsertUserSession(r.Context, user.ID)
	if err != nil {
		return nil, err
	}

	setSessionCookie(r.Writer, session)
	return session, nil
}

func Logout(r *foundation.Request) (*foundation.Session, error) {
	session, err := getSessionFromRequest(r)
	if err != nil {
		return nil, err
	}
	if session != nil {
		err = r.DB.Sessions.Delete(r.Context, session.ID)
		if err != nil {
			return nil, err
		}
	}

	session, err = r.DB.Sessions.InsertAnonymousSession(r.Context)
	if err != nil {
		return nil, err
	}

	setSessionCookie(r.Writer, session)
	return session, nil
}
