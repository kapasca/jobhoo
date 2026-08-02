package ai

import (
	"context"
	"time"
)

// AICallLog records everything about a single AI call for the admin
// Activity page: the exact prompt sent, the model's raw response, timing,
// and — via context (see WithUserContext) — who triggered it.
type AICallLog struct {
	Timestamp   time.Time
	Provider    string
	Method      string
	Prompt      string
	Response    string
	ExecutionMs int
	Status      string
	Error       string
	TokensUsed  int

	UserID    string
	UserName  string
	UserEmail string
	UserRole  string
}

// CallLogger persists AI call logs. Implemented by
// internal/database.AICallLogsRepo; kept as an interface here so this
// package doesn't need to depend on the database layer.
type CallLogger interface {
	Log(ctx context.Context, entry AICallLog)
}

// userContextKey namespaces the context value set by WithUserContext so it
// never collides with keys used by other packages.
type userContextKey struct{}

// UserContext identifies who triggered an AI call, so the Activity log can
// show which candidate/recruiter/admin used a given feature.
type UserContext struct {
	ID    string
	Name  string
	Email string
	Role  string
}

// WithUserContext attaches the current user to ctx so AI call logging can
// record who triggered the request. Handlers should call this before every
// Provider method call.
func WithUserContext(ctx context.Context, u UserContext) context.Context {
	return context.WithValue(ctx, userContextKey{}, u)
}

// userFromContext retrieves the user attached via WithUserContext, if any.
func userFromContext(ctx context.Context) (UserContext, bool) {
	u, ok := ctx.Value(userContextKey{}).(UserContext)
	return u, ok
}
