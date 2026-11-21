package link_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/thronejuice/go-url-shortener/internal/link"
)

type countingStore struct {
	link.Store
	gets int
}

func (c *countingStore) Get(ctx context.Context, code string) (link.Link, error) {
	c.gets++
	return c.Store.Get(ctx, code)
}

func newCached(t *testing.T) (*link.CachedStore, *countingStore, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	inner := &countingStore{Store: link.NewMemoryStore()}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return link.NewCachedStore(inner, rdb, time.Minute, logger), inner, mr
}

func seed(t *testing.T, inner link.Store) link.Link {
	t.Helper()
	l := link.Link{Code: "abc1234", URL: "https://go.dev", CreatedAt: time.Now().UTC()}
	if err := inner.Create(context.Background(), l); err != nil {
		t.Fatal(err)
	}
	return l
}

func TestGetMissThenHit(t *testing.T) {
	cached, inner, _ := newCached(t)
	l := seed(t, inner)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		got, err := cached.Get(ctx, l.Code)
		if err != nil || got.URL != l.URL {
			t.Fatalf("get #%d = %+v, %v", i, got, err)
		}
	}

	if inner.gets != 1 {
		t.Fatalf("inner.Get called %d time, want 1 (thd rest should hit the cache)", inner.gets)
	}
}

func TestGetNotFoundPassedThrough(t *testing.T) {
	cached, _, _ := newCached(t)
	if _, err := cached.Get(context.Background(),
		"nope"); !errors.Is(err, link.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestTTLExpirygoesBacktoStore(t *testing.T) {
	cached, inner, mr := newCached(t)
	l := seed(t, inner)
	ctx := context.Background()

	_, _ = cached.Get(ctx, l.Code)
	mr.FastForward(2 * time.Minute)
	_, _ = cached.Get(ctx, l.Code)

	if inner.gets != 2 {
		t.Fatalf("inner.Get called %d times, want 2 after TTL exoiry", inner.gets)
	}
}

func TestRedisDownFallsBackToStore(t *testing.T) {
	cached, inner, mr := newCached(t)
	l := seed(t, inner)

	mr.Close()
	got, err := cached.Get(context.Background(), l.Code)
	if err != nil || got.URL != l.URL {
		t.Fatalf("get with redis down = %+v, %v; want data from the store", got, err)
	}
}

func TestCreateWarmsCache(t *testing.T) {
	cached, inner, _ := newCached(t)
	l := link.Link{Code: "new1234", URL: "https://example.com", CreatedAt: time.Now().UTC()}
	ctx := context.Background()

	if err := cached.Create(ctx, l); err != nil {
		t.Fatal(err)
	}
	if _, err := cached.Get(ctx, l.Code); err != nil {
		t.Fatal(err)
	}
	if inner.gets != 0 {
		t.Fatalf("inner.Get caled %d times, wat 0 (Create should warm the cache)", inner.gets)
	}
}
