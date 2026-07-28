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
