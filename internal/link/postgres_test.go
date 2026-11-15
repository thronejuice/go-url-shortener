package link_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/thronejuice/go-url-shortener/internal/db"
	"github.com/thronejuice/go-url-shortener/internal/link"
)

// newPostgresStore returns a store backed by the database in DATABASE_URL,
// or skips the test when it is not set (e.g. plain `go test` without Postgres).
func newPostgresStore(t *testing.T) (*link.PostgresStore, *pgxpool.Pool) {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set; skipping PostgreSQL integration test")
	}
	if err := db.Migrate(url); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := db.Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return link.NewPostgresStore(pool), pool
}

// cleanupCode removes the test row when the test ends.
func cleanupCode(t *testing.T, pool *pgxpool.Pool, code string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM links WHERE code = $1`, code)
	})
}

func TestPostgresStoreCreateAndGet(t *testing.T) {
	store, pool := newPostgresStore(t)
	ctx := context.Background()

	want := link.Link{
		Code:      "it-" + time.Now().Format("150405.000000"),
		URL:       "https://example.com/integration",
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond), // Postgres stores microseconds
	}
	cleanupCode(t, pool, want.Code)

	if err := store.Create(ctx, want); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := store.Get(ctx, want.Code)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Code != want.Code || got.URL != want.URL || !got.CreatedAt.Equal(want.CreatedAt) {
		t.Fatalf("get = %+v, want %+v", got, want)
	}
}

func TestPostgresStoreDuplicateCodeIsConflict(t *testing.T) {
	store, pool := newPostgresStore(t)
	ctx := context.Background()

	l := link.Link{Code: "dup-" + time.Now().Format("150405.000000"), URL: "https://example.com/dup", CreatedAt: time.Now().UTC()}
	cleanupCode(t, pool, l.Code)

	if err := store.Create(ctx, l); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if err := store.Create(ctx, l); !errors.Is(err, link.ErrConflict) {
		t.Fatalf("second create err = %v, want ErrConflict", err)
	}
}

func TestPostgresStoreGetMissingIsNotFound(t *testing.T) {
	store, _ := newPostgresStore(t)

	if _, err := store.Get(context.Background(), "does-not-exist"); !errors.Is(err, link.ErrNotFound) {
		t.Fatalf("get err = %v, want ErrNotFound", err)
	}
}
