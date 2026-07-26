package repository

import (
	"database/sql"
	"time"
)

type SessionRepo struct {
	db *sql.DB
}

func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(id string, userID int64, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)`,
		id, userID, expiresAt,
	)
	return err
}

// GetUserID returns the user id for a valid, non-expired session.
func (r *SessionRepo) GetUserID(id string) (int64, error) {
	var userID int64
	var expiresAt time.Time
	err := r.db.QueryRow(`SELECT user_id, expires_at FROM sessions WHERE id = $1`, id).Scan(&userID, &expiresAt)
	if err == sql.ErrNoRows {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	if time.Now().After(expiresAt) {
		_ = r.Delete(id)
		return 0, ErrNotFound
	}
	return userID, nil
}

func (r *SessionRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id = $1`, id)
	return err
}
