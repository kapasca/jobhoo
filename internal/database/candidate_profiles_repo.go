package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobhoo/jobhoo/internal/models"
)

type CandidateProfilesRepo struct {
	pool *pgxpool.Pool
}

func NewCandidateProfilesRepo(pool *pgxpool.Pool) *CandidateProfilesRepo {
	return &CandidateProfilesRepo{pool: pool}
}

func (r *CandidateProfilesRepo) GetByUserID(ctx context.Context, userID string) (*models.CandidateProfile, error) {
	var p models.CandidateProfile
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, coalesce(headline,''), coalesce(resume_text,''), coalesce(resume_file_url,''),
		       skills, coalesce(location,''), updated_at
		FROM candidate_profiles WHERE user_id = $1
	`, userID).Scan(&p.UserID, &p.Headline, &p.ResumeText, &p.ResumeFileURL, &p.Skills, &p.Location, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// Upsert creates or updates a candidate's profile. If resumeFileURL is empty
// the existing stored URL is preserved (old files are never deleted).
func (r *CandidateProfilesRepo) Upsert(ctx context.Context, userID, headline, resumeText, resumeFileURL, location string, skills []string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO candidate_profiles (user_id, headline, resume_text, resume_file_url, location, skills, updated_at)
		VALUES ($1, $2, $3, NULLIF($4,''), $5, $6, now())
		ON CONFLICT (user_id) DO UPDATE SET
			headline = EXCLUDED.headline,
			resume_text = EXCLUDED.resume_text,
			resume_file_url = CASE WHEN $4 = '' THEN candidate_profiles.resume_file_url ELSE NULLIF($4,'') END,
			location = EXCLUDED.location,
			skills = EXCLUDED.skills,
			updated_at = now()
	`, userID, headline, resumeText, resumeFileURL, location, skills)
	return err
}
