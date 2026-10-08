package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port        int
	Environment string // "development" or "production"

	DatabaseURL string

	AppUsername     string
	AppPasswordHash string
	SessionDuration time.Duration

	CORSOrigins []string

	FetchTimeout      time.Duration
	FetchMaxSize      int64
	FetchMaxRedirects int
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{}

	cfg.Port = getEnvInt("PORT", 8080)
	cfg.Environment = getEnvStr("ENVIRONMENT", "development")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	cfg.DatabaseURL = dbURL

	cfg.AppUsername = os.Getenv("APP_USERNAME")
	if cfg.AppUsername == "" {
		return nil, fmt.Errorf("APP_USERNAME is required")
	}

	cfg.AppPasswordHash = os.Getenv("APP_PASSWORD_HASH")
	if cfg.AppPasswordHash == "" {
		return nil, fmt.Errorf("APP_PASSWORD_HASH is required")
	}

	var err error
	cfg.SessionDuration, err = getEnvDuration("SESSION_DURATION", 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("SESSION_DURATION: %w", err)
	}

	cfg.CORSOrigins = getEnvStrSlice("CORS_ORIGINS", []string{})

	cfg.FetchTimeout, err = getEnvDuration("FETCH_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("FETCH_TIMEOUT: %w", err)
	}

	cfg.FetchMaxSize = getEnvInt64("FETCH_MAX_SIZE", 5*1024*1024) // 5 MB
	cfg.FetchMaxRedirects = getEnvInt("FETCH_MAX_REDIRECTS", 3)

	return cfg, nil
}

// IsDevelopment returns true when running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func getEnvStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvInt64(key string, fallback int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", v, err)
	}
	return d, nil
}

func getEnvStrSlice(key string, fallback []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
