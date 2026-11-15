// Package link holds the URL-shortening domain: model, storage interface and service.
package link

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound   = errors.New("link not found")
	ErrConflict   = errors.New("code already exists")
	ErrInvalidURL = errors.New("invalid url")
)

type Link struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

// Store is the persistence boundary. Day 2 uses MemoryStore; week 2 swaps in PostgreSQL.
type Store interface {
	Create(ctx context.Context, l Link) error
	Get(ctx context.Context, code string) (Link, error)
}
