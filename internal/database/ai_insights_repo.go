package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobhoo/jobhoo/internal/models"
)

// AIInsightsRepo persists AI ranking scores and match explanations against
// the (job, candidate) pair they were computed for. This exists so results
// survive page reloads and don't need to be recomputed — and re-billed on
// the AI gateway — every time a recruiter opens the ATS board or a
// candidate's detail card.
type AIInsightsRepo struct {
	pool *pgxpool.Pool
}

func NewAIInsightsRepo(pool *pgxpool.Pool) *AIInsightsRepo {
	return &AIInsightsRepo{pool: pool}
}

func scanAIMatchInsight(row pgx.Row) (*models.AIMatchInsight, error) {
	var m models.AIMatchInsight
	var score *float64
	err := row.Scan(&m.JobID, &m.CandidateID, &m.Provider, &score, &m.Summary,
		&m.Strengths, &m.Gaps, &m.OverallNote, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	if score != nil {
		m.Score = *score
	}
	return &m, nil
}

// Get returns the cached insight for a (job, candidate) pair, or ErrNotFound
// if no AI call has been made for it yet.
func (r *AIInsightsRepo) Get(ctx context.Context, jobID, candidateID string) (*models.AIMatchInsight, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT job_id, candidate_id, provider, score, coalesce(summary,''),
		       strengths, gaps, coalesce(overall_note,''), created_at
		FROM ai_match_insights WHERE job_id = $1 AND candidate_id = $2
	`, jobID, candidateID)
	m, err := scanAIMatchInsight(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m, nil
}

// ListByJob returns every cached insight for a job, keyed by candidate_id,
// so the ATS board can show previously-computed scores without an AI call.
func (r *AIInsightsRepo) ListByJob(ctx context.Context, jobID string) (map[string]models.AIMatchInsight, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT job_id, candidate_id, provider, score, coalesce(summary,''),
		       strengths, gaps, coalesce(overall_note,''), created_at
		FROM ai_match_insights WHERE job_id = $1
	`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]models.AIMatchInsight)
	for rows.Next() {
		m, err := scanAIMatchInsight(rows)
		if err != nil {
			return nil, err
		}
		out[m.CandidateID] = *m
	}
	return out, rows.Err()
}

// UpsertRanking stores/updates the score+summary produced by RankCandidates,
// preserving any existing strengths/gaps/overall_note from a prior
// ExplainMatch call for the same pair.
func (r *AIInsightsRepo) UpsertRanking(ctx context.Context, jobID, candidateID, provider string, score float64, summary string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ai_match_insights (job_id, candidate_id, provider, score, summary)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (job_id, candidate_id) DO UPDATE
		SET provider = EXCLUDED.provider, score = EXCLUDED.score, summary = EXCLUDED.summary
	`, jobID, candidateID, provider, score, summary)
	return err
}

// UpsertExplanation stores/updates the strengths/gaps/overall_note produced
// by ExplainMatch, preserving any existing score/summary from a prior
// RankCandidates call for the same pair.
func (r *AIInsightsRepo) UpsertExplanation(ctx context.Context, jobID, candidateID, provider string, strengths, gaps []string, overallNote string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ai_match_insights (job_id, candidate_id, provider, strengths, gaps, overall_note)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (job_id, candidate_id) DO UPDATE
		SET provider = EXCLUDED.provider, strengths = EXCLUDED.strengths,
		    gaps = EXCLUDED.gaps, overall_note = EXCLUDED.overall_note
	`, jobID, candidateID, provider, strengths, gaps, overallNote)
	return err
}
