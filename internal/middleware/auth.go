package middleware

import (
	"context"
	"net/http"

	"jobhoo/internal/models"
	"jobhoo/internal/repository"
	authsvc "jobhoo/internal/services/auth"
)

type contextKey string

const userContextKey contextKey = "current_user"

type AuthMiddleware struct {
	auth  *authsvc.Service
	users *repository.UserRepo
}

func NewAuthMiddleware(auth *authsvc.Service, users *repository.UserRepo) *AuthMiddleware {
	return &AuthMiddleware{auth: auth, users: users}
}

// LoadUser reads the session cookie (if any) and, when valid, attaches the
// authenticated user to the request context. It never blocks the request —
// use RequireAuth / RequireRole for that.
func (m *AuthMiddleware) LoadUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(authsvc.SessionCookieName)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		userID, err := m.auth.UserIDFromSession(cookie.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.users.GetByID(userID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		authUser := &models.AuthUser{ID: user.ID, Email: user.Email, Role: user.Role}
		if user.Role == models.RoleRecruiter {
			if rp, err := m.users.GetRecruiterProfile(user.ID); err == nil {
				authUser.RecruiterStatus = rp.Status
			}
		}

		ctx := context.WithValue(r.Context(), userContextKey, authUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CurrentUser(r *http.Request) *models.AuthUser {
	u, _ := r.Context().Value(userContextKey).(*models.AuthUser)
	return u
}

// RequireAuth redirects unauthenticated requests to the login page.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if CurrentUser(r) == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// RequireRole restricts access to one or more roles. It assumes RequireAuth
// (or LoadUser + a nil check) already ran.
func RequireRole(roles ...models.UserRole) func(http.HandlerFunc) http.HandlerFunc {
	allowed := make(map[models.UserRole]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			u := CurrentUser(r)
			if u == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			if !allowed[u.Role] {
				http.Error(w, "403 Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		}
	}
}

// RequireApprovedRecruiter additionally checks that a recruiter has been approved.
func RequireApprovedRecruiter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := CurrentUser(r)
		if u == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if u.Role != models.RoleRecruiter {
			http.Error(w, "403 Forbidden", http.StatusForbidden)
			return
		}
		if u.RecruiterStatus != models.RecruiterApproved {
			http.Redirect(w, r, "/recruiter/pending-approval", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}
