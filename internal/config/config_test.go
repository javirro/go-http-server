package config_test

import (
	"testing"
	"time"

	"github.com/javier/go-http-server/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("PORT", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("Host = %q, want 0.0.0.0", cfg.Host)
	}
	if cfg.ReadTimeout != 5*time.Second {
		t.Errorf("ReadTimeout = %v, want 5s", cfg.ReadTimeout)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("LogFormat = %q, want json", cfg.LogFormat)
	}
}

func TestLoad_CustomEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "text")
	t.Setenv("ENV", "production")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want 9090", cfg.Port)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug", cfg.LogLevel)
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %q, want production", cfg.Env)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("PORT", "99999")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid port, got nil")
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "verbose")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid log level, got nil")
	}
}

func TestLoad_TLSMismatch(t *testing.T) {
	t.Setenv("TLS_CERT_FILE", "cert.pem")
	t.Setenv("TLS_KEY_FILE", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() expected error when only cert is set, got nil")
	}
}

func TestConfig_Addr(t *testing.T) {
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("PORT", "3000")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if got := cfg.Addr(); got != "127.0.0.1:3000" {
		t.Errorf("Addr() = %q, want 127.0.0.1:3000", got)
	}
}
