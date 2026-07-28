package ai

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// OpenAIProvider is a real-provider skeleton. It demonstrates that
// swapping the AI backend only requires implementing the Provider interface
// once, here — nothing in handlers, templates, or routing changes.
//
// TODO before production use: implement each method by calling
// https://api.openai.com/v1/chat/completions with a task-specific prompt built
// from the given context, then parse the structured JSON response into the
// corresponding result type. Keep prompts and parsing local to this file.
type OpenAIProvider struct {
	apiKey     string
	httpClient *http.Client
}

func NewOpenAIProvider(apiKey string) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, errors.New("AI_API_KEY is required for the openai provider")
	}
	return &OpenAIProvider{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) RankCandidates(ctx context.Context, job JobContext, candidates []CandidateContext) ([]CandidateRanking, error) {
	return nil, errors.New("openai provider not yet implemented: see provider_openai.go")
}

func (p *OpenAIProvider) ExplainMatch(ctx context.Context, job JobContext, candidate CandidateContext) (MatchExplanation, error) {
	return MatchExplanation{}, errors.New("openai provider not yet implemented: see provider_openai.go")
}

func (p *OpenAIProvider) SummarizeResume(ctx context.Context, resumeText string) (ResumeSummary, error) {
	return ResumeSummary{}, errors.New("openai provider not yet implemented: see provider_openai.go")
}

func (p *OpenAIProvider) RecommendJobs(ctx context.Context, candidate CandidateContext, jobs []JobContext) ([]JobRecommendation, error) {
	return nil, errors.New("openai provider not yet implemented: see provider_openai.go")
}

func (p *OpenAIProvider) SuggestResumeImprovements(ctx context.Context, resumeText string, job *JobContext) ([]string, error) {
	return nil, errors.New("openai provider not yet implemented: see provider_openai.go")
}
