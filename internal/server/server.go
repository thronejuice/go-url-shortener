// Package server wires HTTP routes and middleware.
package server

import (
	"log/slog"
	"net/http"
	"time"
)

func New(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)

	// TODO(week 2): POST /api/links, GET /{code}

	return logging(logger, mux)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}
