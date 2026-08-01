package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"fmt"

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
	Token    string
}

func (h *Handlers) SignupPage(w http.ResponseWriter, r *http.Request) {
	h.Render.Render(w, http.StatusOK, "signup.html", authPageData{BasePageData: newBasePageData(r, "signup")})
}

func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, resumeMaxBytes+4096)
	if err := r.ParseMultipartForm(resumeMaxBytes); err != nil {
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
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

	// Candidates must supply a resume at registration.
	if role == models.RoleCandidate {
		if _, _, err := r.FormFile("resume_file"); err == http.ErrMissingFile {
			data.Error = "Please upload your resume to complete registration as a candidate."
			h.Render.Render(w, http.StatusBadRequest, "signup.html", data)
			return
		}
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

	// Create and send email verification token (advisory: not required to use site).
	rawToken, _, err := auth.NewSessionToken()
	if err == nil {
		// token valid for 48 hours
		_ = h.Tokens.CreateEmailVerification(r.Context(), user.ID, rawToken, 48*time.Hour)
		scheme := "https"
		if r.TLS == nil {
			scheme = "http"
		}
		link := fmt.Sprintf("%s://%s/verify-email?token=%s", scheme, r.Host, rawToken)
		subj := "Verify your JOBHOO email"
		text := "Please verify your email by visiting: " + link
		html := `<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #001d3d; margin-bottom: 1rem;">Verify Your Email</h2>
			<p>Welcome to JOBHOO! Please verify your email address by clicking the link below:</p>
			<div style="text-align: center; margin: 1.5rem 0;">
				<a href="` + link + `" style="display: inline-block; padding: 12px 24px; background-color: #ff9500; color: #fff; text-decoration: none; border-radius: 4px; font-weight: bold;">Verify Email</a>
			</div>
			<p style="color: #666; font-size: 0.9rem;">If you didn't create this account, please ignore this email.</p>
			<p style="color: #666; font-size: 0.85rem; margin-top: 1.5rem;">This link expires in 48 hours.</p>
		</div>`
		userID := user.ID
		_ = h.Email.SendWithUserID(r.Context(), user.Email, subj, html, text, "email_verification", &userID)
	}

	// Persist the resume file if one was uploaded during candidate signup.
	// Validation already confirmed a file exists at this point.
	if role == models.RoleCandidate {
		if resumeURL, err := handleResumeUpload(r); err != nil {
			// File present but invalid type/size — create profile without URL.
			_ = h.Profiles.Upsert(r.Context(), user.ID, "", "", "", "", []string{})
		} else if resumeURL != "" {
			_ = h.Profiles.Upsert(r.Context(), user.ID, "", "", resumeURL, "", []string{})
		}
	}

	http.Redirect(w, r, dashboardPathForRole(user.Role), http.StatusSeeOther)
}

func (h *Handlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := authPageData{BasePageData: newBasePageData(r, "login"), NextURL: r.URL.Query().Get("next")}
	if r.URL.Query().Get("reason") == "frozen" {
		data.Error = "Your session was terminated because your account has been frozen by JOBHOO administrators."
	}

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

	if user.IsFrozen {
		data.Error = "Your account has been frozen by the administrator. Please contact JOBHOO support for more information."
		if r.Header.Get("HX-Request") == "true" {
			h.Render.RenderBlock(w, "login-modal.html", "login-modal", data)
		} else {
			h.Render.Render(w, http.StatusForbidden, "login.html", data)
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

// VerifyEmail handles GET /verify-email?token=xxxxx
func (h *Handlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}
	userID, err := h.Tokens.ConsumeEmailVerification(r.Context(), token)
	if err != nil {
		h.Render.Render(w, http.StatusOK, "verify-email.html", authPageData{BasePageData: newBasePageData(r, "verify-email"), Error: "Invalid or expired token."})
		return
	}
	if err := h.Users.SetEmailVerified(r.Context(), userID); err != nil {
		http.Error(w, "could not verify email", http.StatusInternalServerError)
		return
	}
	h.Render.Render(w, http.StatusOK, "verify-email.html", authPageData{BasePageData: newBasePageData(r, "verify-email")})
}

// ForgotPasswordPage shows the forgot password form
func (h *Handlers) ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	h.Render.Render(w, http.StatusOK, "forgot-password.html", authPageData{BasePageData: newBasePageData(r, "forgot-password")})
}

// ForgotPassword handles form post to request a password reset
func (h *Handlers) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	// Always show success page to avoid account enumeration
	data := authPageData{BasePageData: newBasePageData(r, "forgot-password")}
	user, err := h.Users.GetByEmail(r.Context(), email)
	if err == nil && user != nil {
		rawToken, _, err2 := auth.NewSessionToken()
		if err2 == nil {
			_ = h.Tokens.CreatePasswordReset(r.Context(), user.ID, rawToken, 2*time.Hour)
			scheme := "https"
			if r.TLS == nil {
				scheme = "http"
			}
			link := fmt.Sprintf("%s://%s/reset-password?token=%s", scheme, r.Host, rawToken)
			subj := "Reset Your JOBHOO Password"
			text := "Click this link to reset your password: " + link + "\n\nThis link expires in 2 hours."
			html := `<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
				<h2 style="color: #001d3d; margin-bottom: 1rem;">Password Reset Request</h2>
				<p>We received a request to reset your JOBHOO password. Click the button below to set a new password:</p>
				<div style="text-align: center; margin: 1.5rem 0;">
					<a href="` + link + `" style="display: inline-block; padding: 12px 24px; background-color: #ff9500; color: #fff; text-decoration: none; border-radius: 4px; font-weight: bold;">Reset Password</a>
				</div>
				<p style="color: #666; font-size: 0.9rem;">If you didn't request this, please ignore this email.</p>
				<p style="color: #666; font-size: 0.85rem; margin-top: 1.5rem;">This link expires in 2 hours.</p>
			</div>`
			userID := user.ID
			_ = h.Email.SendWithUserID(r.Context(), user.Email, subj, html, text, "password_reset", &userID)
		}
	}
	data.Error = "If an account exists for that email, a reset link has been sent."
	h.Render.Render(w, http.StatusOK, "forgot-password.html", data)
}

// ResetPasswordPage shows a form to set a new password (token required)
func (h *Handlers) ResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	data := authPageData{BasePageData: newBasePageData(r, "reset-password"), Error: "", Token: token}
	if token == "" {
		data.Error = "Missing token."
	}
	h.Render.Render(w, http.StatusOK, "reset-password.html", data)
}

// ResetPassword processes the new password submission
func (h *Handlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	token := r.FormValue("token")
	pw := r.FormValue("password")
	if token == "" || pw == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	userID, err := h.Tokens.ConsumePasswordReset(r.Context(), token)
	if err != nil {
		h.Render.Render(w, http.StatusBadRequest, "reset-password.html", authPageData{BasePageData: newBasePageData(r, "reset-password"), Error: "Invalid or expired token."})
		return
	}
	hash, err := auth.HashPassword(pw)
	if err != nil {
		h.Render.Render(w, http.StatusBadRequest, "reset-password.html", authPageData{BasePageData: newBasePageData(r, "reset-password"), Error: "Password must be at least 8 characters."})
		return
	}
	if err := h.Users.SetPasswordHash(r.Context(), userID, hash); err != nil {
		http.Error(w, "could not set password", http.StatusInternalServerError)
		return
	}
	// Revoke all existing sessions for security
	_ = h.Sessions.RevokeAllForUser(r.Context(), userID)
	h.Render.Render(w, http.StatusOK, "reset-password.html", authPageData{BasePageData: newBasePageData(r, "reset-password")})
}
