package link

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"
	"time"
)

const (
	alphabet   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	codeLength = 7
	maxRetries = 5
)

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: time.Now}
}

// Shorten validates rawURL and stores it under a fresh random code.
func (s *Service) Shorten(ctx context.Context, rawURL string) (Link, error) {
	if err := validateURL(rawURL); err != nil {
		return Link{}, err
	}
	for range maxRetries {
		code, err := newCode()
		if err != nil {
			return Link{}, err
		}
		l := Link{Code: code, URL: rawURL, CreatedAt: s.now().UTC()}
		err = s.store.Create(ctx, l)
		if errors.Is(err, ErrConflict) {
			continue // astronomically rare collision: try another code
		}
		if err != nil {
			return Link{}, err
		}
		return l, nil
	}
	return Link{}, ErrConflict
}

func (s *Service) Resolve(ctx context.Context, code string) (Link, error) {
	return s.store.Get(ctx, code)
}

func validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return ErrInvalidURL
	}
	return nil
}

func newCode() (string, error) {
	b := make([]byte, codeLength)
	limit := big.NewInt(int64(len(alphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", err
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b), nil
}
