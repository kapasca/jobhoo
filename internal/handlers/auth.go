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
	Success  string
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

	// Create and send email verification token (required to login).
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
		html := `<!DOCTYPE html>
		<html style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6;">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Verify Your Email - JOBHOO</title>
		</head>
		<body style="margin: 0; padding: 0; background-color: #f5f7fa;">
			<table role="presentation" style="width: 100%; border-collapse: collapse; background-color: #f5f7fa; padding: 20px 0;">
				<tr>
					<td align="center" style="padding: 40px 20px;">
						<table role="presentation" style="max-width: 600px; width: 100%; border-collapse: collapse; background-color: #fff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
							<!-- Header -->
							<tr>
								<td style="background: linear-gradient(135deg, #001d3d 0%, #003366 100%); padding: 40px 30px; text-align: center; border-radius: 8px 8px 0 0;">
									<h1 style="color: #fff; margin: 0; font-size: 28px; font-weight: 700;">JOBHOO</h1>
									<p style="color: rgba(255,255,255,0.8); margin: 8px 0 0 0; font-size: 14px;">Connecting Jobs & Talent</p>
								</td>
							</tr>
							<!-- Content -->
							<tr>
								<td style="padding: 40px 30px; color: #333;">
									<h2 style="color: #001d3d; margin: 0 0 20px 0; font-size: 24px; font-weight: 600;">Verify Your Email Address</h2>
									<p style="margin: 0 0 10px 0; color: #555; font-size: 15px;">Welcome to JOBHOO! You're almost there.</p>
									<p style="margin: 0 0 30px 0; color: #777; font-size: 14px;">Please verify your email address to complete your registration and start exploring opportunities.</p>
									<table role="presentation" style="width: 100%; border-collapse: collapse; margin: 30px 0;">
										<tr>
											<td align="center" style="padding: 20px 0;">
												<a href="` + link + `" style="display: inline-block; background-color: #ff9500; color: #fff; padding: 14px 40px; text-decoration: none; border-radius: 6px; font-weight: 600; font-size: 16px; transition: background-color 0.3s ease;">Verify Email Address</a>
											</td>
										</tr>
									</table>
									<p style="margin: 20px 0 0 0; padding-top: 20px; border-top: 1px solid #eee; color: #999; font-size: 13px; line-height: 1.8;">
										<strong>Can't click the button?</strong><br>
										Copy and paste this link in your browser:<br>
										<span style="word-break: break-all; color: #666;">` + link + `</span>
									</p>
								</td>
							</tr>
							<!-- Footer -->
							<tr>
								<td style="background-color: #f9fafb; padding: 30px; text-align: center; border-radius: 0 0 8px 8px; border-top: 1px solid #eee;">
									<p style="margin: 0 0 10px 0; color: #999; font-size: 12px;">
										<strong>Link expires in 48 hours</strong>
									</p>
									<p style="margin: 10px 0; color: #bbb; font-size: 12px;">
										If you didn't create this account, you can safely ignore this email.
									</p>
									<p style="margin: 15px 0 0 0; padding-top: 15px; border-top: 1px solid #ddd; color: #aaa; font-size: 11px;">
										© 2026 JOBHOO. All rights reserved.
									</p>
								</td>
							</tr>
						</table>
					</td>
				</tr>
			</table>
		</body>
		</html>`
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

	// Show verification message instead of auto-login
	data.Success = "Account created successfully, check your email to verify your address before logging in."
	h.Render.Render(w, http.StatusOK, "signup.html", data)
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

	if !user.EmailVerified {
		data.Error = "Please verify your email before logging in. Check your inbox for the verification link."
		if r.Header.Get("HX-Request") == "true" {
			h.Render.RenderBlock(w, "login-modal.html", "login-modal", data)
		} else {
			h.Render.Render(w, http.StatusForbidden, "login.html", data)
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
			html := `<!DOCTYPE html>
			<html style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6;">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Reset Your Password - JOBHOO</title>
			</head>
			<body style="margin: 0; padding: 0; background-color: #f5f7fa;">
				<table role="presentation" style="width: 100%; border-collapse: collapse; background-color: #f5f7fa; padding: 20px 0;">
					<tr>
						<td align="center" style="padding: 40px 20px;">
							<table role="presentation" style="max-width: 600px; width: 100%; border-collapse: collapse; background-color: #fff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
								<!-- Header -->
								<tr>
									<td style="background: linear-gradient(135deg, #001d3d 0%, #003366 100%); padding: 40px 30px; text-align: center; border-radius: 8px 8px 0 0;">
										<h1 style="color: #fff; margin: 0; font-size: 28px; font-weight: 700;">JOBHOO</h1>
										<p style="color: rgba(255,255,255,0.8); margin: 8px 0 0 0; font-size: 14px;">Connecting Jobs & Talent</p>
									</td>
								</tr>
								<!-- Content -->
								<tr>
									<td style="padding: 40px 30px; color: #333;">
										<h2 style="color: #001d3d; margin: 0 0 20px 0; font-size: 24px; font-weight: 600;">Reset Your Password</h2>
										<p style="margin: 0 0 10px 0; color: #555; font-size: 15px;">We received a request to reset your password.</p>
										<p style="margin: 0 0 30px 0; color: #777; font-size: 14px;">Click the button below to set a new password. This link is only valid for 2 hours.</p>
										<table role="presentation" style="width: 100%; border-collapse: collapse; margin: 30px 0;">
											<tr>
												<td align="center" style="padding: 20px 0;">
													<a href="` + link + `" style="display: inline-block; background-color: #ff9500; color: #fff; padding: 14px 40px; text-decoration: none; border-radius: 6px; font-weight: 600; font-size: 16px; transition: background-color 0.3s ease;">Reset Password</a>
												</td>
											</tr>
										</table>
										<p style="margin: 20px 0 0 0; padding-top: 20px; border-top: 1px solid #eee; color: #999; font-size: 13px; line-height: 1.8;">
											<strong>Can't click the button?</strong><br>
											Copy and paste this link in your browser:<br>
											<span style="word-break: break-all; color: #666;">` + link + `</span>
										</p>
									</td>
								</tr>
								<!-- Footer -->
								<tr>
									<td style="background-color: #f9fafb; padding: 30px; text-align: center; border-radius: 0 0 8px 8px; border-top: 1px solid #eee;">
										<p style="margin: 0 0 10px 0; color: #999; font-size: 12px;">
											<strong>Link expires in 2 hours</strong>
										</p>
										<p style="margin: 10px 0; color: #bbb; font-size: 12px;">
											If you didn't request this, you can safely ignore this email.
										</p>
										<p style="margin: 15px 0 0 0; padding-top: 15px; border-top: 1px solid #ddd; color: #aaa; font-size: 11px;">
											© 2026 JOBHOO. All rights reserved.
										</p>
									</td>
								</tr>
							</table>
						</td>
					</tr>
				</table>
			</body>
			</html>`
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
