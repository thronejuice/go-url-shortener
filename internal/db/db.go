// Package db opens the PostgreSQL connection pool and applies migrations.
package db

import (
	"context"
	"errors"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // registers the pgx5:// database driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/thronejuice/go-url-shortener/migrations"
)

// Migrate applies all pending embedded migrations; it is a no-op when up to date.
func Migrate(databaseURL string) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, migrateURL(databaseURL))
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// Connect opens a pgx pool and verifies the database is reachable.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// migrateURL maps postgres:// and postgresql:// to the pgx5:// scheme golang-migrate expects.
func migrateURL(u string) string {
	for _, p := range []string{"postgresql://", "postgres://"} {
		if rest, ok := strings.CutPrefix(u, p); ok {
			return "pgx5://" + rest
		}
	}
	return u
}
