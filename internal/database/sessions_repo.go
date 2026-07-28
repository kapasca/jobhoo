package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionsRepo struct {
	pool *pgxpool.Pool
}

func NewSessionsRepo(pool *pgxpool.Pool) *SessionsRepo {
	return &SessionsRepo{pool: pool}
}

// Create stores a new session keyed by the SHA-256 hash of the raw cookie
// token. The raw token itself is never persisted — only its hash — so a
// database leak alone can't be used to hijack sessions.
func (r *SessionsRepo) Create(ctx context.Context, userID, tokenHash, userAgent, ipAddress string, ttl time.Duration) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions (user_id, token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, now() + make_interval(secs => $5))
	`, userID, tokenHash, userAgent, ipAddress, int(ttl.Seconds()))
	return err
}

// UserIDForToken returns the user ID for a valid, non-expired session, or
// ErrNotFound if the token is missing/expired/revoked.
func (r *SessionsRepo) UserIDForToken(ctx context.Context, tokenHash string) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx, `
		SELECT user_id FROM sessions
		WHERE token_hash = $1 AND expires_at > now()
	`, tokenHash).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return userID, nil
}

// Revoke deletes a single session (logout on this device).
func (r *SessionsRepo) Revoke(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	return err
}

// RevokeAllForUser deletes every session for a user (logout everywhere,
// password change, account suspension).
func (r *SessionsRepo) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}
