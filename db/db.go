package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mbertschler/foundation"
	"github.com/mbertschler/foundation/db/migrations"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
)

type DB struct {
	Users    *usersDB
	Sessions *sessionsDB
	Links    *linksDB
	Visits   *visitsDB

	sqlDB *sql.DB
}

func (db *DB) SetSQLDB(sqlDB *sql.DB) {
	db.sqlDB = sqlDB
}

func (db *DB) Close() error {
	return db.sqlDB.Close()
}

func StartDB(context *foundation.Context) (*DB, error) {
	ctx := context.Context

	path := context.Config.DBPath

	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, errors.Wrapf(err, "MkdirAll %q", dir)
	}
	connString := fmt.Sprintf("file:%s?_journal_mode=WAL&cache=shared", context.Config.DBPath)
	sqldb, err := sql.Open(sqliteshim.ShimName, connString)
	if err != nil {
		return nil, errors.Wrapf(err, "sql.Open with %q", connString)
	}

	// TODO: close somewhere else on shutdown
	// defer sqldb.Close()

	// Create Bun database instance
	db := bun.NewDB(sqldb, sqlitedialect.New())

	// Add query debugging (optional)
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
	))

	// Initialize and run migrations
	migrator := migrations.NewMigrator(db)

	// Initialize migration table if it doesn't exist
	err = migrator.Init(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "init migrations")
	}

	// Run migrations
	err = migrator.Migrate(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "run migrations")
	}
	sessionDB := &sessionsDB{db: db}
	sessionDB.startCleanup()

	fdb := &DB{
		Users:    &usersDB{db: db},
		Sessions: sessionDB,
		Links:    &linksDB{db: db},
		Visits:   &visitsDB{db: db},
	}

	fdb.SetSQLDB(sqldb)

	return fdb, nil
}
