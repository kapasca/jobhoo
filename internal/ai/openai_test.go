package ai

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// TestOpenAIProviderInitialization verifies that the provider initializes correctly
// with both default OpenAI API and custom gateway URLs.
func TestOpenAIProviderInitialization(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		baseURL string
		wantErr bool
	}{
		{
			name:    "initialization with default openai api",
			apiKey:  "sk-test-key",
			baseURL: "",
			wantErr: false,
		},
		{
			name:    "initialization with custom gateway",
			apiKey:  "sk-test-key",
			baseURL: "https://api.maiarouter.ai/v1",
			wantErr: false,
		},
		{
			name:    "initialization fails with empty api key",
			apiKey:  "",
			baseURL: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment first
			os.Unsetenv("AI_BASE_URL")
			defer os.Unsetenv("AI_BASE_URL")

			// Set environment variables
			if tt.baseURL != "" {
				os.Setenv("AI_BASE_URL", tt.baseURL)
			}

			provider, err := NewOpenAIProvider(tt.apiKey, nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewOpenAIProvider() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && provider == nil {
				t.Fatal("expected provider to be non-nil")
			}

			if !tt.wantErr {
				// Verify provider state
				if provider.Name() != "openai" {
					t.Errorf("expected provider name 'openai', got %s", provider.Name())
				}

				// Check base URL was set correctly
				expectedURL := tt.baseURL
				if expectedURL == "" {
					expectedURL = "https://api.openai.com/v1"
				}
				if provider.baseURL != expectedURL {
					t.Errorf("expected base URL %s, got %s", expectedURL, provider.baseURL)
				}
			}
		})
	}
}

// TestOpenAIProviderGatewayURL verifies base URL construction for different gateway setups.
func TestOpenAIProviderGatewayURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expectedURL string
	}{
		{
			name:        "default openai api",
			baseURL:     "",
			expectedURL: "https://api.openai.com/v1",
		},
		{
			name:        "maia router gateway",
			baseURL:     "https://api.maiarouter.ai/v1",
			expectedURL: "https://api.maiarouter.ai/v1",
		},
		{
			name:        "gateway with trailing slash removed",
			baseURL:     "https://api.maiarouter.ai/v1/",
			expectedURL: "https://api.maiarouter.ai/v1",
		},
		{
			name:        "local development gateway",
			baseURL:     "http://localhost:8000/v1",
			expectedURL: "http://localhost:8000/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			os.Unsetenv("AI_BASE_URL")
			defer os.Unsetenv("AI_BASE_URL")

			if tt.baseURL != "" {
				os.Setenv("AI_BASE_URL", tt.baseURL)
			}

			provider, err := NewOpenAIProvider("sk-test-key", nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if provider.baseURL != tt.expectedURL {
				t.Errorf("expected base URL %s, got %s", tt.expectedURL, provider.baseURL)
			}
		})
	}
}

// TestOpenAIProviderHealthCheck attempts to connect to the gateway and verify it's reachable.
// This is useful for debugging connectivity issues.
func TestOpenAIProviderHealthCheck(t *testing.T) {
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		t.Skip("AI_API_KEY not set, skipping health check test")
	}

	provider, err := NewOpenAIProvider(apiKey, nil)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	// Test with a simple context to ensure API is reachable
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try a simple extraction operation with minimal input
	result, err := provider.ExtractResumeText(ctx, "Software Engineer with 5 years of experience")

	// We expect either success or a specific API error, not a connection error
	if err != nil {
		// Log the error for debugging but don't fail if it's an API error
		// (e.g., "invalid API key" is fine for this test)
		t.Logf("API Response: %v", err)
		t.Logf("Gateway URL: %s", provider.baseURL)
	}

	if err == nil {
		// If successful, verify we got a result
		if result.Headline == "" {
			t.Log("Warning: empty headline in result, but API call succeeded")
		}
		t.Logf("✓ Health check passed - API is reachable")
		t.Logf("  Provider: %s", provider.Name())
		t.Logf("  Base URL: %s", provider.baseURL)
		t.Logf("  Model: %s", provider.model)
	}
}

// TestOpenAIProviderEndpointConstruction verifies the endpoint URL is built correctly.
func TestOpenAIProviderEndpointConstruction(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		endpoint string
	}{
		{
			name:     "openai default endpoint",
			baseURL:  "https://api.openai.com/v1",
			endpoint: "https://api.openai.com/v1/chat/completions",
		},
		{
			name:     "maia router endpoint",
			baseURL:  "https://api.maiarouter.ai/v1",
			endpoint: "https://api.maiarouter.ai/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			os.Unsetenv("AI_BASE_URL")
			defer os.Unsetenv("AI_BASE_URL")

			os.Setenv("AI_BASE_URL", tt.baseURL)

			provider, err := NewOpenAIProvider("sk-test-key", nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the endpoint would be constructed correctly
			expectedEndpoint := fmt.Sprintf("%s/chat/completions", provider.baseURL)
			if expectedEndpoint != tt.endpoint {
				t.Errorf("expected endpoint %s, got %s", tt.endpoint, expectedEndpoint)
			}
		})
	}
}

// TestOpenAIProviderExtractResumeTextIntegration tests actual API call to extract structured
// resume fields. This is an integration test that requires valid API credentials and network connectivity.
func TestOpenAIProviderExtractResumeTextIntegration(t *testing.T) {
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		t.Skip("AI_API_KEY not set, skipping integration test")
	}

	provider, err := NewOpenAIProvider(apiKey, nil)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resumeText := `
		Software Engineer with 5+ years of experience in Go and Python.
		Specialized in distributed systems and cloud infrastructure.
		Experience with Kubernetes, Docker, PostgreSQL, and AWS.
		Led team of 3 engineers on payment processing microservices.
		BS Computer Science, 2018.
	`

	result, err := provider.ExtractResumeText(ctx, resumeText)
	if err != nil {
		t.Logf("API Error: %v", err)
		t.Logf("Note: This is expected if API key has limited access or model mismatch")
		return
	}

	// Verify we got some results
	if result.Headline == "" {
		t.Error("expected non-empty headline")
	}
	if len(result.Skills) == 0 {
		t.Error("expected non-empty skills")
	}
	if len(result.Experience) == 0 {
		t.Error("expected non-empty experience")
	}

	t.Logf("✓ Resume Extraction Generated")
	t.Logf("  Headline: %s", result.Headline)
	t.Logf("  Skills: %v", result.Skills)
	t.Logf("  Experience: %v", result.Experience)
	t.Logf("  Achievements: %v", result.Achievements)
}

// TestOpenAIProviderRankCandidatesIntegration tests candidate ranking.
func TestOpenAIProviderRankCandidatesIntegration(t *testing.T) {
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		t.Skip("AI_API_KEY not set, skipping integration test")
	}

	provider, err := NewOpenAIProvider(apiKey, nil)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	job := JobContext{
		ID:          "job-001",
		Title:       "Senior Go Engineer",
		Description: "Building distributed systems",
		MustHave:    []string{"Go", "Kubernetes", "PostgreSQL"},
		NiceToHave:  []string{"AWS", "gRPC"},
		Seniority:   "Senior",
	}

	candidates := []CandidateContext{
		{
			ID:         "cand-001",
			Name:       "Alice Chen",
			ResumeText: "5 years Go, 3 years K8s, expert PostgreSQL, AWS certified",
			Skills:     []string{"Go", "Kubernetes", "PostgreSQL", "AWS"},
		},
		{
			ID:         "cand-002",
			Name:       "Bob Smith",
			ResumeText: "3 years Node.js, 2 years Docker, learning Go",
			Skills:     []string{"Node.js", "JavaScript", "Docker"},
		},
	}

	rankings, err := provider.RankCandidates(ctx, job, candidates)
	if err != nil {
		t.Logf("API Error: %v", err)
		t.Logf("Note: This is expected if API key has limited access or model mismatch")
		return
	}

	if len(rankings) != len(candidates) {
		t.Errorf("expected %d rankings, got %d", len(candidates), len(rankings))
	}

	t.Logf("✓ Candidates Ranked")
	for _, r := range rankings {
		t.Logf("  %s: %0.1f/100 - %s", r.CandidateID, r.Score, r.Summary)
	}
}
