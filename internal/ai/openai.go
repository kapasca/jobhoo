package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// OpenAIProvider implements Provider by calling OpenAI's Chat Completions API
// or any compatible gateway (via AI_BASE_URL env var). It's a fully-featured implementation
// supporting GPT-4o with vision capabilities for parsing resume PDFs, making it ideal
// for JOBHOO's AI features with low cost and high reliability.
//
// Key features:
// - Vision support: can read and parse resume PDFs directly
// - Reasoning: access to o1-mini for complex candidate matching
// - JSON mode: ensures strict structured JSON responses
// - Cost-effective: cheaper than Anthropic for most use cases
// - Gateway support: works with any OpenAI-compatible API gateway
//
// Every method follows the same pattern: build a task-specific prompt (with vision
// support where applicable), call complete() or completeWithVision(), parse the JSON
// response into the shared result types from provider.go.
type OpenAIProvider struct {
	apiKey      string
	model       string
	visionModel string
	baseURL     string
	httpClient  *http.Client
	logger      CallLogger
}

func NewOpenAIProvider(apiKey string, logger CallLogger) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, errors.New("AI_API_KEY is required for the openai provider")
	}
	model := os.Getenv("AI_MODEL")
	if model == "" {
		model = "openai/gpt-5-nano"
	}
	// Vision model: use the same model as default. This is configurable per deployment.
	// If AI_VISION_MODEL env var is set, use that; otherwise use the configured model.
	visionModel := os.Getenv("AI_VISION_MODEL")
	if visionModel == "" {
		visionModel = model
	}
	// Base URL for API calls — supports custom gateways via AI_BASE_URL env var
	baseURL := os.Getenv("AI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	// Ensure baseURL doesn't have trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &OpenAIProvider{
		apiKey:      apiKey,
		model:       model,
		visionModel: visionModel,
		baseURL:     baseURL,
		httpClient:  &http.Client{Timeout: 60 * time.Second}, // Increased timeout for vision processing
		logger:      logger,
	}, nil
}

func (p *OpenAIProvider) Name() string { return "openai" }

// --- Provider methods --------------------------------------------------------

func (p *OpenAIProvider) RankCandidates(ctx context.Context, job JobContext, candidates []CandidateContext) ([]CandidateRanking, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	prompt := fmt.Sprintf(`Rank these candidates for the job below, most to least suitable.

Job: %s
Must-have skills: %s
Nice-to-have skills: %s
Seniority level: %s

Candidates (id, name, resume summary):
%s

Respond with ONLY a JSON object, no prose, matching exactly:
{"rankings": [{"candidate_id": string, "score": number (0-100), "summary": string (one sentence)}]}`,
		job.Title, strings.Join(job.MustHave, ", "), strings.Join(job.NiceToHave, ", "),
		job.Seniority, formatCandidatesWithResume(candidates))

	var result struct {
		Rankings []struct {
			CandidateID string  `json:"candidate_id"`
			Score       float64 `json:"score"`
			Summary     string  `json:"summary"`
		} `json:"rankings"`
	}
	if err := p.completeJSON(ctx, rankingSystemPrompt, prompt, &result); err != nil {
		return nil, err
	}

	rankings := make([]CandidateRanking, 0, len(result.Rankings))
	for _, r := range result.Rankings {
		rankings = append(rankings, CandidateRanking{CandidateID: r.CandidateID, Score: r.Score, Summary: r.Summary})
	}
	return rankings, nil
}

func (p *OpenAIProvider) ExplainMatch(ctx context.Context, job JobContext, candidate CandidateContext) (MatchExplanation, error) {
	prompt := fmt.Sprintf(`Analyze the match between candidate and job position.

Job: %s
Must-have skills: %s
Nice-to-have skills: %s
Seniority: %s

Candidate:
Name: %s
Resume: %s

Respond with ONLY JSON matching exactly:
{"strengths": [string], "gaps": [string], "overall_note": string (one sentence max)}`,
		job.Title, strings.Join(job.MustHave, ", "), strings.Join(job.NiceToHave, ", "),
		job.Seniority, candidate.Name, candidate.ResumeText)

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

func (p *OpenAIProvider) ExtractResumeText(ctx context.Context, rawText string) (ResumeExtraction, error) {
	prompt := fmt.Sprintf(`Extract structured, factual information from this resume. Do not invent anything not present in the text.

Resume:
%s

Respond with ONLY JSON matching exactly:
{"headline": string, "experience": [string], "education": [string], "skills": [string], "achievements": [string], "candidate_story": string}

- "experience": one entry per role/position, e.g. "Senior Engineer at Acme Corp (2020-2023): led backend team".
- "education": one entry per degree/certification.
- "skills": flat list of skills explicitly mentioned in the resume.
- "achievements": notable accomplishments, awards, or measurable results.
- "candidate_story": a short 2-3 sentence factual narrative summarizing who this candidate is professionally, based only on the resume.`, rawText)

	var result ResumeExtraction
	if err := p.completeJSON(ctx, resumeExtractionSystemPrompt, prompt, &result); err != nil {
		return ResumeExtraction{}, err
	}
	return result, nil
}

func (p *OpenAIProvider) RecommendJobs(ctx context.Context, candidate CandidateContext, jobs []JobContext) ([]JobRecommendation, error) {
	if len(jobs) == 0 {
		return nil, nil
	}
	prompt := fmt.Sprintf(`Match candidate to suitable jobs based on skills and experience.

Candidate:
Name: %s
Skills: %s
Experience: %s
Resume: %s

Available Jobs (id, title, must-have skills, seniority):
%s

Respond with ONLY a JSON object, no prose, matching exactly:
{"recommendations": [{"job_id": string, "score": number (0-100), "reason": string (one sentence)}]}`,
		candidate.Name, strings.Join(candidate.Skills, ", "),
		strings.Join(candidate.Experience, "; "),
		candidate.ResumeText, formatJobsDetail(jobs))

	var result struct {
		Recommendations []struct {
			JobID  string  `json:"job_id"`
			Score  float64 `json:"score"`
			Reason string  `json:"reason"`
		} `json:"recommendations"`
	}
	if err := p.completeJSON(ctx, recommendSystemPrompt, prompt, &result); err != nil {
		return nil, err
	}
	recs := make([]JobRecommendation, 0, len(result.Recommendations))
	for _, r := range result.Recommendations {
		recs = append(recs, JobRecommendation{JobID: r.JobID, Score: r.Score, Reason: r.Reason})
	}
	return recs, nil
}

// ExtractResumeFile parses a resume file (PDF or image) directly via vision
// into the same structured fields as ExtractResumeText. This is the primary
// entry point for turning an uploaded resume into candidate_profiles data.
func (p *OpenAIProvider) ExtractResumeFile(ctx context.Context, fileData []byte, mediaType string) (ResumeExtraction, error) {
	base64Data := base64.StdEncoding.EncodeToString(fileData)

	prompt := `Extract structured, factual information from this resume document. Do not invent anything not present in it.

Respond with ONLY JSON matching exactly:
{"headline": string, "experience": [string], "education": [string], "skills": [string], "achievements": [string], "candidate_story": string}

- "experience": one entry per role/position, e.g. "Senior Engineer at Acme Corp (2020-2023): led backend team".
- "education": one entry per degree/certification.
- "skills": flat list of skills explicitly mentioned in the resume.
- "achievements": notable accomplishments, awards, or measurable results.
- "candidate_story": a short 2-3 sentence factual narrative summarizing who this candidate is professionally, based only on the resume.`

	var result ResumeExtraction
	if err := p.completeWithVision(ctx, resumeExtractionSystemPrompt, prompt, base64Data, mediaType, &result); err != nil {
		return ResumeExtraction{}, err
	}
	return result, nil
}

// --- OpenAI API plumbing ---------------------------------------------------
// System prompts are defined in prompts.go and shared across all providers.

type openaiRequest struct {
	Model               string          `json:"model"`
	Messages            []openaiMessage `json:"messages"`
	Temperature         *float64        `json:"temperature,omitempty"`
	TopP                *float64        `json:"top_p,omitempty"`
	MaxTokens           *int            `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int            `json:"max_completion_tokens,omitempty"`
	ReasoningEffort     string          `json:"reasoning_effort,omitempty"`
	ResponseFormat      *responseFormat `json:"response_format,omitempty"`
}

// isReasoningModel reports whether model is a reasoning-family model
// (gpt-5, o1, o3, o4...). These models use max_completion_tokens instead of
// max_tokens, spend part of that budget on invisible reasoning tokens before
// producing visible output, and reject non-default temperature/top_p.
func isReasoningModel(model string) bool {
	m := model
	if idx := strings.LastIndex(m, "/"); idx != -1 {
		m = m[idx+1:]
	}
	return strings.HasPrefix(m, "gpt-5") || strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") || strings.HasPrefix(m, "o4")
}

// newCompletionRequest builds the model-specific parts of a chat completion
// request, since reasoning models and classic gpt-4-class models disagree on
// which sampling/token-budget params are supported.
func newCompletionRequest(model string, messages []openaiMessage, format *responseFormat) openaiRequest {
	return newCompletionRequestWithBudget(model, messages, format, 8192)
}

// newCompletionRequestWithBudget is like newCompletionRequest but lets the
// caller raise the reasoning-model token budget for tasks that need more
// headroom (e.g. vision/PDF parsing, which OpenAI's own docs recommend
// reserving ~25k tokens for). reasoning_effort=low keeps invisible reasoning
// tokens from eating the whole budget on straightforward extraction/ranking
// tasks that don't need deep multi-step reasoning.
func newCompletionRequestWithBudget(model string, messages []openaiMessage, format *responseFormat, reasoningTokenBudget int) openaiRequest {
	req := openaiRequest{Model: model, Messages: messages, ResponseFormat: format}
	if isReasoningModel(model) {
		tokens := reasoningTokenBudget
		req.MaxCompletionTokens = &tokens
		req.ReasoningEffort = "low"
	} else {
		temp, topP, tokens := 0.7, 0.9, 2048
		req.Temperature = &temp
		req.TopP = &topP
		req.MaxTokens = &tokens
	}
	return req
}

type responseFormat struct {
	Type string `json:"type"`
}

type openaiMessage struct {
	Role    string         `json:"role"`
	Content []contentBlock `json:"content"`
}

type contentBlock interface{}

type textContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type imageContent struct {
	Type     string       `json:"type"`
	ImageURL imageURLData `json:"image_url"`
}

type imageURLData struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// fileContent is the Chat Completions content part for document inputs
// (PDFs and other non-image files). Unlike image_url, this lets
// vision-capable models like gpt-4o and gpt-5 read PDFs directly — the API
// extracts both text and page images server-side.
type fileContent struct {
	Type string   `json:"type"`
	File fileData `json:"file"`
}

type fileData struct {
	Filename string `json:"filename"`
	FileData string `json:"file_data"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// detectMethodName maps a system prompt back to the Provider method that
// issued it, purely for AI call logging/diagnostics (token usage per
// feature, so cost on the gateway can be attributed).
func detectMethodName(systemPrompt string) string {
	switch systemPrompt {
	case rankingSystemPrompt:
		return "RankCandidates"
	case matchSystemPrompt:
		return "ExplainMatch"
	case recommendSystemPrompt:
		return "RecommendJobs"
	case resumeExtractionSystemPrompt:
		return "ExtractResume"
	default:
		return "unknown"
	}
}

// completeJSON sends a text-only request to OpenAI Chat Completions API and
// unmarshals the JSON response into target. Uses the standard model
// configured via AI_MODEL.
func (p *OpenAIProvider) completeJSON(ctx context.Context, systemPrompt, userPrompt string, target any) error {
	messages := []openaiMessage{
		{Role: "system", Content: []contentBlock{textContent{Type: "text", Text: systemPrompt}}},
		{Role: "user", Content: []contentBlock{textContent{Type: "text", Text: userPrompt}}},
	}
	reqBody := newCompletionRequest(p.model, messages, &responseFormat{Type: "json_object"})
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}
	return p.doRequestLogged(ctx, systemPrompt, userPrompt, body, target)
}

// completeWithVision sends a request with vision capability. PDFs are sent
// as a "file" content part (Chat Completions extracts text + page images
// server-side); other media types (actual images) use "image_url" as before.
// Uses the model configured via AI_VISION_MODEL.
func (p *OpenAIProvider) completeWithVision(ctx context.Context, systemPrompt, userPrompt, base64Data, mediaType string, target any) error {
	var attachment contentBlock
	if mediaType == "application/pdf" {
		attachment = fileContent{Type: "file", File: fileData{
			Filename: "resume.pdf",
			FileData: fmt.Sprintf("data:%s;base64,%s", mediaType, base64Data),
		}}
	} else {
		attachment = imageContent{Type: "image_url", ImageURL: imageURLData{
			URL:    fmt.Sprintf("data:%s;base64,%s", mediaType, base64Data),
			Detail: "high",
		}}
	}
	messages := []openaiMessage{
		{Role: "system", Content: []contentBlock{textContent{Type: "text", Text: systemPrompt}}},
		{Role: "user", Content: []contentBlock{
			textContent{Type: "text", Text: userPrompt},
			attachment,
		}},
	}
	// PDF/vision parsing burns far more reasoning tokens than plain text
	// tasks (page images + document understanding), so it gets a much
	// bigger budget to avoid finish_reason=length with empty output.
	reqBody := newCompletionRequestWithBudget(p.visionModel, messages, &responseFormat{Type: "json_object"}, 24000)
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshaling vision request: %w", err)
	}
	return p.doRequestLogged(ctx, systemPrompt, userPrompt+" [+file attachment]", body, target)
}

// doRequestLogged wraps doRequest with AI call logging (prompt, raw
// response, duration, status, token usage, and who triggered it) so the
// admin Activity page can show exactly what was sent to the AI gateway and
// how it responded.
func (p *OpenAIProvider) doRequestLogged(ctx context.Context, systemPrompt, userPrompt string, body []byte, target any) error {
	startTime := time.Now()
	tokensUsed, rawResponse, err := p.doRequest(ctx, body, target)
	logEntry := AICallLog{
		Timestamp:   startTime,
		Provider:    "openai",
		Method:      detectMethodName(systemPrompt),
		Prompt:      userPrompt,
		Response:    rawResponse,
		ExecutionMs: int(time.Since(startTime).Milliseconds()),
		TokensUsed:  tokensUsed,
	}
	if user, ok := userFromContext(ctx); ok {
		logEntry.UserID, logEntry.UserName, logEntry.UserEmail, logEntry.UserRole = user.ID, user.Name, user.Email, user.Role
	}
	if err != nil {
		logEntry.Status = "error"
		logEntry.Error = err.Error()
	} else {
		logEntry.Status = "success"
	}
	if p.logger != nil {
		p.logger.Log(ctx, logEntry)
	}
	return err
}

// doRequest performs the actual HTTP request to OpenAI API, returning the
// number of tokens the call consumed and the model's raw text response (for
// logging) alongside any error.
func (p *OpenAIProvider) doRequest(ctx context.Context, body []byte, target any) (tokensUsed int, rawResponse string, err error) {
	endpoint := fmt.Sprintf("%s/chat/completions", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, "", fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("calling openai api: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("reading response: %w", err)
	}

	var parsed openaiResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return 0, "", fmt.Errorf("unmarshaling response: %w", err)
	}
	if parsed.Usage != nil {
		tokensUsed = parsed.Usage.TotalTokens
	}
	if len(parsed.Choices) > 0 {
		rawResponse = parsed.Choices[0].Message.Content
	}

	if parsed.Error != nil {
		return tokensUsed, rawResponse, fmt.Errorf("openai api error (%s): %s", parsed.Error.Type, parsed.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return tokensUsed, rawResponse, fmt.Errorf("openai api returned status %d: %s", resp.StatusCode, string(respBody))
	}
	if len(parsed.Choices) == 0 {
		return tokensUsed, rawResponse, errors.New("openai api returned no choices")
	}

	text := strings.TrimSpace(parsed.Choices[0].Message.Content)
	// Remove markdown code fences if present
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	if text == "" {
		return tokensUsed, rawResponse, fmt.Errorf("openai api returned empty content (finish_reason=%s) — the model likely ran out of tokens before producing visible output", parsed.Choices[0].FinishReason)
	}
	if err := json.Unmarshal([]byte(text), target); err != nil {
		return tokensUsed, rawResponse, fmt.Errorf("parsing model output as JSON: %w (raw: %s)", err, text)
	}
	return tokensUsed, rawResponse, nil
}

// Helper functions for formatting context

func formatCandidatesWithResume(candidates []CandidateContext) string {
	var b strings.Builder
	for _, c := range candidates {
		// Truncate resume to first 200 chars for brevity
		resumePreview := c.ResumeText
		if len(resumePreview) > 200 {
			resumePreview = resumePreview[:200] + "..."
		}
		fmt.Fprintf(&b, "- id=%s name=%s skills=%s\n  resume_preview=%s\n",
			c.ID, c.Name, strings.Join(c.Skills, ", "), resumePreview)
	}
	return b.String()
}

func formatJobsDetail(jobs []JobContext) string {
	var b strings.Builder
	for _, j := range jobs {
		fmt.Fprintf(&b, "- id=%s title=%s must_have=%s nice_to_have=%s seniority=%s\n",
			j.ID, j.Title, strings.Join(j.MustHave, ", "),
			strings.Join(j.NiceToHave, ", "), j.Seniority)
	}
	return b.String()
}
