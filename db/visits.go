package db

import (
	"context"

	"github.com/mbertschler/foundation"
	"github.com/uptrace/bun"
)

var (
	nilVisit *foundation.LinkVisit

	_ foundation.VisitDB = (*visitsDB)(nil)
)

type visitsDB struct {
	db *bun.DB
}

func (v *visitsDB) Insert(ctx context.Context, visit *foundation.LinkVisit) error {
	_, err := v.db.NewInsert().Model(visit).Exec(ctx)
	return err
}

func (v *visitsDB) CountByLink(ctx context.Context, shortLink string) (int64, error) {
	count, err := v.db.NewSelect().Table("link_visits").Where("short_link = ?", shortLink).Count(ctx)
	return int64(count), err
}
