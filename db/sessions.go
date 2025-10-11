package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"time"

	"errors"

	"github.com/mbertschler/foundation"
	"github.com/uptrace/bun"
)

var (
	SessionLength           = 32
	CSRFTokenLength         = 32
	SessionDuration         = 90 * 24 * time.Hour
	SessionRotationInterval = 30 * time.Minute

	_ foundation.SessionDB = (*sessionsDB)(nil)
)

var nilSession *foundation.Session

type sessionsDB struct {
	db *bun.DB
}

// InsertUserSession creates a new session for a user
func (s *sessionsDB) InsertUserSession(ctx context.Context, userID int64) (*foundation.Session, error) {
	return s.insertSession(ctx, sql.NullInt64{Int64: userID, Valid: true})
}

// InsertAnonymousSession creates a new anonymous session (no user ID)
func (s *sessionsDB) InsertAnonymousSession(ctx context.Context) (*foundation.Session, error) {
	return s.insertSession(ctx, sql.NullInt64{Valid: false})
}

func (s *sessionsDB) insertSession(ctx context.Context, userID sql.NullInt64) (*foundation.Session, error) {
	sessionID, err := generateRandomID(SessionLength)
	if err != nil {
		return nil, err
	}

	csrfToken, err := generateRandomID(CSRFTokenLength)
	if err != nil {
		return nil, err
	}

	session := &foundation.Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(SessionDuration),
		CSRFToken: csrfToken,
	}

	_, err = s.db.NewInsert().Model(session).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *sessionsDB) ByID(ctx context.Context, sessionID string) (*foundation.Session, error) {
	var session foundation.Session
	err := s.db.NewSelect().Model(&session).Where("id = ?", sessionID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *sessionsDB) Delete(ctx context.Context, sessionID string) error {
	_, err := s.db.NewDelete().Model(nilSession).Where("id = ?", sessionID).Exec(ctx)
	return err
}

func (s *sessionsDB) RotateSessionIfNeeded(ctx context.Context, sessionID string) (*foundation.Session, error) {
	// Get the current session
	currentSession, err := s.ByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Check if session is expired
	if time.Now().After(currentSession.ExpiresAt) {
		return nil, sql.ErrNoRows // or a custom error
	}

	// Only rotate user sessions, not anonymous
	if !currentSession.UserID.Valid {
		return currentSession, nil
	}

	// Check if rotation is needed (more than 30 minutes since creation)
	if time.Since(currentSession.CreatedAt) <= SessionRotationInterval {
		return currentSession, nil
	}

	// Generate new session ID and CSRF token
	newSessionID, err := generateRandomID(SessionLength)
	if err != nil {
		return nil, err
	}

	newCSRFToken, err := generateRandomID(CSRFTokenLength)
	if err != nil {
		return nil, err
	}

	// Create new session
	newSession := &foundation.Session{
		ID:        newSessionID,
		UserID:    currentSession.UserID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(SessionDuration),
		CSRFToken: newCSRFToken,
	}

	// Insert new session
	_, err = s.db.NewInsert().Model(newSession).Exec(ctx)
	if err != nil {
		return nil, err
	}

	// Delete old session
	err = s.Delete(ctx, sessionID)
	if err != nil {
		// If delete fails, we should probably delete the new session to avoid duplicates
		newErr := s.Delete(ctx, newSessionID)
		if newErr != nil {
			err = errors.Join(err, newErr)
		}
		return nil, err
	}

	return newSession, nil
}

func (s *sessionsDB) startCleanup() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			// Delete expired sessions from database
			err := s.deleteExpired(context.Background())
			if err != nil {
				// Log error but continue
				continue
			}
		}
	}()
}

func (s *sessionsDB) deleteExpired(ctx context.Context) error {
	_, err := s.db.NewDelete().Model(nilSession).Where("expires_at < ?", time.Now()).Exec(ctx)
	return err
}

// generateRandomID generates a random ID of the specified length
func generateRandomID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
