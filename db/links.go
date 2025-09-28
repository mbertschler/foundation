package db

import (
	"context"

	"github.com/mbertschler/foundation"
	"github.com/uptrace/bun"
)

var (
	nilLink *foundation.Link

	_ foundation.LinkDB = (*linksDB)(nil)
)

type linksDB struct {
	db *bun.DB
}

func (l *linksDB) Insert(ctx context.Context, link *foundation.Link) error {
	_, err := l.db.NewInsert().Model(link).Exec(ctx)
	return err
}

func (l *linksDB) Update(ctx context.Context, link *foundation.Link) error {
	_, err := l.db.NewUpdate().Model(link).WherePK().Exec(ctx)
	return err
}

func (l *linksDB) ByShortLink(ctx context.Context, shortLink string) (*foundation.Link, error) {
	var link foundation.Link
	err := l.db.NewSelect().Model(&link).Where("short_link = ?", shortLink).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (l *linksDB) All(ctx context.Context) ([]*foundation.Link, error) {
	var links []*foundation.Link
	err := l.db.NewSelect().Model(&links).Relation("User").Order("short_link ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return links, nil
}

func (l *linksDB) AllWithVisitCounts(ctx context.Context) ([]*foundation.Link, error) {
	var links []*foundation.Link
	err := l.db.NewSelect().Model(&links).Relation("User").ColumnExpr("l.*").
		ColumnExpr("(SELECT COUNT(*) FROM link_visits WHERE short_link = l.short_link) AS visits_count").
		Order("l.short_link").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return links, nil
}

func (l *linksDB) Delete(ctx context.Context, shortLink string) error {
	_, err := l.db.NewDelete().Model(nilLink).Where("short_link = ?", shortLink).Exec(ctx)
	return err
}
