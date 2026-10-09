package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Session represents a row in the sessions table.
type Session struct {
	ID        string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionRepository handles database operations for sessions.
type SessionRepository struct {
	pool *pgxpool.Pool
}

// NewSessionRepository creates a new SessionRepository.
func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

// Create inserts a new session into the database.
func (r *SessionRepository) Create(ctx context.Context, id string, expiresAt time.Time) (*Session, error) {
	s := &Session{
		ID:        id,
		ExpiresAt: expiresAt,
	}

	err := r.pool.QueryRow(ctx,
		`INSERT INTO sessions (id, expires_at) VALUES ($1, $2) RETURNING created_at`,
		id, expiresAt,
	).Scan(&s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting session: %w", err)
	}

	return s, nil
}

// FindByID looks up a session by its token ID. Returns nil if not found.
func (r *SessionRepository) FindByID(ctx context.Context, id string) (*Session, error) {
	s := &Session{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, created_at, expires_at FROM sessions WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.CreatedAt, &s.ExpiresAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding session: %w", err)
	}

	return s, nil
}

// Delete removes a session by its token ID.
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}

// DeleteExpired removes all sessions that have passed their expiry time.
func (r *SessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at < NOW()`)
	if err != nil {
		return 0, fmt.Errorf("deleting expired sessions: %w", err)
	}
	return tag.RowsAffected(), nil
}
