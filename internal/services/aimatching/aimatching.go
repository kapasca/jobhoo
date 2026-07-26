// Package aimatching isolates all resume-matching logic behind a single
// interface. Today it's backed by a deterministic mock (mockProvider) so the
// whole application works end-to-end without any external AI dependency.
//
// To integrate a real provider (OpenAI, Gemini, Claude, or a local model)
// later, implement the Provider interface below and swap it in at
// NewProvider() — nothing else in the codebase needs to change.
package aimatching

import (
	"hash/fnv"
)

// Result is the outcome of matching a candidate's resume against a job.
type Result struct {
	MatchScore      int // 0-100 overall score
	SkillMatch      int // 0-100
	ExperienceMatch int // 0-100
	EducationMatch  int // 0-100
}

// Provider is the contract any AI matching backend must satisfy.
type Provider interface {
	Analyze(resumePath string, jobDescription string, jobRequirements string) (Result, error)
}

// NewProvider returns the currently active matching provider.
// Swap this out (e.g. based on an env var) to point to a real AI backend.
func NewProvider() Provider {
	return &mockProvider{}
}

// mockProvider produces deterministic, plausible-looking scores derived from
// a hash of the inputs, so the same resume+job pair always yields the same
// result (useful for demos and tests) without calling any external API.
type mockProvider struct{}

func (m *mockProvider) Analyze(resumePath, jobDescription, jobRequirements string) (Result, error) {
	seed := hashString(resumePath + "|" + jobDescription + "|" + jobRequirements)

	skill := 55 + int(seed%40)           // 55-94
	experience := 50 + int((seed/7)%45)  // 50-94
	education := 60 + int((seed/13)%35)  // 60-94
	overall := (skill + experience + education) / 3

	return Result{
		MatchScore:      clamp(overall),
		SkillMatch:      clamp(skill),
		ExperienceMatch: clamp(experience),
		EducationMatch:  clamp(education),
	}, nil
}

func hashString(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}

func clamp(v int) int {
	if v > 100 {
		return 100
	}
	if v < 0 {
		return 0
	}
	return v
}
