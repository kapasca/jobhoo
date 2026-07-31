package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobhoo/jobhoo/internal/models"
)

const companySelectColumns = `
	id, owner_id, name, coalesce(logo_url, ''), coalesce(website, ''),
	coalesce(description, ''), coalesce(industry, ''), status, approved_at,
	coalesce(approved_by::text, ''), coalesce(rejection_reason, ''), created_at
`

type CompaniesRepo struct {
	pool *pgxpool.Pool
}

func NewCompaniesRepo(pool *pgxpool.Pool) *CompaniesRepo {
	return &CompaniesRepo{pool: pool}
}

func scanCompany(row rowScanner) (models.Company, error) {
	var c models.Company
	err := row.Scan(
		&c.ID, &c.OwnerID, &c.Name, &c.LogoURL, &c.Website,
		&c.Description, &c.Industry, &c.Status, &c.ApprovedAt,
		&c.ApprovedBy, &c.RejectionReason, &c.CreatedAt,
	)
	return c, err
}

func (r *CompaniesRepo) GetByOwnerID(ctx context.Context, ownerID string) (*models.Company, error) {
	c, err := scanCompany(r.pool.QueryRow(ctx, `
		SELECT `+companySelectColumns+`
		FROM companies WHERE owner_id = $1
		ORDER BY created_at ASC LIMIT 1
	`, ownerID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *CompaniesRepo) GetByID(ctx context.Context, id string) (*models.Company, error) {
	c, err := scanCompany(r.pool.QueryRow(ctx, `
		SELECT `+companySelectColumns+`
		FROM companies WHERE id = $1
	`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

// Create inserts a new company in 'pending' status (the database default) —
// it cannot post jobs until an admin approves it. See requireApprovedCompany
// in the handlers package for where this is enforced.
func (r *CompaniesRepo) Create(ctx context.Context, ownerID, name, website, description, industry, logoURL string) (*models.Company, error) {
	c, err := scanCompany(r.pool.QueryRow(ctx, `
		INSERT INTO companies (owner_id, name, website, description, industry, logo_url)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6,''))
		RETURNING `+companySelectColumns, ownerID, name, website, description, industry, logoURL))
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Update lets a recruiter edit their own company's public profile.
// Deliberately does not touch status/approval fields — only an admin action
// (Approve/Reject) can change those.
func (r *CompaniesRepo) Update(ctx context.Context, id, name, website, description, industry, logoURL string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE companies SET name = $1, website = $2, description = $3, industry = $4, logo_url = $5
		WHERE id = $6
	`, name, website, description, industry, logoURL, id)
	return err
}

// CompanyWithJobCount pairs a company with how many published jobs it
// currently has open, for the public company directory.
type CompanyWithJobCount struct {
	models.Company
	OpenJobCount int
}

// ListAll returns every APPROVED company, most-actively-hiring first, for
// the public "Explore Companies" directory. Pending/rejected companies are
// never shown publicly — only their own recruiter (via the dashboard) and
// admins (via the approval queue) can see those.
func (r *CompaniesRepo) ListAll(ctx context.Context) ([]CompanyWithJobCount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			c.id, c.owner_id, c.name, coalesce(c.logo_url, ''), coalesce(c.website, ''),
			coalesce(c.description, ''), coalesce(c.industry, ''), c.status, c.approved_at,
			coalesce(c.approved_by::text, ''), coalesce(c.rejection_reason, ''), c.created_at,
			count(j.id) FILTER (WHERE j.status = 'published') AS open_job_count
		FROM companies c
		LEFT JOIN jobs j ON j.company_id = c.id
		WHERE c.status = 'approved'
		GROUP BY c.id
		ORDER BY open_job_count DESC, c.name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []CompanyWithJobCount
	for rows.Next() {
		var c CompanyWithJobCount
		if err := rows.Scan(
			&c.ID, &c.OwnerID, &c.Name, &c.LogoURL, &c.Website,
			&c.Description, &c.Industry, &c.Status, &c.ApprovedAt,
			&c.ApprovedBy, &c.RejectionReason, &c.CreatedAt, &c.OpenJobCount,
		); err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}
	return companies, rows.Err()
}

// CompanyWithOwner extends Company with the registering recruiter's details
// for the admin approval queue.
type CompanyWithOwner struct {
	models.Company
	OwnerName  string
	OwnerEmail string
}
// ListAllRegistrations returns one page of companies (any status) with owner
// info and the total count, newest first, for the admin registration log.
func (r *CompaniesRepo) ListAllRegistrations(ctx context.Context, limit, offset int) ([]CompanyWithOwner, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM companies`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.owner_id, c.name, coalesce(c.logo_url,''), coalesce(c.website,''),
		       coalesce(c.description,''), coalesce(c.industry,''), c.status, c.approved_at,
		       coalesce(c.approved_by::text,''), coalesce(c.rejection_reason,''), c.created_at,
		       u.full_name, u.email
		FROM companies c
		JOIN users u ON u.id = c.owner_id
		ORDER BY c.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []CompanyWithOwner
	for rows.Next() {
		var co CompanyWithOwner
		if err := rows.Scan(
			&co.ID, &co.OwnerID, &co.Name, &co.LogoURL, &co.Website,
			&co.Description, &co.Industry, &co.Status, &co.ApprovedAt,
			&co.ApprovedBy, &co.RejectionReason, &co.CreatedAt,
			&co.OwnerName, &co.OwnerEmail,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, co)
	}
	return out, total, rows.Err()
}
// ListPendingWithOwner returns pending companies joined with their owner’s
// name and email, oldest-first.
func (r *CompaniesRepo) ListPendingWithOwner(ctx context.Context) ([]CompanyWithOwner, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.owner_id, c.name, coalesce(c.logo_url,''), coalesce(c.website,''),
		       coalesce(c.description,''), coalesce(c.industry,''), c.status, c.approved_at,
		       coalesce(c.approved_by::text,''), coalesce(c.rejection_reason,''), c.created_at,
		       u.full_name, u.email
		FROM companies c
		JOIN users u ON u.id = c.owner_id
		WHERE c.status = 'pending'
		ORDER BY c.created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CompanyWithOwner
	for rows.Next() {
		var co CompanyWithOwner
		if err := rows.Scan(
			&co.ID, &co.OwnerID, &co.Name, &co.LogoURL, &co.Website,
			&co.Description, &co.Industry, &co.Status, &co.ApprovedAt,
			&co.ApprovedBy, &co.RejectionReason, &co.CreatedAt,
			&co.OwnerName, &co.OwnerEmail,
		); err != nil {
			return nil, err
		}
		out = append(out, co)
	}
	return out, rows.Err()
}

// ListPending returns companies awaiting admin review, oldest first (so the
// longest-waiting applicant is reviewed first).
func (r *CompaniesRepo) ListPending(ctx context.Context) ([]models.Company, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+companySelectColumns+`
		FROM companies WHERE status = 'pending'
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []models.Company
	for rows.Next() {
		c, err := scanCompany(rows)
		if err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}
	return companies, rows.Err()
}

// Approve marks a company approved by the given admin user ID.
func (r *CompaniesRepo) Approve(ctx context.Context, companyID, adminUserID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE companies SET status = 'approved', approved_at = now(), approved_by = $1, rejection_reason = NULL
		WHERE id = $2
	`, adminUserID, companyID)
	return err
}

// Reject marks a company rejected with a reason, by the given admin user ID.
func (r *CompaniesRepo) Reject(ctx context.Context, companyID, adminUserID, reason string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE companies SET status = 'rejected', approved_at = now(), approved_by = $1, rejection_reason = $2
		WHERE id = $3
	`, adminUserID, reason, companyID)
	return err
}

// Blacklist permanently blocks a company; its recruiter cannot re-apply.
func (r *CompaniesRepo) Blacklist(ctx context.Context, companyID, adminUserID, reason string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE companies SET status = 'blacklisted', approved_at = now(), approved_by = $1, rejection_reason = $2
		WHERE id = $3
	`, adminUserID, reason, companyID)
	return err
}

// Resubmit resets a rejected company back to pending so the recruiter can
// request approval again. Only works on rejected companies — blacklisted
// companies cannot resubmit.
func (r *CompaniesRepo) Resubmit(ctx context.Context, companyID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE companies SET status = 'pending', approved_at = NULL, approved_by = NULL, rejection_reason = NULL
		WHERE id = $1 AND status = 'rejected'
	`, companyID)
	return err
}
