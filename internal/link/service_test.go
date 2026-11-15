package link

import (
	"context"
	"errors"
	"testing"
)

func TestShortenAndResolve(t *testing.T) {
	svc := NewService(NewMemoryStore())
	ctx := context.Background()

	l, err := svc.Shorten(ctx, "https://example.com/a?b=c")
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Code) != codeLength {
		t.Fatalf("code length = %d, want %d", len(l.Code), codeLength)
	}
	got, err := svc.Resolve(ctx, l.Code)
	if err != nil || got.URL != "https://example.com/a?b=c" {
		t.Fatalf("resolve = %+v, %v", got, err)
	}
}

func TestShortenRejectsInvalidURL(t *testing.T) {
	svc := NewService(NewMemoryStore())
	for _, raw := range []string{"", "not a url", "ftp://example.com", "javascript:alert(1)", "http://"} {
		if _, err := svc.Shorten(context.Background(), raw); !errors.Is(err, ErrInvalidURL) {
			t.Errorf("Shorten(%q) err = %v, want ErrInvalidURL", raw, err)
		}
	}
}

func TestResolveNotFound(t *testing.T) {
	svc := NewService(NewMemoryStore())
	if _, err := svc.Resolve(context.Background(), "nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
