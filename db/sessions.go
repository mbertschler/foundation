package db

import (
	"context"
	"time"

	"github.com/mbertschler/foundation"
	"github.com/uptrace/bun"
)

var nilSession *foundation.Session

type sessionsDB struct {
	db *bun.DB
}

func (s *sessionsDB) Create(ctx context.Context, session *foundation.Session) error {
	_, err := s.db.NewInsert().Model(session).Exec(ctx)
	return err
}

func (s *sessionsDB) Get(ctx context.Context, sessionID string) (*foundation.Session, error) {
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

func (s *sessionsDB) DeleteExpired(ctx context.Context) error {
	_, err := s.db.NewDelete().Model(nilSession).Where("expires_at < ?", time.Now()).Exec(ctx)
	return err
}
