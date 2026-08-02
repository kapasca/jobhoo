// Package ai defines JOBHOO's AI abstraction layer.
//
// Every AI-assisted feature in the product (candidate ranking, match
// explanations, resume summarization, job recommendations, gap analysis)
// talks ONLY to the Provider interface below. Handlers and services never
// import a specific vendor SDK directly. This means the underlying model or
// vendor can be replaced — or multiple providers can be run side by side for
// different tasks — by adding a new implementation of Provider and flipping
// the AI_PROVIDER config value, with zero changes to handlers, templates, or
// business logic.
//
// AI output is always advisory. Nothing in this package makes a hiring
// decision; it only produces information a human reviews.
package ai

import "context"

// Provider is the single seam between JOBHOO and any AI backend.
// Concrete implementations live under internal/ai/providers/.
type Provider interface {
	// RankCandidates scores and orders candidates against a job, returning
	// one result per candidate. It never removes or hides a candidate —
	// ranking is a suggestion for recruiter review, not a filter.
	RankCandidates(ctx context.Context, job JobContext, candidates []CandidateContext) ([]CandidateRanking, error)

	// ExplainMatch produces a short, human-readable explanation of why a
	// specific candidate does or doesn't fit a specific job.
	ExplainMatch(ctx context.Context, job JobContext, candidate CandidateContext) (MatchExplanation, error)

	// RecommendJobs suggests jobs a candidate is likely to be a strong fit
	// for, with a reason for each recommendation.
	RecommendJobs(ctx context.Context, candidate CandidateContext, jobs []JobContext) ([]JobRecommendation, error)

	// ExtractResumeText parses raw resume text (from a TXT upload, or text
	// pulled out of a DOCX file) into structured, factual fields. This is
	// the primary source of truth for candidate profiling — see
	// ResumeExtraction.
	ExtractResumeText(ctx context.Context, rawText string) (ResumeExtraction, error)

	// ExtractResumeFile parses a resume file directly via vision (PDF or
	// image) into the same structured fields as ExtractResumeText.
	ExtractResumeFile(ctx context.Context, fileData []byte, mediaType string) (ResumeExtraction, error)

	// Name identifies the provider for logging/diagnostics (e.g. "openai",
	// "anthropic", "mock").
	Name() string
}

// --- Shared context/result types -------------------------------------------------

type JobContext struct {
	ID          string
	Title       string
	Description string
	MustHave    []string
	NiceToHave  []string
	Seniority   string
}

// CandidateContext carries everything an AI call needs to know about one
// candidate. ResumeText should always be built from the candidate's parsed
// resume (see ResumeExtraction / FormatResumeExtractionText) — it is the
// PRIMARY source of truth. Skills is the candidate's separately entered
// profile field; every prompt in this package is written to treat it as a
// SECONDARY, lower-weight cross-check, never a replacement for the resume.
type CandidateContext struct {
	ID         string
	Name       string
	ResumeText string
	Skills     []string
	Experience []string
}

type CandidateRanking struct {
	CandidateID string
	Score       float64 // 0-100, relative fit for this job
	Summary     string  // one-line rationale
}

type MatchExplanation struct {
	Strengths   []string
	Gaps        []string
	OverallNote string
}

type JobRecommendation struct {
	JobID  string
	Score  float64
	Reason string
}
