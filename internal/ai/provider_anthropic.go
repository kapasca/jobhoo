package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// AnthropicProvider implements Provider by calling the Anthropic Messages
// API (https://api.anthropic.com/v1/messages) and asking Claude to respond
// with strict JSON matching each task's result shape. It's the reference
// "real" implementation demonstrating how to plug a vendor into JOBHOO's AI
// layer — every method below follows the same pattern: build a task-specific
// prompt, call complete(), parse the JSON response into the shared result
// types from provider.go.
type AnthropicProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewAnthropicProvider(apiKey string) (*AnthropicProvider, error) {
	if apiKey == "" {
		return nil, errors.New("AI_API_KEY is required for the anthropic provider")
	}
	model := os.Getenv("AI_MODEL")
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}
	return &AnthropicProvider{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

// --- Provider methods --------------------------------------------------------

func (p *AnthropicProvider) RankCandidates(ctx context.Context, job JobContext, candidates []CandidateContext) ([]CandidateRanking, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	prompt := fmt.Sprintf(`Rank these candidates for the job below, most to least suitable.

Job: %s
Must-have skills: %s
Nice-to-have skills: %s

Candidates (id, name, skills):
%s

Respond with ONLY a JSON array, no prose, matching exactly:
[{"candidate_id": string, "score": number (0-100), "summary": string (one sentence)}]`,
		job.Title, strings.Join(job.MustHave, ", "), strings.Join(job.NiceToHave, ", "), formatCandidates(candidates))

	var result []struct {
		CandidateID string  `json:"candidate_id"`
		Score       float64 `json:"score"`
		Summary     string  `json:"summary"`
	}
	if err := p.completeJSON(ctx, rankingSystemPrompt, prompt, &result); err != nil {
		return nil, err
	}

	rankings := make([]CandidateRanking, 0, len(result))
	for _, r := range result {
		rankings = append(rankings, CandidateRanking{CandidateID: r.CandidateID, Score: r.Score, Summary: r.Summary})
	}
	return rankings, nil
}

func (p *AnthropicProvider) ExplainMatch(ctx context.Context, job JobContext, candidate CandidateContext) (MatchExplanation, error) {
	prompt := fmt.Sprintf(`Job: %s. Must-have: %s. Candidate skills: %s.

Respond with ONLY JSON matching exactly:
{"strengths": [string], "gaps": [string], "overall_note": string (one sentence)}`,
		job.Title, strings.Join(job.MustHave, ", "), strings.Join(candidate.Skills, ", "))

	var result struct {
		Strengths   []string `json:"strengths"`
		Gaps        []string `json:"gaps"`
		OverallNote string   `json:"overall_note"`
	}
	if err := p.completeJSON(ctx, matchSystemPrompt, prompt, &result); err != nil {
		return MatchExplanation{}, err
	}
	return MatchExplanation{Strengths: result.Strengths, Gaps: result.Gaps, OverallNote: result.OverallNote}, nil
}

func (p *AnthropicProvider) SummarizeResume(ctx context.Context, resumeText string) (ResumeSummary, error) {
	prompt := fmt.Sprintf(`Resume:
%s

Respond with ONLY JSON matching exactly:
{"headline": string, "key_skills": [string], "experience": string (one sentence), "highlights": [string]}`, resumeText)

	var result struct {
		Headline   string   `json:"headline"`
		KeySkills  []string `json:"key_skills"`
		Experience string   `json:"experience"`
		Highlights []string `json:"highlights"`
	}
	if err := p.completeJSON(ctx, summarySystemPrompt, prompt, &result); err != nil {
		return ResumeSummary{}, err
	}
	return ResumeSummary{
		Headline: result.Headline, KeySkills: result.KeySkills,
		Experience: result.Experience, Highlights: result.Highlights,
	}, nil
}

func (p *AnthropicProvider) RecommendJobs(ctx context.Context, candidate CandidateContext, jobs []JobContext) ([]JobRecommendation, error) {
	if len(jobs) == 0 {
		return nil, nil
	}
	prompt := fmt.Sprintf(`Candidate skills: %s.

Jobs (id, title, must-have skills):
%s

Respond with ONLY a JSON array, no prose, matching exactly:
[{"job_id": string, "score": number (0-100), "reason": string (one sentence)}]`,
		strings.Join(candidate.Skills, ", "), formatJobs(jobs))

	var result []struct {
		JobID  string  `json:"job_id"`
		Score  float64 `json:"score"`
		Reason string  `json:"reason"`
	}
	if err := p.completeJSON(ctx, recommendSystemPrompt, prompt, &result); err != nil {
		return nil, err
	}
	recs := make([]JobRecommendation, 0, len(result))
	for _, r := range result {
		recs = append(recs, JobRecommendation{JobID: r.JobID, Score: r.Score, Reason: r.Reason})
	}
	return recs, nil
}

func (p *AnthropicProvider) SuggestResumeImprovements(ctx context.Context, resumeText string, job *JobContext) ([]string, error) {
	targetLine := "in general"
	if job != nil {
		targetLine = "for the role: " + job.Title + " (must-have: " + strings.Join(job.MustHave, ", ") + ")"
	}
	prompt := fmt.Sprintf(`Resume:
%s

Suggest concrete improvements %s.

Respond with ONLY a JSON array of strings, no prose: ["suggestion 1", "suggestion 2", ...]`, resumeText, targetLine)

	var result []string
	if err := p.completeJSON(ctx, resumeAdviceSystemPrompt, prompt, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Anthropic API plumbing ---------------------------------------------------

const (
	rankingSystemPrompt      = "You are an impartial recruiting assistant. You score candidate fit objectively based only on stated skills. You never fabricate information not given to you. Output strict JSON only, no markdown fences, no commentary."
	matchSystemPrompt        = "You are an impartial recruiting assistant explaining candidate-job fit factually. Output strict JSON only, no markdown fences, no commentary."
	summarySystemPrompt      = "You summarize resumes factually and concisely for recruiter scanning. Output strict JSON only, no markdown fences, no commentary."
	recommendSystemPrompt    = "You are an impartial job-matching assistant. Output strict JSON only, no markdown fences, no commentary."
	resumeAdviceSystemPrompt = "You give concrete, actionable resume feedback. Output strict JSON only, no markdown fences, no commentary."
)

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// completeJSON sends a single-turn request to Claude and unmarshals the
// (expected-JSON) response text into target. AI output is advisory-only in
// JOBHOO (see provider.go); a malformed or failed response here simply
// surfaces as an error to the caller, which the handler treats as "AI
// ranking unavailable right now" rather than a hard failure of the page.
func (p *AnthropicProvider) completeJSON(ctx context.Context, systemPrompt, userPrompt string, target any) error {
	reqBody := anthropicRequest{
		Model:     p.model,
		MaxTokens: 2048,
		System:    systemPrompt,
		Messages:  []anthropicMessage{{Role: "user", Content: userPrompt}},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling anthropic api: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return fmt.Errorf("unmarshaling response: %w", err)
	}
	if parsed.Error != nil {
		return fmt.Errorf("anthropic api error: %s", parsed.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("anthropic api returned status %d: %s", resp.StatusCode, string(respBody))
	}
	if len(parsed.Content) == 0 {
		return errors.New("anthropic api returned no content")
	}

	text := strings.TrimSpace(parsed.Content[0].Text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	if err := json.Unmarshal([]byte(text), target); err != nil {
		return fmt.Errorf("parsing model output as JSON: %w (raw: %s)", err, text)
	}
	return nil
}

func formatCandidates(candidates []CandidateContext) string {
	var b strings.Builder
	for _, c := range candidates {
		fmt.Fprintf(&b, "- id=%s name=%s skills=%s\n", c.ID, c.Name, strings.Join(c.Skills, ", "))
	}
	return b.String()
}

func formatJobs(jobs []JobContext) string {
	var b strings.Builder
	for _, j := range jobs {
		fmt.Fprintf(&b, "- id=%s title=%s must_have=%s\n", j.ID, j.Title, strings.Join(j.MustHave, ", "))
	}
	return b.String()
}
