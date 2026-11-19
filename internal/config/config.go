// Package config loads application settings from environment variables (12-factor).
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"time"
)

type Config struct {
	Env             string
	Addr            string
	DatabaseURL     string
	RedisURL        string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Env:         getenv("APP_ENV", "dev"),
		Addr:        getenv("ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
	}

	var errs []error

	d, err := getDuration("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		errs = append(errs, err)
	}
	cfg.ShutdownTimeout = d
	if cfg.Env != "dev" && cfg.Env != "prod" {
		errs = append(errs, fmt.Errorf("APP_ENV must be dev or prod, got %q", cfg.Env))
	}
	if _, _, err := net.SplitHostPort(cfg.Addr); err != nil {
		errs = append(errs, fmt.Errorf("ADDR %q must be host:port: %w", cfg.Addr, err))
	}
	if cfg.Env == "prod" && cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required when APP_ENV=prod"))
	}
	if cfg.DatabaseURL != "" {
		if u, err := url.Parse(cfg.DatabaseURL); err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
			errs = append(errs, errors.New("DATABASE_URL must be a postgres:// or postgresql:// URL"))
		}
	}
	return cfg, errors.Join(errs...)
}

// LogValue makes slog print the config with secrets redacted.
func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("env", c.Env),
		slog.String("addr", c.Addr),
		slog.String("database_url", redact(c.DatabaseURL)),
		slog.String("redis_url", redact(c.RedisURL)),
		slog.String("shutdown_timeout", c.ShutdownTimeout.String()),
	)
}

func redact(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "<invalid>"
	}
	return u.Redacted()
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback, fmt.Errorf("%s: %w", key, err)
	}
	return d, nil
}
