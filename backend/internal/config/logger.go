package config

import (
	"log/slog"
	"os"
)

// SetupLogger configures the global slog logger.
// In production: JSON output. In development: human-readable text output.
func SetupLogger(cfg *Config) *slog.Logger {
	var handler slog.Handler

	if cfg.IsDevelopment() {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
