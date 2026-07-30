package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jobhoo/jobhoo/internal/auth"
	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/middleware"
	"github.com/jobhoo/jobhoo/internal/models"
)

const sessionTTL = 30 * 24 * time.Hour // 30 days

type authPageData struct {
	BasePageData
	Error    string
	NextURL  string
	Email    string
	FullName string
	Role     string
}

func (h *Handlers) SignupPage(w http.ResponseWriter, r *http.Request) {
	h.Render.Render(w, http.StatusOK, "signup.html", authPageData{BasePageData: newBasePageData(r, "signup")})
}

func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")
	fullName := strings.TrimSpace(r.FormValue("full_name"))
	roleInput := r.FormValue("role")

	data := authPageData{BasePageData: newBasePageData(r, "signup"), Email: email, FullName: fullName, Role: roleInput}

	role := models.RoleCandidate
	if roleInput == "recruiter" {
		role = models.RoleRecruiter
	}

	if email == "" || password == "" || fullName == "" {
		data.Error = "All fields are required."
		h.Render.Render(w, http.StatusBadRequest, "signup.html", data)
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		data.Error = "Password must be at least 8 characters."
		h.Render.Render(w, http.StatusBadRequest, "signup.html", data)
		return
	}

	user, err := h.Users.Create(r.Context(), email, hash, fullName, role)
	if err != nil {
		if errors.Is(err, database.ErrDuplicateEmail) {
			data.Error = "An account with that email already exists."
		} else {
			data.Error = "Something went wrong. Please try again."
		}
		h.Render.Render(w, http.StatusBadRequest, "signup.html", data)
		return
	}

	if err := h.startSession(w, r, user.ID); err != nil {
		http.Error(w, "could not start session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, dashboardPathForRole(user.Role), http.StatusSeeOther)
}

func (h *Handlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := authPageData{BasePageData: newBasePageData(r, "login"), NextURL: r.URL.Query().Get("next")}

	// HTMX requests (clicking "Sign in" in the nav) get the modal fragment.
	// Everything else — RequireAuth's redirect when an anonymous user hits a
	// protected page, a direct/bookmarked visit to /login, browser
	// back/forward, a shared link — is a normal browser navigation and must
	// get a real page, or the user hits a dead end.
	if r.Header.Get("HX-Request") == "true" {
		h.Render.RenderBlock(w, "login-modal.html", "login-modal", data)
		return
	}

	h.Render.Render(w, http.StatusOK, "login.html", data)
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")
	next := r.FormValue("next")

	data := authPageData{BasePageData: newBasePageData(r, "login"), Email: email, NextURL: next}

	user, err := h.Users.GetByEmail(r.Context(), email)
	if err != nil || !auth.VerifyPassword(user.PasswordHash, password) {
		data.Error = "Incorrect email or password."
		if r.Header.Get("HX-Request") == "true" {
			h.Render.RenderBlock(w, "login-modal.html", "login-modal", data)
		} else {
			h.Render.Render(w, http.StatusUnauthorized, "login.html", data)
		}
		return
	}

	if err := h.startSession(w, r, user.ID); err != nil {
		http.Error(w, "could not start session", http.StatusInternalServerError)
		return
	}

	// For HTMX requests, redirect via HX-Redirect header
	if r.Header.Get("HX-Request") == "true" {
		redirectTo := next
		if redirectTo == "" {
			redirectTo = dashboardPathForRole(user.Role)
		}
		w.Header().Set("HX-Redirect", redirectTo)
		w.WriteHeader(http.StatusOK)
		return
	}

	// For normal requests, redirect as before
	if next != "" {
		http.Redirect(w, r, next, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, dashboardPathForRole(user.Role), http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(middleware.SessionCookieName); err == nil {
		_ = h.Sessions.Revoke(r.Context(), auth.HashToken(cookie.Value))
	}
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) startSession(w http.ResponseWriter, r *http.Request, userID string) error {
	rawToken, tokenHash, err := auth.NewSessionToken()
	if err != nil {
		return err
	}
	if err := h.Sessions.Create(r.Context(), userID, tokenHash, r.UserAgent(), r.RemoteAddr, sessionTTL); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    rawToken,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure should be true once served over HTTPS in production.
	})
	return nil
}

func dashboardPathForRole(role models.UserRole) string {
	switch role {
	case models.RoleRecruiter:
		return "/dashboard/recruiter"
	case models.RoleAdmin:
		return "/dashboard/admin"
	default:
		return "/dashboard/candidate"
	}
}
