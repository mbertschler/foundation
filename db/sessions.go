package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/mbertschler/foundation"
	"github.com/uptrace/bun"
)

var (
	SessionLength   = 32
	SessionDuration = 24 * time.Hour

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
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	session := &foundation.Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(SessionDuration),
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

// generateSessionID generates a random session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, SessionLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
