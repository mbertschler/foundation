package migrations

import (
	"context"
	"log"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

var Migrations = migrate.NewMigrations()

func init() {
	if err := Migrations.DiscoverCaller(); err != nil {
		panic(err)
	}
}

type Migrator struct {
	migrator *migrate.Migrator
}

func NewMigrator(db *bun.DB) *Migrator {
	return &Migrator{
		migrator: migrate.NewMigrator(db, Migrations),
	}
}

func (m *Migrator) Migrate(ctx context.Context) error {
	group, err := m.migrator.Migrate(ctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		log.Printf("no new migrations to run\n")
		return nil
	}
	log.Printf("migrated to %s\n", group)
	return nil
}

func (m *Migrator) Rollback(ctx context.Context) error {
	group, err := m.migrator.Rollback(ctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		log.Printf("no migrations to rollback\n")
		return nil
	}
	log.Printf("rolled back %s\n", group)
	return nil
}

func (m *Migrator) Status(ctx context.Context) (migrate.MigrationSlice, error) {
	return m.migrator.MigrationsWithStatus(ctx)
}

func (m *Migrator) Init(ctx context.Context) error {
	return m.migrator.Init(ctx)
}
