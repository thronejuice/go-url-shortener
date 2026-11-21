package link

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type CachedStore struct {
	inner  Store
	rdb    *redis.Client
	ttl    time.Duration
	logger *slog.Logger
}

var _ Store = (*CachedStore)(nil)

func NewCachedStore(inner Store, rdb *redis.Client, ttl time.Duration, logger *slog.Logger) *CachedStore {
	return &CachedStore{inner: inner, rdb: rdb, ttl: ttl, logger: logger}
}

func (c *CachedStore) Get(ctx context.Context, code string) (Link, error) {
	raw, err := c.rdb.Get(ctx, cacheKey(code)).Bytes()
	switch {
	case err == nil:
		var l Link
		if jerr := json.Unmarshal(raw, &l); jerr == nil {
			return l, nil
		}
		c.logger.Warn("corrupt cache entry, falling back to store", "code", code)
	case !errors.Is(err, redis.Nil):
		c.logger.Warn("cache read failed, falling back tostore", "err", err)
	}
	l, err := c.inner.Get(ctx, code)
	if err != nil {
		return Link{}, err
	}
	c.set(ctx, l)
	return l, nil
}

func (c *CachedStore) Create(ctx context.Context, l Link) error {
	if err := c.inner.Create(ctx, l); err != nil {
		return err
	}
	c.set(ctx, l)
	return nil
}

func (c *CachedStore) set(ctx context.Context, l Link) {
	raw, err := json.Marshal(l)
	if err != nil {
		c.logger.Warn("cache encode failed", "err", err)
		return
	}
	if err := c.rdb.Set(ctx, cacheKey(l.Code), raw, c.ttl).Err(); err != nil {
		c.logger.Warn("cache write failed", "err", err)
	}
}

func cacheKey(code string) string {
	return "link:" + code
}
