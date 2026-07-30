package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobhoo/jobhoo/internal/models"
)

type CompaniesRepo struct {
	pool *pgxpool.Pool
}

func NewCompaniesRepo(pool *pgxpool.Pool) *CompaniesRepo {
	return &CompaniesRepo{pool: pool}
}

func (r *CompaniesRepo) GetByOwnerID(ctx context.Context, ownerID string) (*models.Company, error) {
	var c models.Company
	err := r.pool.QueryRow(ctx, `
		SELECT id, owner_id, name, coalesce(logo_url, ''), coalesce(website, ''), coalesce(description, ''), coalesce(industry, ''), created_at
		FROM companies WHERE owner_id = $1
		ORDER BY created_at ASC LIMIT 1
	`, ownerID).Scan(&c.ID, &c.OwnerID, &c.Name, &c.LogoURL, &c.Website, &c.Description, &c.Industry, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *CompaniesRepo) Create(ctx context.Context, ownerID, name, website, description, industry string) (*models.Company, error) {
	var c models.Company
	err := r.pool.QueryRow(ctx, `
		INSERT INTO companies (owner_id, name, website, description, industry)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, owner_id, name, coalesce(logo_url, ''), coalesce(website, ''), coalesce(description, ''), coalesce(industry, ''), created_at
	`, ownerID, name, website, description, industry).Scan(
		&c.ID, &c.OwnerID, &c.Name, &c.LogoURL, &c.Website, &c.Description, &c.Industry, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// CompanyWithJobCount pairs a company with how many published jobs it
// currently has open, for the public company directory.
type CompanyWithJobCount struct {
	models.Company
	OpenJobCount int
}

// ListAll returns every company, most-actively-hiring first (published job
// count descending), for the public "Explore Companies" directory.
func (r *CompaniesRepo) ListAll(ctx context.Context) ([]CompanyWithJobCount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			c.id, c.owner_id, c.name, coalesce(c.logo_url, ''), coalesce(c.website, ''),
			coalesce(c.description, ''), coalesce(c.industry, ''), c.created_at,
			count(j.id) FILTER (WHERE j.status = 'published') AS open_job_count
		FROM companies c
		LEFT JOIN jobs j ON j.company_id = c.id
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
			&c.Description, &c.Industry, &c.CreatedAt, &c.OpenJobCount,
		); err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}
	return companies, rows.Err()
}
