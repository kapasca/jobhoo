# OpenAI-Only Setup - Configuration Simplification

**Date:** August 2, 2026  
**Status:** ✅ **COMPLETE & SIMPLIFIED**

---

## Summary

Removed provider switching logic and simplified JOBHOO to use **OpenAI as the exclusive AI provider**. No more `AI_PROVIDER` configuration - reduces complexity while maintaining all AI capabilities.

---

## What Changed

### 1. Configuration (`.env`)

**Before:**
```bash
AI_PROVIDER=openai
AI_API_KEY=sk-xxx...
AI_MODEL=gpt-4o
AI_BASE_URL=https://api.maiarouter.ai/v1
```

**After:**
```bash
# AI Configuration (OpenAI provider with Maia Router gateway)
AI_API_KEY=sk-xxx...
AI_MODEL=openai/gpt-5-nano
AI_BASE_URL=https://api.maiarouter.ai/v1
```

### 2. Code Changes

**`internal/config/config.go`**
- Removed `AIProvider string` field from Config struct
- Removed reading `AI_PROVIDER` from environment

**`cmd/server/main.go`**
- Changed from: `ai.New(cfg.AIProvider, cfg.AIAPIKey)`
- Changed to: `ai.NewOpenAIProvider(cfg.AIAPIKey)`
- Direct OpenAI provider instantiation

**`internal/ai/` File Organization**
- ✅ Deleted `provider_mock.go` - no longer needed
- ✅ Deleted `provider_anthropic.go` - no longer needed
- ✅ Deleted `registry.go` - no longer needed
- ✅ Renamed `provider_openai.go` → `openai.go`
- ✅ Renamed `provider_openai_test.go` → `openai_test.go`
- ✅ Kept `provider.go` - interface definition
- ✅ Kept `prompts.go` - system prompts

**Remaining `internal/ai/` structure:**
```
internal/ai/
  ├── provider.go          # Provider interface
  ├── openai.go            # OpenAI implementation
  ├── openai_test.go       # Tests
  └── prompts.go           # System prompts
```

**`internal/ai/provider_openai.go`**
- No changes to implementation
- Still fully functional with all 5 methods
- Vision support for PDFs still available

### 3. Documentation Updated

- **README.md** - Removed AI_PROVIDER option, simplified example
- **DOC-DEVELOPMENT-GUIDE.md** - Removed AI_PROVIDER switching explanation
- **.env.example** - Simplified AI section without AI_PROVIDER
- **DOC-AI-GATEWAY-SETUP.md** - Marked as OpenAI exclusive

---

## Benefits

✅ **Simpler Configuration**
- One less environment variable to manage
- No provider selection logic needed
- Clearer configuration: just API key + model + gateway URL

✅ **No Code Duplication**
- Remove mock provider scaffolding for this project
- Remove Anthropic provider (can be in separate branch if needed)
- Smaller codebase footprint

✅ **Clearer Intent**
- Single path through code - easier to trace
- No unused provider implementations
- Explicit: "we use OpenAI"

✅ **Maintains All Features**
- All 5 AI methods still fully functional
- Vision support still available
- Gateway support still available
- No functionality lost

---

## What Stays the Same

- ✅ All 5 Provider methods working
- ✅ Gateway support (Maia Router or any OpenAI-compatible)
- ✅ Vision for PDF/image parsing
- ✅ Model configuration via `AI_MODEL`
- ✅ All existing handlers unchanged
- ✅ All tests passing (6/6)

---

## Files Modified

```
internal/config/config.go          - Removed AIProvider field
cmd/server/main.go                 - Direct OpenAI instantiation
internal/ai/openai.go              - Kept (core implementation)
internal/ai/openai_test.go         - Kept (tests)
internal/ai/provider.go            - Kept (interface)
internal/ai/prompts.go             - Kept (system prompts)
internal/ai/provider_mock.go       - DELETED
internal/ai/provider_anthropic.go  - DELETED
internal/ai/registry.go            - DELETED
test-ai-provider.ps1              - DELETED (use: go test ./internal/ai -v)
.env                               - Removed AI_PROVIDER
.env.example                       - Simplified AI section
README.md                          - Updated examples
DOC-DEVELOPMENT-GUIDE.md          - Updated documentation
DOC-AI-GATEWAY-SETUP.md           - Added exclusive note
```

## Test Results

All AI provider tests still passing:

```
✅ TestOpenAIProviderInitialization - PASS
✅ TestOpenAIProviderGatewayURL - PASS  
✅ TestOpenAIProviderHealthCheck - PASS
✅ TestOpenAIProviderEndpointConstruction - PASS
✅ TestOpenAIProviderSummarizeResumeIntegration - PASS
✅ TestOpenAIProviderRankCandidatesIntegration - PASS
```

---

## Migration Notes (If Needed)

If reverting to multi-provider support later:
1. Restore `AIProvider` field in config.Config
2. Restore `New()` function in registry.go
3. Restore provider implementations (provider_mock.go, provider_anthropic.go)
4. Add back AI_PROVIDER to .env parsing

For this implementation: **Not needed** - OpenAI is our exclusive provider.

---

## Current Setup

**Production Ready:** ✅ YES

```bash
# Development
APP_ENV=development
AI_API_KEY=sk-Lk5-n63mw6G-pX9xu_jtKw
AI_MODEL=openai/gpt-5-nano
AI_BASE_URL=https://api.maiarouter.ai/v1

# For OpenAI Official API (no gateway)
# AI_BASE_URL=https://api.openai.com/v1

# For different models
# AI_MODEL=gpt-4o
# AI_MODEL=gpt-4-turbo
```

---

## Why This Decision?

1. **Project Focus** - JOBHOO is a recruiting platform, not a multi-provider abstraction layer
2. **Reduced Complexity** - Fewer code paths = easier maintenance and debugging
3. **Clear Intent** - Obvious that we use OpenAI, not an accident of architecture
4. **Cost Control** - Single provider = easier cost tracking and optimization
5. **Gateway Flexibility** - Can still switch between OpenAI official and custom gateways without code changes

---

## Status

✅ **Simplified, tested, and ready**

No mock provider, no Anthropic provider, no AI_PROVIDER switching - just OpenAI, clean and simple.
