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

// Upsert creates or updates a candidate's profile in one call, since the
// candidate/profile relationship is 1:1 and the form always submits the
// full profile.
func (r *CandidateProfilesRepo) Upsert(ctx context.Context, userID, headline, resumeText, location string, skills []string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO candidate_profiles (user_id, headline, resume_text, location, skills, updated_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (user_id) DO UPDATE SET
			headline = EXCLUDED.headline,
			resume_text = EXCLUDED.resume_text,
			location = EXCLUDED.location,
			skills = EXCLUDED.skills,
			updated_at = now()
	`, userID, headline, resumeText, location, skills)
	return err
}
