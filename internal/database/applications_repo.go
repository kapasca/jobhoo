package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobhoo/jobhoo/internal/models"
)

var ErrAlreadyApplied = errors.New("candidate already applied to this job")

const pgUniqueViolationCode = "23505"

type ApplicationsRepo struct {
	pool *pgxpool.Pool
}

func NewApplicationsRepo(pool *pgxpool.Pool) *ApplicationsRepo {
	return &ApplicationsRepo{pool: pool}
}

func (r *ApplicationsRepo) Create(ctx context.Context, jobID, candidateID, coverNote string) (*models.Application, error) {
	var a models.Application
	err := r.pool.QueryRow(ctx, `
		INSERT INTO applications (job_id, candidate_id, cover_note)
		VALUES ($1, $2, $3)
		RETURNING id, job_id, candidate_id, stage, coalesce(cover_note, ''), created_at, updated_at
	`, jobID, candidateID, coverNote).Scan(&a.ID, &a.JobID, &a.CandidateID, &a.Stage, &a.CoverNote, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
			return nil, ErrAlreadyApplied
		}
		return nil, err
	}
	return &a, nil
}

// ListByCandidate returns every application a candidate has made, most
// recent first, joined with job/company info for display.
func (r *ApplicationsRepo) ListByCandidate(ctx context.Context, candidateID string) ([]models.Application, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id, a.job_id, a.candidate_id, a.stage, coalesce(a.cover_note,''), a.created_at, a.updated_at,
		       j.title, c.name, j.location
		FROM applications a
		JOIN jobs j ON j.id = a.job_id
		JOIN companies c ON c.id = j.company_id
		WHERE a.candidate_id = $1
		ORDER BY a.created_at DESC
	`, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []models.Application
	for rows.Next() {
		var a models.Application
		if err := rows.Scan(&a.ID, &a.JobID, &a.CandidateID, &a.Stage, &a.CoverNote, &a.CreatedAt, &a.UpdatedAt,
			&a.JobTitle, &a.CompanyName, &a.Location); err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, rows.Err()
}

// ListByJob returns every application for a job, joined with candidate info
// and their profile skills. Callers group the flat list by Stage
// client-side to render the ATS board columns.
func (r *ApplicationsRepo) ListByJob(ctx context.Context, jobID string) ([]models.Application, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id, a.job_id, a.candidate_id, a.stage, coalesce(a.cover_note,''), a.created_at, a.updated_at,
		       u.full_name, u.email, coalesce(cp.skills, '{}')
		FROM applications a
		JOIN users u ON u.id = a.candidate_id
		LEFT JOIN candidate_profiles cp ON cp.user_id = a.candidate_id
		WHERE a.job_id = $1
		ORDER BY a.created_at ASC
	`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []models.Application
	for rows.Next() {
		var a models.Application
		if err := rows.Scan(&a.ID, &a.JobID, &a.CandidateID, &a.Stage, &a.CoverNote, &a.CreatedAt, &a.UpdatedAt,
			&a.CandidateName, &a.CandidateEmail, &a.CandidateSkills); err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, rows.Err()
}

func (r *ApplicationsRepo) GetByID(ctx context.Context, id string) (*models.Application, error) {
	var a models.Application
	err := r.pool.QueryRow(ctx, `
		SELECT id, job_id, candidate_id, stage, coalesce(cover_note,''), created_at, updated_at
		FROM applications WHERE id = $1
	`, id).Scan(&a.ID, &a.JobID, &a.CandidateID, &a.Stage, &a.CoverNote, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *ApplicationsRepo) UpdateStage(ctx context.Context, id string, stage models.ApplicationStage) error {
	_, err := r.pool.Exec(ctx, `UPDATE applications SET stage = $1, updated_at = now() WHERE id = $2`, stage, id)
	return err
}

// HasApplied reports whether a candidate has already applied to a job, used
// to toggle the apply button/form on the job detail page.
func (r *ApplicationsRepo) HasApplied(ctx context.Context, jobID, candidateID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM applications WHERE job_id = $1 AND candidate_id = $2)
	`, jobID, candidateID).Scan(&exists)
	return exists, err
}

// CountByJob returns the number of applications for a job, used for the
// recruiter dashboard's job list ("12 applicants").
func (r *ApplicationsRepo) CountByJob(ctx context.Context, jobID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM applications WHERE job_id = $1`, jobID).Scan(&count)
	return count, err
}
