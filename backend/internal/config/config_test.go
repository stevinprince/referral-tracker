package config

import (
	"os"
	"testing"
	"time"
)

// setEnvs sets environment variables for a test and returns a cleanup function.
func setEnvs(t *testing.T, envs map[string]string) {
	t.Helper()
	for k, v := range envs {
		t.Setenv(k, v)
	}
}

// requiredEnvs returns the minimum required env vars for a valid config.
func requiredEnvs() map[string]string {
	return map[string]string{
		"DATABASE_URL":      "postgres://user:pass@localhost:5432/test?sslmode=disable",
		"APP_USERNAME":      "admin",
		"APP_PASSWORD_HASH": "$2a$12$fakehashvalue",
	}
}

func TestLoad_Defaults(t *testing.T) {
	setEnvs(t, requiredEnvs())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"Port", cfg.Port, 8080},
		{"Environment", cfg.Environment, "development"},
		{"SessionDuration", cfg.SessionDuration, 24 * time.Hour},
		{"FetchTimeout", cfg.FetchTimeout, 10 * time.Second},
		{"FetchMaxSize", cfg.FetchMaxSize, int64(5 * 1024 * 1024)},
		{"FetchMaxRedirects", cfg.FetchMaxRedirects, 3},
		{"CORSOrigins length", len(cfg.CORSOrigins), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestLoad_CustomValues(t *testing.T) {
	envs := requiredEnvs()
	envs["PORT"] = "3000"
	envs["ENVIRONMENT"] = "production"
	envs["SESSION_DURATION"] = "12h"
	envs["CORS_ORIGINS"] = "http://localhost:5173, https://example.com"
	envs["FETCH_TIMEOUT"] = "5s"
	envs["FETCH_MAX_SIZE"] = "1048576"
	envs["FETCH_MAX_REDIRECTS"] = "5"
	setEnvs(t, envs)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"Port", cfg.Port, 3000},
		{"Environment", cfg.Environment, "production"},
		{"IsDevelopment", cfg.IsDevelopment(), false},
		{"SessionDuration", cfg.SessionDuration, 12 * time.Hour},
		{"FetchTimeout", cfg.FetchTimeout, 5 * time.Second},
		{"FetchMaxSize", cfg.FetchMaxSize, int64(1048576)},
		{"FetchMaxRedirects", cfg.FetchMaxRedirects, 5},
		{"CORSOrigins count", len(cfg.CORSOrigins), 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}

	if cfg.CORSOrigins[0] != "http://localhost:5173" {
		t.Errorf("CORSOrigins[0] = %q, want %q", cfg.CORSOrigins[0], "http://localhost:5173")
	}
	if cfg.CORSOrigins[1] != "https://example.com" {
		t.Errorf("CORSOrigins[1] = %q, want %q", cfg.CORSOrigins[1], "https://example.com")
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	tests := []struct {
		name    string
		envs    map[string]string
		wantErr string
	}{
		{
			name:    "missing DATABASE_URL",
			envs:    map[string]string{"APP_USERNAME": "admin", "APP_PASSWORD_HASH": "$2a$12$hash"},
			wantErr: "DATABASE_URL is required",
		},
		{
			name:    "missing APP_USERNAME",
			envs:    map[string]string{"DATABASE_URL": "postgres://localhost/test", "APP_PASSWORD_HASH": "$2a$12$hash"},
			wantErr: "APP_USERNAME is required",
		},
		{
			name:    "missing APP_PASSWORD_HASH",
			envs:    map[string]string{"DATABASE_URL": "postgres://localhost/test", "APP_USERNAME": "admin"},
			wantErr: "APP_PASSWORD_HASH is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars that might be set
			for _, key := range []string{"DATABASE_URL", "APP_USERNAME", "APP_PASSWORD_HASH"} {
				t.Setenv(key, "")
			}
			// Then set the ones for this test case
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}

			_, err := Load()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestLoad_InvalidDuration(t *testing.T) {
	envs := requiredEnvs()
	envs["SESSION_DURATION"] = "not-a-duration"
	setEnvs(t, envs)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid SESSION_DURATION, got nil")
	}
}

func TestLoad_InvalidFetchTimeout(t *testing.T) {
	envs := requiredEnvs()
	envs["FETCH_TIMEOUT"] = "bad"
	setEnvs(t, envs)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid FETCH_TIMEOUT, got nil")
	}
}

func TestLoad_InvalidPortFallsBackToDefault(t *testing.T) {
	envs := requiredEnvs()
	envs["PORT"] = "not-a-number"
	setEnvs(t, envs)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080 (default)", cfg.Port)
	}
}

func TestLoad_CORSOriginsEmpty(t *testing.T) {
	envs := requiredEnvs()
	setEnvs(t, envs)

	// Explicitly ensure CORS_ORIGINS is not set
	os.Unsetenv("CORS_ORIGINS")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.CORSOrigins) != 0 {
		t.Errorf("CORSOrigins = %v, want empty", cfg.CORSOrigins)
	}
}

func TestIsDevelopment(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"development", true},
		{"production", false},
		{"staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := &Config{Environment: tt.env}
			if got := cfg.IsDevelopment(); got != tt.want {
				t.Errorf("IsDevelopment() = %v, want %v", got, tt.want)
			}
		})
	}
}
