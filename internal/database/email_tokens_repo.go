package database

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailTokensRepo struct {
	pool *pgxpool.Pool
}

func NewEmailTokensRepo(pool *pgxpool.Pool) *EmailTokensRepo {
	return &EmailTokensRepo{pool: pool}
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// CreateEmailVerification stores a single-use verification token (store hash)
func (r *EmailTokensRepo) CreateEmailVerification(ctx context.Context, userID, rawToken string, ttl time.Duration) error {
	tokenHash := hashToken(rawToken)
	_, err := r.pool.Exec(ctx, `
        INSERT INTO email_verifications (user_id, token_hash, expires_at)
        VALUES ($1, $2, now() + make_interval(secs => $3))
    `, userID, tokenHash, int(ttl.Seconds()))
	return err
}

// ConsumeEmailVerification validates and marks a verification token used; returns user_id
func (r *EmailTokensRepo) ConsumeEmailVerification(ctx context.Context, rawToken string) (string, error) {
	tokenHash := hashToken(rawToken)
	var userID string
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT user_id FROM email_verifications WHERE token_hash = $1 AND used = FALSE AND expires_at > now()`, tokenHash)
	if err := row.Scan(&userID); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `UPDATE email_verifications SET used = TRUE WHERE token_hash = $1`, tokenHash); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return userID, nil
}

// CreatePasswordReset stores a single-use reset token
func (r *EmailTokensRepo) CreatePasswordReset(ctx context.Context, userID, rawToken string, ttl time.Duration) error {
	tokenHash := hashToken(rawToken)
	_, err := r.pool.Exec(ctx, `
        INSERT INTO password_resets (user_id, token_hash, expires_at)
        VALUES ($1, $2, now() + make_interval(secs => $3))
    `, userID, tokenHash, int(ttl.Seconds()))
	return err
}

// ConsumePasswordReset validates and marks a reset token used; returns user_id
func (r *EmailTokensRepo) ConsumePasswordReset(ctx context.Context, rawToken string) (string, error) {
	tokenHash := hashToken(rawToken)
	var userID string
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT user_id FROM password_resets WHERE token_hash = $1 AND used = FALSE AND expires_at > now()`, tokenHash)
	if err := row.Scan(&userID); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `UPDATE password_resets SET used = TRUE WHERE token_hash = $1`, tokenHash); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return userID, nil
}
