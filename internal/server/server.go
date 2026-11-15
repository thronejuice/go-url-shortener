// Package server wires HTTP routes and middleware.
package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/thronejuice/go-url-shortener/internal/link"
)

type Server struct {
	logger *slog.Logger
	links  *link.Service
}

func New(logger *slog.Logger, links *link.Service) http.Handler {
	s := &Server{logger: logger, links: links}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("POST /api/links", s.createLink)
	mux.HandleFunc("GET /{code}", s.redirect)

	return logging(logger, mux)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type createLinkRequest struct {
	URL string `json:"url"`
}

func (s *Server) createLink(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	l, err := s.links.Shorten(r.Context(), req.URL)
	switch {
	case errors.Is(err, link.ErrInvalidURL):
		writeError(w, http.StatusUnprocessableEntity, "url must be an absolute http(s) URL")
	case err != nil:
		s.logger.Error("shorten failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusCreated, l)
	}
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	l, err := s.links.Resolve(r.Context(), r.PathValue("code"))
	switch {
	case errors.Is(err, link.ErrNotFound):
		writeError(w, http.StatusNotFound, "link not found")
	case err != nil:
		s.logger.Error("resolve failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		http.Redirect(w, r, l.URL, http.StatusFound)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}
