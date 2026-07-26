package repository

import (
	"database/sql"
	"errors"

	"jobhoo/internal/models"
)

var ErrNotFound = errors.New("record not found")

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateUser(email, passwordHash string, role models.UserRole) (int64, error) {
	var id int64
	err := r.db.QueryRow(
		`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3) RETURNING id`,
		email, passwordHash, role,
	).Scan(&id)
	return id, err
}

func (r *UserRepo) GetByEmail(email string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, role, created_at, updated_at FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *UserRepo) GetByID(id int64) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, role, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return u, err
}

// --- Candidate profile ---

func (r *UserRepo) CreateCandidateProfile(userID int64, fullName, resumePath, resumeFilename string) error {
	_, err := r.db.Exec(
		`INSERT INTO candidate_profiles (user_id, full_name, resume_path, resume_filename) VALUES ($1, $2, $3, $4)`,
		userID, fullName, resumePath, resumeFilename,
	)
	return err
}

func (r *UserRepo) GetCandidateProfile(userID int64) (*models.CandidateProfile, error) {
	p := &models.CandidateProfile{}
	err := r.db.QueryRow(
		`SELECT id, user_id, full_name, resume_path, resume_filename, created_at, updated_at
		 FROM candidate_profiles WHERE user_id = $1`, userID,
	).Scan(&p.ID, &p.UserID, &p.FullName, &p.ResumePath, &p.ResumeFilename, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *UserRepo) UpdateCandidateResume(userID int64, resumePath, resumeFilename string) error {
	_, err := r.db.Exec(
		`UPDATE candidate_profiles SET resume_path = $1, resume_filename = $2, updated_at = now() WHERE user_id = $3`,
		resumePath, resumeFilename, userID,
	)
	return err
}

func (r *UserRepo) CountCandidates() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'candidate'`).Scan(&n)
	return n, err
}

func (r *UserRepo) ListCandidates() ([]models.User, error) {
	return r.ListCandidatesPaginated(1000, 0)
}

func (r *UserRepo) ListCandidatesPaginated(limit, offset int) ([]models.User, error) {
	rows, err := r.db.Query(`SELECT id, email, password_hash, role, created_at, updated_at FROM users WHERE role = 'candidate' ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// --- Recruiter profile ---

func (r *UserRepo) CreateRecruiterProfile(userID int64, companyName, docPath, docFilename string) error {
	_, err := r.db.Exec(
		`INSERT INTO recruiter_profiles (user_id, company_name, document_path, document_filename) VALUES ($1, $2, $3, $4)`,
		userID, companyName, docPath, docFilename,
	)
	return err
}

func (r *UserRepo) GetRecruiterProfile(userID int64) (*models.RecruiterProfile, error) {
	p := &models.RecruiterProfile{}
	err := r.db.QueryRow(
		`SELECT id, user_id, company_name, document_path, document_filename, status, reviewed_by, reviewed_at, created_at, updated_at
		 FROM recruiter_profiles WHERE user_id = $1`, userID,
	).Scan(&p.ID, &p.UserID, &p.CompanyName, &p.DocumentPath, &p.DocumentFilename, &p.Status, &p.ReviewedBy, &p.ReviewedAt, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return p, err
}

type RecruiterWithUser struct {
	models.RecruiterProfile
	Email string
}

func (r *UserRepo) ListRecruitersByStatus(status models.RecruiterStatus) ([]RecruiterWithUser, error) {
	return r.ListRecruitersByStatusPaginated(status, 1000, 0)
}

func (r *UserRepo) ListRecruitersByStatusPaginated(status models.RecruiterStatus, limit, offset int) ([]RecruiterWithUser, error) {
	rows, err := r.db.Query(
		`SELECT rp.id, rp.user_id, rp.company_name, rp.document_path, rp.document_filename, rp.status,
		        rp.reviewed_by, rp.reviewed_at, rp.created_at, rp.updated_at, u.email
		 FROM recruiter_profiles rp
		 JOIN users u ON u.id = rp.user_id
		 WHERE rp.status = $1
		 ORDER BY rp.created_at DESC
		 LIMIT $2 OFFSET $3`, status, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RecruiterWithUser
	for rows.Next() {
		var p RecruiterWithUser
		if err := rows.Scan(&p.ID, &p.UserID, &p.CompanyName, &p.DocumentPath, &p.DocumentFilename, &p.Status,
			&p.ReviewedBy, &p.ReviewedAt, &p.CreatedAt, &p.UpdatedAt, &p.Email); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *UserRepo) ListAllRecruiters() ([]RecruiterWithUser, error) {
	return r.ListAllRecruitersPaginated(1000, 0)
}

func (r *UserRepo) ListAllRecruitersPaginated(limit, offset int) ([]RecruiterWithUser, error) {
	rows, err := r.db.Query(
		`SELECT rp.id, rp.user_id, rp.company_name, rp.document_path, rp.document_filename, rp.status,
		        rp.reviewed_by, rp.reviewed_at, rp.created_at, rp.updated_at, u.email
		 FROM recruiter_profiles rp
		 JOIN users u ON u.id = rp.user_id
		 ORDER BY rp.created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RecruiterWithUser
	for rows.Next() {
		var p RecruiterWithUser
		if err := rows.Scan(&p.ID, &p.UserID, &p.CompanyName, &p.DocumentPath, &p.DocumentFilename, &p.Status,
			&p.ReviewedBy, &p.ReviewedAt, &p.CreatedAt, &p.UpdatedAt, &p.Email); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *UserRepo) UpdateRecruiterStatus(userID, reviewerID int64, status models.RecruiterStatus) error {
	_, err := r.db.Exec(
		`UPDATE recruiter_profiles SET status = $1, reviewed_by = $2, reviewed_at = now(), updated_at = now() WHERE user_id = $3`,
		status, reviewerID, userID,
	)
	return err
}

func (r *UserRepo) CountRecruiters() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM recruiter_profiles`).Scan(&n)
	return n, err
}

func (r *UserRepo) CountRecruitersByStatus(status models.RecruiterStatus) (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM recruiter_profiles WHERE status = $1`, status).Scan(&n)
	return n, err
}
