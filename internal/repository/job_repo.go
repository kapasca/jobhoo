package repository

import (
	"database/sql"
	"time"

	"jobhoo/internal/models"
)

type JobRepo struct {
	db *sql.DB
}

func NewJobRepo(db *sql.DB) *JobRepo {
	return &JobRepo{db: db}
}

type JobInput struct {
	Title           string
	Position        string
	EmploymentType  models.EmploymentType
	WorkArrangement models.WorkArrangement
	Location        string
	SalaryMin       *int64
	SalaryMax       *int64
	Benefits        string
	Requirements    string
	Description     string
	ClosingDate     time.Time
}

func (r *JobRepo) Create(recruiterID int64, in JobInput) (int64, error) {
	var id int64
	err := r.db.QueryRow(
		`INSERT INTO jobs (recruiter_id, title, position, employment_type, work_arrangement, location,
		                    salary_min, salary_max, benefits, requirements, description, closing_date, status)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,'open') RETURNING id`,
		recruiterID, in.Title, in.Position, in.EmploymentType, in.WorkArrangement, in.Location,
		in.SalaryMin, in.SalaryMax, in.Benefits, in.Requirements, in.Description, in.ClosingDate,
	).Scan(&id)
	return id, err
}

func (r *JobRepo) Update(id int64, in JobInput) error {
	_, err := r.db.Exec(
		`UPDATE jobs SET title=$1, position=$2, employment_type=$3, work_arrangement=$4, location=$5,
		                  salary_min=$6, salary_max=$7, benefits=$8, requirements=$9, description=$10,
		                  closing_date=$11, updated_at=now()
		 WHERE id=$12`,
		in.Title, in.Position, in.EmploymentType, in.WorkArrangement, in.Location,
		in.SalaryMin, in.SalaryMax, in.Benefits, in.Requirements, in.Description, in.ClosingDate, id,
	)
	return err
}

func (r *JobRepo) Close(id int64) error {
	_, err := r.db.Exec(`UPDATE jobs SET status = 'closed', updated_at = now() WHERE id = $1`, id)
	return err
}

func (r *JobRepo) GetByID(id int64) (*models.Job, error) {
	j := &models.Job{}
	err := r.db.QueryRow(
		`SELECT id, recruiter_id, title, position, employment_type, work_arrangement, location,
		        salary_min, salary_max, benefits, requirements, description, closing_date, status, created_at, updated_at
		 FROM jobs WHERE id = $1`, id,
	).Scan(&j.ID, &j.RecruiterID, &j.Title, &j.Position, &j.EmploymentType, &j.WorkArrangement, &j.Location,
		&j.SalaryMin, &j.SalaryMax, &j.Benefits, &j.Requirements, &j.Description, &j.ClosingDate, &j.Status, &j.CreatedAt, &j.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return j, err
}

// ListOpen returns all open jobs, most recent first, for the public listing page.
func (r *JobRepo) ListOpen() ([]models.Job, error) {
	return r.ListOpenPaginated(1000, 0)
}

func (r *JobRepo) ListOpenPaginated(limit, offset int) ([]models.Job, error) {
	rows, err := r.db.Query(
		`SELECT id, recruiter_id, title, position, employment_type, work_arrangement, location,
		        salary_min, salary_max, benefits, requirements, description, closing_date, status, created_at, updated_at
		 FROM jobs WHERE status = 'open' ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobs(rows)
}

func (r *JobRepo) ListByRecruiter(recruiterID int64) ([]models.Job, error) {
	return r.ListByRecruiterPaginated(recruiterID, 1000, 0)
}

func (r *JobRepo) ListByRecruiterPaginated(recruiterID int64, limit, offset int) ([]models.Job, error) {
	rows, err := r.db.Query(
		`SELECT j.id, j.recruiter_id, j.title, j.position, j.employment_type, j.work_arrangement, j.location,
		        j.salary_min, j.salary_max, j.benefits, j.requirements, j.description, j.closing_date, j.status, j.created_at, j.updated_at
		 FROM jobs j WHERE j.recruiter_id = $1 ORDER BY j.created_at DESC
		 LIMIT $2 OFFSET $3`, recruiterID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobs(rows)
}

func (r *JobRepo) ListAll() ([]models.Job, error) {
	return r.ListAllPaginated(1000, 0)
}

func (r *JobRepo) ListAllPaginated(limit, offset int) ([]models.Job, error) {
	rows, err := r.db.Query(
		`SELECT id, recruiter_id, title, position, employment_type, work_arrangement, location,
		        salary_min, salary_max, benefits, requirements, description, closing_date, status, created_at, updated_at
		 FROM jobs ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobs(rows)
}

func (r *JobRepo) CountOpen() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM jobs WHERE status = 'open'`).Scan(&n)
	return n, err
}

func (r *JobRepo) CountByRecruiterAndStatus(recruiterID int64, status models.JobStatus) (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM jobs WHERE recruiter_id = $1 AND status = $2`, recruiterID, status).Scan(&n)
	return n, err
}

func (r *JobRepo) CountByRecruiter(recruiterID int64) (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM jobs WHERE recruiter_id = $1`, recruiterID).Scan(&n)
	return n, err
}

func (r *JobRepo) CountAll() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM jobs`).Scan(&n)
	return n, err
}

func scanJobs(rows *sql.Rows) ([]models.Job, error) {
	var out []models.Job
	for rows.Next() {
		var j models.Job
		if err := rows.Scan(&j.ID, &j.RecruiterID, &j.Title, &j.Position, &j.EmploymentType, &j.WorkArrangement, &j.Location,
			&j.SalaryMin, &j.SalaryMax, &j.Benefits, &j.Requirements, &j.Description, &j.ClosingDate, &j.Status, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}
