package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thronejuice/go-url-shortener/internal/link"
)

func newTestHandler() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(logger, link.NewService(link.NewMemoryStore()))
}

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

func TestHealthz(t *testing.T) {
	rec := do(newTestHandler(), http.MethodGet, "/healthz", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestCreateAndRedirect(t *testing.T) {
	h := newTestHandler()

	rec := do(h, http.MethodPost, "/api/links", `{"url":"https://example.com/x"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body=%s", rec.Code, rec.Body)
	}
	var created link.Link
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.Code == "" {
		t.Fatalf("bad create response: %s", rec.Body)
	}

	rec = do(h, http.MethodGet, "/"+created.Code, "")
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "https://example.com/x" {
		t.Fatalf("redirect = %d Location=%q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestCreateValidation(t *testing.T) {
	h := newTestHandler()
	tests := []struct {
		name, body string
		want       int
	}{
		{"bad json", `{`, http.StatusBadRequest},
		{"missing url", `{}`, http.StatusUnprocessableEntity},
		{"bad scheme", `{"url":"ftp://x.com"}`, http.StatusUnprocessableEntity},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := do(h, http.MethodPost, "/api/links", tc.body).Code; got != tc.want {
				t.Fatalf("status = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestRedirectNotFound(t *testing.T) {
	if got := do(newTestHandler(), http.MethodGet, "/missing", "").Code; got != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", got)
	}
}
