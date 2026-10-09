package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/stevin/referral-tracker/backend/internal/config"
	"github.com/stevin/referral-tracker/backend/internal/handler"
	"github.com/stevin/referral-tracker/backend/internal/middleware"
	"github.com/stevin/referral-tracker/backend/internal/repository"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Set up structured logging
	logger := config.SetupLogger(cfg)
	logger.Info("starting referral-tracker server",
		"port", cfg.Port,
		"environment", cfg.Environment,
	)

	// Connect to database
	pool, err := repository.ConnectDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()
	logger.Info("connected to database")

	// Run migrations
	migrationsDir, err := findMigrationsDir()
	if err != nil {
		return err
	}
	if err := repository.RunMigrations(ctx, pool, migrationsDir); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	logger.Info("migrations complete")

	// Set up router
	reqLogger := middleware.NewRequestLogger(logger)
	r := handler.NewRouter(reqLogger, cfg.Environment)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", srv.Addr)
		serverErr <- srv.ListenAndServe()
	}()

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-serverErr:
		if err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	}

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	logger.Info("shutting down server")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	logger.Info("server stopped")
	return nil
}

// findMigrationsDir returns the path to the migrations directory.
// It works both when running from the repo root and from the backend directory.
func findMigrationsDir() (string, error) {
	candidates := []string{
		"backend/migrations",
		"migrations",
	}

	_, filename, _, ok := runtime.Caller(0)
	if ok {
		srcDir := filepath.Dir(filename)
		candidates = append(candidates, filepath.Join(srcDir, "..", "..", "migrations"))
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
	}

	return "", fmt.Errorf("migrations directory not found (tried %v)", candidates)
}
