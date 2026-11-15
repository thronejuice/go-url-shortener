package link

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolation = "23505" // รหัส SQLSTATE ของ unique constraint

type PostgresStore struct {
	pool *pgxpool.Pool
}

var _ Store = (*PostgresStore)(nil)

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(ctx context.Context, l Link) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO links (code, url, created_at) VALUES ($1, $2, $3)`,
		l.Code, l.URL, l.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return ErrConflict // Service.Shorten จะลองสุ่ม code ใหม่ให้เอง
	}
	return err
}

func (s *PostgresStore) Get(ctx context.Context, code string) (Link, error) {
	var l Link
	err := s.pool.QueryRow(ctx,
		`SELECT code, url, created_at FROM links WHERE code = $1`, code,
	).Scan(&l.Code, &l.URL, &l.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Link{}, ErrNotFound
	}
	return l, err
}
