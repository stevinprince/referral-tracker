package repository

import "github.com/jackc/pgx/v5/pgxpool"

// TestPool wraps a pgxpool.Pool for use in integration tests from other packages.
// This avoids test packages needing to import pgxpool directly.
type TestPool struct {
	Pool *pgxpool.Pool
}
