package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobhoo/jobhoo/internal/ai"
)

// AICallLogsRepo persists every AI provider call for the admin Activity
// page and implements ai.CallLogger.
type AICallLogsRepo struct {
	pool *pgxpool.Pool
}

func NewAICallLogsRepo(pool *pgxpool.Pool) *AICallLogsRepo {
	return &AICallLogsRepo{pool: pool}
}

// Log persists one AI call. It's best-effort: a logging failure must never
// break the AI feature itself, so errors are only printed, not returned.
func (r *AICallLogsRepo) Log(ctx context.Context, entry ai.AICallLog) {
	var userID *string
	if entry.UserID != "" {
		userID = &entry.UserID
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ai_call_logs
			(created_at, provider, method, status, execution_ms, tokens_used, prompt, response, error_message, user_id, user_name, user_email, user_role)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, entry.Timestamp, entry.Provider, entry.Method, entry.Status, entry.ExecutionMs, entry.TokensUsed,
		entry.Prompt, entry.Response, entry.Error, userID, entry.UserName, entry.UserEmail, entry.UserRole)
	if err != nil {
		log.Printf("ai_call_logs: failed to persist entry: %v", err)
	}
}

// AICallLogRow is one row for the admin Activity page.
type AICallLogRow struct {
	ID           string
	CreatedAt    time.Time
	Provider     string
	Method       string
	Status       string
	ExecutionMs  int
	TokensUsed   int
	Prompt       string
	Response     string
	ErrorMessage string
	UserName     string
	UserEmail    string
	UserRole     string
}

// List returns the most recent AI call logs, newest first.
func (r *AICallLogsRepo) List(ctx context.Context, limit, offset int) ([]AICallLogRow, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM ai_call_logs`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, created_at, provider, method, status, execution_ms, tokens_used,
		       coalesce(prompt,''), coalesce(response,''), coalesce(error_message,''),
		       coalesce(user_name,''), coalesce(user_email,''), coalesce(user_role,'')
		FROM ai_call_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []AICallLogRow
	for rows.Next() {
		var row AICallLogRow
		if err := rows.Scan(&row.ID, &row.CreatedAt, &row.Provider, &row.Method, &row.Status,
			&row.ExecutionMs, &row.TokensUsed, &row.Prompt, &row.Response, &row.ErrorMessage,
			&row.UserName, &row.UserEmail, &row.UserRole); err != nil {
			return nil, 0, err
		}
		out = append(out, row)
	}
	return out, total, rows.Err()
}
