package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration

	// TLS (optional)
	TLSCertFile string
	TLSKeyFile  string

	// Logging
	LogLevel  string
	LogFormat string // "json" | "text"

	// CORS
	CORSAllowedOrigins []string
	CORSAllowedMethods []string
	CORSAllowedHeaders []string

	// Rate limiting (requests per second per IP)
	RateLimitRPS   float64
	RateLimitBurst int

	// Environment
	Env string // "development" | "staging" | "production"
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Host:            getEnv("HOST", "0.0.0.0"),
		Port:            getEnvInt("PORT", 8080),
		ReadTimeout:     getEnvDuration("READ_TIMEOUT", 5*time.Second),
		WriteTimeout:    getEnvDuration("WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:     getEnvDuration("IDLE_TIMEOUT", 120*time.Second),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
		TLSCertFile:     getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:      getEnv("TLS_KEY_FILE", ""),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		LogFormat:       getEnv("LOG_FORMAT", "json"),
		CORSAllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
		CORSAllowedMethods: getEnvSlice("CORS_ALLOWED_METHODS", []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		}),
		CORSAllowedHeaders: getEnvSlice("CORS_ALLOWED_HEADERS", []string{
			"Accept", "Authorization", "Content-Type", "X-Request-ID",
		}),
		RateLimitRPS:   getEnvFloat("RATE_LIMIT_RPS", 100),
		RateLimitBurst: getEnvInt("RATE_LIMIT_BURST", 200),
		Env:            getEnv("ENV", "development"),
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	var errs []error

	if c.Port < 1 || c.Port > 65535 {
		errs = append(errs, fmt.Errorf("PORT must be between 1 and 65535, got %d", c.Port))
	}

	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[strings.ToLower(c.LogLevel)] {
		errs = append(errs, fmt.Errorf("LOG_LEVEL must be one of debug|info|warn|error, got %q", c.LogLevel))
	}

	if c.LogFormat != "json" && c.LogFormat != "text" {
		errs = append(errs, fmt.Errorf("LOG_FORMAT must be json or text, got %q", c.LogFormat))
	}

	tlsCert := c.TLSCertFile != ""
	tlsKey := c.TLSKeyFile != ""
	if tlsCert != tlsKey {
		errs = append(errs, errors.New("TLS_CERT_FILE and TLS_KEY_FILE must both be set or both be empty"))
	}

	if c.RateLimitRPS <= 0 {
		errs = append(errs, fmt.Errorf("RATE_LIMIT_RPS must be positive, got %f", c.RateLimitRPS))
	}

	if c.RateLimitBurst < 1 {
		errs = append(errs, fmt.Errorf("RATE_LIMIT_BURST must be at least 1, got %d", c.RateLimitBurst))
	}

	return errors.Join(errs...)
}

// Addr returns the host:port address string.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// TLSEnabled reports whether TLS is configured.
func (c *Config) TLSEnabled() bool {
	return c.TLSCertFile != "" && c.TLSKeyFile != ""
}

// IsDevelopment reports whether the environment is development.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return n
}

func getEnvFloat(key string, defaultValue float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return defaultValue
	}
	return f
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return defaultValue
	}
	return d
}

func getEnvSlice(key string, defaultValue []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}
