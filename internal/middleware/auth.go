// Package middleware holds cross-cutting HTTP middleware: authentication
// context and role-based route guards. Route-specific logic stays in
// handlers; only concerns that apply across many routes belong here.
package middleware

import (
	"context"
	"net/http"

	"github.com/jobhoo/jobhoo/internal/auth"
	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/models"
)

type ctxKey string

const userCtxKey ctxKey = "jobhoo_current_user"

const SessionCookieName = "jobhoo_session"

// WithUser is middleware that looks up the session cookie (if present),
// resolves it to a user, and attaches that user to the request context.
// It never rejects a request by itself — routes that require a signed-in
// user use RequireAuth in addition to this.
func WithUser(sessions *database.SessionsRepo, users *database.UsersRepo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil || cookie.Value == "" {
				next.ServeHTTP(w, r)
				return
			}

			tokenHash := auth.HashToken(cookie.Value)
			userID, err := sessions.UserIDForToken(r.Context(), tokenHash)
			if err != nil {
				// Invalid/expired session: clear the cookie and continue as anonymous.
				clearSessionCookie(w)
				next.ServeHTTP(w, r)
				return
			}

			user, err := users.GetByID(r.Context(), userID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), userCtxKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CurrentUser retrieves the signed-in user from context, if any.
func CurrentUser(r *http.Request) *models.User {
	u, _ := r.Context().Value(userCtxKey).(*models.User)
	return u
}

// RequireAuth blocks anonymous requests, redirecting browsers to /login.
// For HTMX requests it sets HX-Redirect instead of a normal 3xx, since an
// HTMX swap target (e.g. a bookmark button) should never receive a full
// login page as its "fragment".
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if CurrentUser(r) == nil {
			loginURL := "/login?next=" + r.URL.Path
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", loginURL)
				w.WriteHeader(http.StatusOK)
				return
			}
			http.Redirect(w, r, loginURL, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole blocks requests from signed-in users whose role isn't in the
// allowed set. Must be used after RequireAuth in the middleware chain.
func RequireRole(roles ...models.UserRole) func(http.Handler) http.Handler {
	allowed := make(map[models.UserRole]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := CurrentUser(r)
			if u == nil || !allowed[u.Role] {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
