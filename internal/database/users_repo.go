package database

import (
	"context"
	"errors"
	"time"

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
		RETURNING id, email, password_hash, role, full_name, coalesce(avatar_url, ''), is_frozen, email_verified, created_at
	`, email, passwordHash, role, fullName).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.FullName, &u.AvatarURL, &u.IsFrozen, &u.EmailVerified, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, role, full_name, coalesce(avatar_url, ''), is_frozen, email_verified, created_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.FullName, &u.AvatarURL, &u.IsFrozen, &u.EmailVerified, &u.CreatedAt)
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
		SELECT id, email, password_hash, role, full_name, coalesce(avatar_url, ''), is_frozen, email_verified, created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.FullName, &u.AvatarURL, &u.IsFrozen, &u.EmailVerified, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

// CandidateRegistration is a lightweight row for the admin registration log.
type CandidateRegistration struct {
	ID            string
	FullName      string
	Email         string
	EmailVerified bool
	CreatedAt     time.Time
}

// ListCandidateRegistrations returns one page of candidate accounts and the total count.
func (r *UsersRepo) ListCandidateRegistrations(ctx context.Context, limit, offset int) ([]CandidateRegistration, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE role = 'candidate'`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, full_name, email, email_verified, created_at
		FROM users WHERE role = 'candidate'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []CandidateRegistration
	for rows.Next() {
		var c CandidateRegistration
		if err := rows.Scan(&c.ID, &c.FullName, &c.Email, &c.EmailVerified, &c.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, rows.Err()
}

// ListRecruiterRegistrations returns one page of recruiter accounts, newest first.
func (r *UsersRepo) ListRecruiterRegistrations(ctx context.Context, limit, offset int) ([]CandidateRegistration, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE role = 'recruiter'`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, full_name, email, email_verified, created_at FROM users WHERE role = 'recruiter'
		ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []CandidateRegistration
	for rows.Next() {
		var c CandidateRegistration
		if err := rows.Scan(&c.ID, &c.FullName, &c.Email, &c.EmailVerified, &c.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, rows.Err()
}

// UserRow is a lightweight row for the admin user management table.
type UserRow struct {
	ID            string
	FullName      string
	Email         string
	Role          models.UserRole
	IsFrozen      bool
	EmailVerified bool
	CreatedAt     time.Time
}

// ListAllUsers returns one page of all users for the admin panel.
func (r *UsersRepo) ListAllUsers(ctx context.Context, limit, offset int) ([]UserRow, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, full_name, email, role, is_frozen, email_verified, created_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []UserRow
	for rows.Next() {
		var u UserRow
		if err := rows.Scan(&u.ID, &u.FullName, &u.Email, &u.Role, &u.IsFrozen, &u.EmailVerified, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	return out, total, rows.Err()
}

func (r *UsersRepo) FreezeUser(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET is_frozen = TRUE WHERE id = $1 AND role != 'admin'`, id)
	return err
}

func (r *UsersRepo) UnfreezeUser(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET is_frozen = FALSE WHERE id = $1`, id)
	return err
}

// SetEmailVerified marks a user's email as verified.
func (r *UsersRepo) SetEmailVerified(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET email_verified = TRUE WHERE id = $1`, id)
	return err
}

// SetPasswordHash atomically sets a new bcrypt password hash for the user.
func (r *UsersRepo) SetPasswordHash(ctx context.Context, id, hash string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, hash, id)
	return err
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
