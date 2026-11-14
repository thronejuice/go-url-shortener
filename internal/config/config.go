// Package config loads application settings from environment variables (12-factor).
package config

import "os"

type Config struct {
	Addr        string
	DatabaseURL string
	RedisURL    string
}

func Load() Config {
	return Config{
		Addr:        getenv("ADDR", ":8080"),
		DatabaseURL: getenv("DATABASE_URL", ""),
		RedisURL:    getenv("REDIS_URL", ""),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
