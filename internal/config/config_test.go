package config

import (
	"strings"
	"testing"
	"time"
)

// t.Setenv ใช้ร่วมกับ t.Parallel ไม่ได้ จึงไม่ใช้ Parallel ที่นี่
func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"ADDR", "DATABASE_URL", "REDIS_URL", "APP_ENV", "SHUTDOWN_TIMEOUT"} {
		t.Setenv(k, "")
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Addr != ":8080" || cfg.Env != "dev" || cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadProdRequiresDatabaseURL(t *testing.T) {
	clearEnv(t)
	t.Setenv("APP_ENV", "prod")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("err = %v, want mention of DATABASE_URL", err)
	}
}

func TestLoadReportsAllProblemsAtOnce(t *testing.T) {
	clearEnv(t)
	t.Setenv("ADDR", "not-an-address")
	t.Setenv("SHUTDOWN_TIMEOUT", "soon")
	t.Setenv("DATABASE_URL", "mysql://x")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"ADDR", "SHUTDOWN_TIMEOUT", "DATABASE_URL"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %s", err, want)
		}
	}
}
