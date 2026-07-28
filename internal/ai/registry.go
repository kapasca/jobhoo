package ai

import "fmt"

// New builds the configured Provider. This is the ONLY place in the codebase
// that knows about concrete provider implementations — adding a new vendor
// means adding one case here and a new provider_<name>.go file, nothing else.
// Handlers depend only on the Provider interface returned here.
func New(providerName, apiKey string) (Provider, error) {
	switch providerName {
	case "mock", "":
		return NewMockProvider(), nil
	case "anthropic":
		return NewAnthropicProvider(apiKey)
	case "openai":
		return NewOpenAIProvider(apiKey)
	default:
		return nil, fmt.Errorf("unknown AI provider: %q", providerName)
	}
}
