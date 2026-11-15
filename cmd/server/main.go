package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thronejuice/go-url-shortener/internal/config"
	"github.com/thronejuice/go-url-shortener/internal/db"
	"github.com/thronejuice/go-url-shortener/internal/link"
	"github.com/thronejuice/go-url-shortener/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var store link.Store = link.NewMemoryStore()
	if cfg.DatabaseURL != "" {
		if err := db.Migrate(cfg.DatabaseURL); err != nil {
			logger.Error("migrate failed", "err", err)
			os.Exit(1)
		}
		pool, err := db.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			logger.Error("db connect failed", "err", err)
			os.Exit(1)
		}
		defer pool.Close()
		store = link.NewPostgresStore(pool)
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           server.New(logger, link.NewService(store)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server starting", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
}
