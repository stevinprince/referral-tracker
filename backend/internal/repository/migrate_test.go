package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func testDBURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}
	return url
}

// cleanupDB drops test tables so each test starts fresh.
func cleanupDB(t *testing.T, ctx context.Context, dbURL string) {
	t.Helper()
	pool, err := ConnectDB(ctx, dbURL)
	if err != nil {
		t.Fatalf("connecting to test DB for cleanup: %v", err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, `
		DROP TABLE IF EXISTS jobs CASCADE;
		DROP TABLE IF EXISTS sessions CASCADE;
		DROP TABLE IF EXISTS schema_migrations CASCADE;
	`)
	if err != nil {
		t.Fatalf("cleaning up test DB: %v", err)
	}
}

func TestRunMigrations_AppliesAll(t *testing.T) {
	dbURL := testDBURL(t)
	ctx := context.Background()
	cleanupDB(t, ctx, dbURL)

	pool, err := ConnectDB(ctx, dbURL)
	if err != nil {
		t.Fatalf("connecting to DB: %v", err)
	}
	defer pool.Close()

	migrationsDir := findMigrationsDir(t)

	if err := RunMigrations(ctx, pool, migrationsDir); err != nil {
		t.Fatalf("running migrations: %v", err)
	}

	// Verify jobs table exists with expected columns
	var colCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM information_schema.columns 
		WHERE table_schema = current_schema() AND table_name = 'jobs'
	`).Scan(&colCount)
	if err != nil {
		t.Fatalf("querying jobs columns: %v", err)
	}
	if colCount != 13 {
		t.Errorf("jobs table has %d columns, want 13", colCount)
	}

	// Verify sessions table exists
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM information_schema.columns 
		WHERE table_schema = current_schema() AND table_name = 'sessions'
	`).Scan(&colCount)
	if err != nil {
		t.Fatalf("querying sessions columns: %v", err)
	}
	if colCount != 3 {
		t.Errorf("sessions table has %d columns, want 3", colCount)
	}

	// Verify both migrations were recorded
	var migrationCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&migrationCount)
	if err != nil {
		t.Fatalf("querying schema_migrations: %v", err)
	}
	if migrationCount != 2 {
		t.Errorf("schema_migrations has %d rows, want 2", migrationCount)
	}
}

func TestRunMigrations_Idempotent(t *testing.T) {
	dbURL := testDBURL(t)
	ctx := context.Background()
	cleanupDB(t, ctx, dbURL)

	pool, err := ConnectDB(ctx, dbURL)
	if err != nil {
		t.Fatalf("connecting to DB: %v", err)
	}
	defer pool.Close()

	migrationsDir := findMigrationsDir(t)

	// Run migrations twice
	if err := RunMigrations(ctx, pool, migrationsDir); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if err := RunMigrations(ctx, pool, migrationsDir); err != nil {
		t.Fatalf("second run (should be idempotent): %v", err)
	}

	// Still only 2 recorded migrations
	var migrationCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&migrationCount)
	if err != nil {
		t.Fatalf("querying schema_migrations: %v", err)
	}
	if migrationCount != 2 {
		t.Errorf("schema_migrations has %d rows after two runs, want 2", migrationCount)
	}
}

func TestRunMigrations_InvalidDir(t *testing.T) {
	dbURL := testDBURL(t)
	ctx := context.Background()

	pool, err := ConnectDB(ctx, dbURL)
	if err != nil {
		t.Fatalf("connecting to DB: %v", err)
	}
	defer pool.Close()

	err = RunMigrations(ctx, pool, "/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent migrations dir, got nil")
	}
}

func TestGetMigrationFiles_FiltersNonMigrations(t *testing.T) {
	dir := t.TempDir()

	// Create a mix of files: valid migrations and files that should be excluded
	validFiles := []string{"001_create_jobs.sql", "002_create_sessions.sql"}
	invalidFiles := []string{"scratch.sql", "backup.sql", "README.md", "notes.txt", ".hidden.sql", "003_notes.txt"}

	for _, f := range append(validFiles, invalidFiles...) {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("-- stub"), 0644); err != nil {
			t.Fatalf("creating test file %s: %v", f, err)
		}
	}

	got, err := getMigrationFiles(dir)
	if err != nil {
		t.Fatalf("getMigrationFiles: %v", err)
	}

	if len(got) != len(validFiles) {
		t.Fatalf("got %d files %v, want %d files %v", len(got), got, len(validFiles), validFiles)
	}
	for i, want := range validFiles {
		if got[i] != want {
			t.Errorf("file[%d] = %q, want %q", i, got[i], want)
		}
	}
}

// findMigrationsDir locates the migrations directory relative to the test file.
func findMigrationsDir(t *testing.T) string {
	t.Helper()
	// Tests run from the package directory, migrations are at ../../migrations
	dir := filepath.Join("..", "..", "migrations")
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("migrations directory not found at %s: %v", dir, err)
	}
	return dir
}
