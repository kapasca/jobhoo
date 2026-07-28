package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobhoo/jobhoo/internal/models"
)

type JobsRepo struct {
	pool *pgxpool.Pool
}

func NewJobsRepo(pool *pgxpool.Pool) *JobsRepo {
	return &JobsRepo{pool: pool}
}

const jobSelectColumns = `
	j.id, j.title, j.description, j.location, j.employment_type, j.work_arrangement,
	j.category, j.salary_min, j.salary_max, j.salary_currency, j.must_have_skills,
	j.nice_to_have_skills, j.seniority, j.status, j.published_at, j.created_at,
	c.id, c.name, c.logo_url
`

// JobListFilter narrows ListPublished's results. Zero-value fields are
// treated as "no filter" (e.g. empty Category means all categories).
type JobListFilter struct {
	Search   string
	Category models.JobCategory
	Limit    int
	Offset   int
}

// JobListResult carries a page of jobs plus the total count matching the
// filter (ignoring Limit/Offset), so callers can render pagination controls.
type JobListResult struct {
	Jobs  []models.Job
	Total int
}

// ListPublished returns a page of published jobs, most recent first,
// optionally filtered by free-text search and/or category.
func (r *JobsRepo) ListPublished(ctx context.Context, f JobListFilter) (JobListResult, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}

	countQuery := `
		SELECT count(*)
		FROM jobs j
		WHERE j.status = 'published'
		  AND ($1 = '' OR j.title ILIKE '%' || $1 || '%' OR $1 ILIKE ANY(j.must_have_skills))
		  AND ($2 = '' OR j.category = $2::job_category)
	`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, f.Search, string(f.Category)).Scan(&total); err != nil {
		return JobListResult{}, err
	}

	query := `
		SELECT ` + jobSelectColumns + `
		FROM jobs j
		JOIN companies c ON c.id = j.company_id
		WHERE j.status = 'published'
		  AND ($1 = '' OR j.title ILIKE '%' || $1 || '%' OR $1 ILIKE ANY(j.must_have_skills))
		  AND ($2 = '' OR j.category = $2::job_category)
		ORDER BY j.published_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.pool.Query(ctx, query, f.Search, string(f.Category), f.Limit, f.Offset)
	if err != nil {
		return JobListResult{}, err
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return JobListResult{}, err
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return JobListResult{}, err
	}

	return JobListResult{Jobs: jobs, Total: total}, nil
}

// ListByCompany returns every job (any status) owned by a company, newest
// first, for the recruiter dashboard's job list.
func (r *JobsRepo) ListByCompany(ctx context.Context, companyID string) ([]models.Job, error) {
	query := `
		SELECT ` + jobSelectColumns + `
		FROM jobs j
		JOIN companies c ON c.id = j.company_id
		WHERE j.company_id = $1
		ORDER BY j.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

// CreateJobInput carries every recruiter-editable field for posting a job.
type CreateJobInput struct {
	CompanyID        string
	CreatedBy        string
	Title            string
	Description      string
	Location         string
	EmploymentType   models.EmploymentType
	WorkArrangement  models.WorkArrangement
	Category         models.JobCategory
	Seniority        string
	SalaryMin        *int
	SalaryMax        *int
	MustHaveSkills   []string
	NiceToHaveSkills []string
}

// Create inserts a new job as published immediately — JOBHOO's brief
// prioritizes fast, low-friction posting over a separate draft/review step.
func (r *JobsRepo) Create(ctx context.Context, in CreateJobInput) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO jobs (
			company_id, created_by, title, description, location, employment_type,
			work_arrangement, category, seniority, salary_min, salary_max,
			must_have_skills, nice_to_have_skills, status, published_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, 'published', now()
		)
		RETURNING id
	`, in.CompanyID, in.CreatedBy, in.Title, in.Description, in.Location, in.EmploymentType,
		in.WorkArrangement, in.Category, in.Seniority, in.SalaryMin, in.SalaryMax,
		in.MustHaveSkills, in.NiceToHaveSkills).Scan(&id)
	return id, err
}

func (r *JobsRepo) GetByID(ctx context.Context, id string) (*models.Job, error) {
	query := `
		SELECT ` + jobSelectColumns + `
		FROM jobs j
		JOIN companies c ON c.id = j.company_id
		WHERE j.id = $1
	`
	j, err := scanJob(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		return nil, err
	}
	return &j, nil
}

// rowScanner covers both pgx.Row (QueryRow) and pgx.Rows (Query), which
// share a Scan method but no common interface in pgx v5.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanJob(row rowScanner) (models.Job, error) {
	var j models.Job
	err := row.Scan(
		&j.ID, &j.Title, &j.Description, &j.Location, &j.EmploymentType, &j.WorkArrangement,
		&j.Category, &j.SalaryMin, &j.SalaryMax, &j.SalaryCurrency, &j.MustHaveSkills,
		&j.NiceToHaveSkills, &j.Seniority, &j.Status, &j.PublishedAt, &j.CreatedAt,
		&j.CompanyID, &j.CompanyName, &j.CompanyLogoURL,
	)
	return j, err
}
