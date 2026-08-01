package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/handlers"
	jhmw "github.com/jobhoo/jobhoo/internal/middleware"
	"github.com/jobhoo/jobhoo/internal/models"
)

func New(h *handlers.Handlers, users *database.UsersRepo, sessions *database.SessionsRepo, staticDir string) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(15 * time.Second))
	r.Use(jhmw.WithUser(sessions, users)) // attaches CurrentUser to context for every request
	r.Use(jhmw.CSRF)                      // double-submit CSRF check on every state-changing request

	r.Get("/healthz", h.Health)

	fileServer := http.FileServer(http.Dir(staticDir))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/", h.Home)

	r.Route("/jobs", func(r chi.Router) {
		r.Get("/", h.JobsIndex)
		r.Get("/search", h.JobsSearch)
		r.Get("/{id}", h.JobDetail)

		r.Group(func(r chi.Router) {
			r.Use(jhmw.RequireAuth, jhmw.RequireRole(models.RoleCandidate))
			r.Post("/{id}/apply", h.ApplyToJob)
			r.Post("/{id}/save", h.SaveJob)
		})
	})

	// Authentication routes with rate limiting
	// 5 attempts per 15 minutes for signup and login
	authLimiter := jhmw.NewRateLimiter(5, 15*time.Minute)
	// 10 attempts per 15 minutes for email operations
	emailLimiter := jhmw.NewRateLimiter(10, 15*time.Minute)

	r.Get("/signup", h.SignupPage)
	r.With(jhmw.RateLimitMiddleware(authLimiter, 5, 15*time.Minute)).Post("/signup", h.Signup)
	r.Get("/login", h.LoginPage)
	r.With(jhmw.RateLimitMiddleware(authLimiter, 5, 15*time.Minute)).Post("/login", h.Login)
	r.With(jhmw.RateLimitMiddleware(emailLimiter, 10, 15*time.Minute)).Get("/verify-email", h.VerifyEmail)
	r.Get("/forgot-password", h.ForgotPasswordPage)
	r.With(jhmw.RateLimitMiddleware(emailLimiter, 10, 15*time.Minute)).Post("/forgot-password", h.ForgotPassword)
	r.Get("/reset-password", h.ResetPasswordPage)
	r.With(jhmw.RateLimitMiddleware(emailLimiter, 10, 15*time.Minute)).Post("/reset-password", h.ResetPassword)
	r.Post("/logout", h.Logout)

	// --- Candidate-only routes ---
	r.Group(func(r chi.Router) {
		r.Use(jhmw.RequireAuth, jhmw.RequireRole(models.RoleCandidate))
		r.Get("/dashboard/candidate", h.CandidateDashboard)
		r.Post("/dashboard/candidate/recommendations", h.RecommendedJobs)
		r.Get("/profile", h.ProfilePage)
		r.Post("/profile", h.ProfileUpdate)
		r.Post("/profile/suggestions", h.ResumeSuggestions)
	})

	// --- Recruiter-only routes ---
	r.Group(func(r chi.Router) {
		r.Use(jhmw.RequireAuth, jhmw.RequireRole(models.RoleRecruiter))
		r.Get("/dashboard/recruiter", h.RecruiterDashboard)
		r.Get("/company/setup", h.CompanySetupPage)
		r.Post("/company/setup", h.CompanySetup)
		r.Get("/company/profile", h.CompanyProfilePage)
		r.Post("/company/profile", h.CompanyProfileUpdate)
		r.Get("/company/public", h.CompanyPublicRedirect)
		r.Post("/company/resubmit", h.CompanyResubmit)
		r.Get("/post-job", h.PostJobPage)
		r.Post("/post-job", h.PostJob)
		r.Get("/recruiter/jobs/{id}/edit", h.EditJobPage)
		r.Post("/recruiter/jobs/{id}/edit", h.EditJob)
		r.Post("/recruiter/jobs/{id}/close", h.CloseJob)
		r.Post("/recruiter/jobs/{id}/archive", h.ArchiveJob)
		r.Post("/recruiter/jobs/{id}/reopen", h.ReopenJob)
		r.Post("/recruiter/jobs/{id}/delete", h.DeleteJob)
		r.Get("/recruiter/jobs/{id}/applicants", h.ATSBoard)
		r.Post("/recruiter/jobs/{id}/rank", h.RankCandidates)
		r.Post("/ats/applications/{id}/stage", h.UpdateApplicationStage)
	})

	// --- Admin-only routes ---
	r.Group(func(r chi.Router) {
		r.Use(jhmw.RequireAuth, jhmw.RequireRole(models.RoleAdmin))
		r.Get("/dashboard/admin", h.AdminDashboard)
		r.Get("/admin/approvals", h.CompanyApprovalQueue)
		r.Post("/admin/approvals/{id}/approve", h.ApproveCompany)
		r.Post("/admin/approvals/{id}/reject", h.RejectCompany)
		r.Post("/admin/approvals/{id}/blacklist", h.BlacklistCompany)
		r.Post("/admin/users/{id}/freeze", h.FreezeUser)
		r.Post("/admin/users/{id}/unfreeze", h.UnfreezeUser)
		r.Post("/admin/jobs/{id}/freeze", h.FreezeJob)
		r.Post("/admin/jobs/{id}/unfreeze", h.UnfreezeJob)
		// Admin modal detail routes (API)
		r.Get("/admin/api/users/{userID}", h.AdminUserDetail)
		r.Get("/admin/api/candidates/{candidateID}", h.AdminCandidateDetail)
		r.Get("/admin/api/recruiters/{recruiterID}", h.AdminRecruiterDetail)
		r.Get("/admin/api/companies/{companyID}", h.AdminCompanyDetail)
		r.Get("/admin/api/jobs/{jobID}", h.AdminJobDetail)
	})

	r.Get("/companies", h.CompaniesDirectory)
	r.Get("/companies/{id}", h.CompanyPublicDetail)

	r.NotFound(h.NotFoundPage)

	return r
}
