package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailLog struct {
	ID             string
	UserID         *string
	RecipientEmail string
	Subject        string
	EmailType      string // "verification", "password_reset", "application_status", etc.
	Status         string // "sent", "failed"
	ErrorMessage   *string
	SentAt         time.Time
	CreatedAt      time.Time
}

type EmailLogsRepo struct {
	pool *pgxpool.Pool
}

func NewEmailLogsRepo(pool *pgxpool.Pool) *EmailLogsRepo {
	return &EmailLogsRepo{pool: pool}
}

// Log records an email sent attempt
func (r *EmailLogsRepo) Log(ctx context.Context, userID *string, recipientEmail, subject, emailType, status string, errorMsg *string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO email_logs (user_id, recipient_email, subject, email_type, status, error_message)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, recipientEmail, subject, emailType, status, errorMsg)
	return err
}

// ListByUser retrieves email logs for a specific user
func (r *EmailLogsRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]EmailLog, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM email_logs WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, recipient_email, subject, email_type, status, error_message, sent_at, created_at
		FROM email_logs
		WHERE user_id = $1
		ORDER BY sent_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []EmailLog
	for rows.Next() {
		var log EmailLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.RecipientEmail, &log.Subject, &log.EmailType, &log.Status, &log.ErrorMessage, &log.SentAt, &log.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}
	return logs, total, rows.Err()
}

// ListAll retrieves all email logs (for admin)
func (r *EmailLogsRepo) ListAll(ctx context.Context, limit, offset int) ([]EmailLog, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM email_logs`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, recipient_email, subject, email_type, status, error_message, sent_at, created_at
		FROM email_logs
		ORDER BY sent_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []EmailLog
	for rows.Next() {
		var log EmailLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.RecipientEmail, &log.Subject, &log.EmailType, &log.Status, &log.ErrorMessage, &log.SentAt, &log.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}
	return logs, total, rows.Err()
}
