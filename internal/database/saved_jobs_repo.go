package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobhoo/jobhoo/internal/models"
)

type SavedJobsRepo struct {
	pool *pgxpool.Pool
}

func NewSavedJobsRepo(pool *pgxpool.Pool) *SavedJobsRepo {
	return &SavedJobsRepo{pool: pool}
}

// GetSavedJobIDs returns a set of job IDs saved by the candidate — used to
// mark bookmark icons on job list pages without N extra queries.
func (r *SavedJobsRepo) GetSavedJobIDs(ctx context.Context, candidateID string) (map[string]bool, error) {
	rows, err := r.pool.Query(ctx, `SELECT job_id FROM saved_jobs WHERE candidate_id = $1`, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

func (r *SavedJobsRepo) IsSaved(ctx context.Context, candidateID, jobID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM saved_jobs WHERE candidate_id = $1 AND job_id = $2)
	`, candidateID, jobID).Scan(&exists)
	return exists, err
}

// Toggle saves the job if not already saved, or un-saves it if it is,
// returning the resulting saved state.
func (r *SavedJobsRepo) Toggle(ctx context.Context, candidateID, jobID string) (bool, error) {
	saved, err := r.IsSaved(ctx, candidateID, jobID)
	if err != nil {
		return false, err
	}
	if saved {
		_, err := r.pool.Exec(ctx, `DELETE FROM saved_jobs WHERE candidate_id = $1 AND job_id = $2`, candidateID, jobID)
		return false, err
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO saved_jobs (candidate_id, job_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, candidateID, jobID)
	return true, err
}

// ListByCandidate returns every job a candidate has saved, most recently
// saved first, joined with company info for card rendering.
func (r *SavedJobsRepo) ListByCandidate(ctx context.Context, candidateID string) ([]models.Job, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+jobSelectColumns+`
		FROM saved_jobs sj
		JOIN jobs j ON j.id = sj.job_id
		JOIN companies c ON c.id = j.company_id
		WHERE sj.candidate_id = $1
		ORDER BY sj.created_at DESC
	`, candidateID)
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
