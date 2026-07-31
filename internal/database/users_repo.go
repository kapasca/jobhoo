package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobhoo/jobhoo/internal/models"
)

var ErrNotFound = errors.New("not found")
var ErrDuplicateEmail = errors.New("email already registered")

type UsersRepo struct {
	pool *pgxpool.Pool
}

func NewUsersRepo(pool *pgxpool.Pool) *UsersRepo {
	return &UsersRepo{pool: pool}
}

func (r *UsersRepo) Create(ctx context.Context, email, passwordHash, fullName string, role models.UserRole) (*models.User, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateEmail
	}

	var u models.User
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role, full_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, password_hash, role, full_name, coalesce(avatar_url, ''), created_at
	`, email, passwordHash, role, fullName).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.FullName, &u.AvatarURL, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, role, full_name, coalesce(avatar_url, ''), created_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.FullName, &u.AvatarURL, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UsersRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, role, full_name, coalesce(avatar_url, ''), created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.FullName, &u.AvatarURL, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

type PlatformStats struct {
	TotalUsers          int
	TotalCandidates     int
	TotalRecruiters     int
	TotalCompanies      int
	TotalJobs           int
	TotalApplications   int
	PendingCompanyCount int
}

// PlatformStats gives the admin dashboard a single-query snapshot of
// platform scale. Kept intentionally simple (counts only, no charts) per
// the brief's "avoid enterprise complexity" guidance for the admin view.
func (r *UsersRepo) PlatformStats(ctx context.Context) (PlatformStats, error) {
	var s PlatformStats
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM users),
			(SELECT count(*) FROM users WHERE role = 'candidate'),
			(SELECT count(*) FROM users WHERE role = 'recruiter'),
			(SELECT count(*) FROM companies),
			(SELECT count(*) FROM jobs),
			(SELECT count(*) FROM applications),
			(SELECT count(*) FROM companies WHERE status = 'pending')
	`).Scan(&s.TotalUsers, &s.TotalCandidates, &s.TotalRecruiters, &s.TotalCompanies,
		&s.TotalJobs, &s.TotalApplications, &s.PendingCompanyCount)
	return s, err
}
