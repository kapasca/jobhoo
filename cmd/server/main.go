package main

import (
	"log"
	"net/http"
	"os"

	"jobhoo/internal/db"
	"jobhoo/internal/handlers"
	"jobhoo/internal/middleware"
	"jobhoo/internal/models"
	"jobhoo/internal/repository"
	"jobhoo/internal/services/aimatching"
	authsvc "jobhoo/internal/services/auth"
)

func main() {
	conn, err := db.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close()

	migrationsDir := getenv("MIGRATIONS_DIR", "migrations")
	if err := db.Migrate(conn, migrationsDir); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	uploadDir := getenv("UPLOAD_DIR", "uploads")
	templatesDir := getenv("TEMPLATES_DIR", "web/templates")
	staticDir := getenv("STATIC_DIR", "web/static")

	renderer, err := handlers.NewRenderer(templatesDir)
	if err != nil {
		log.Fatalf("failed to load templates: %v", err)
	}

	userRepo := repository.NewUserRepo(conn)
	sessionRepo := repository.NewSessionRepo(conn)
	jobRepo := repository.NewJobRepo(conn)
	applicationRepo := repository.NewApplicationRepo(conn)

	authService := authsvc.NewService(sessionRepo)
	aiProvider := aimatching.NewProvider()

	app := &handlers.App{
		DB:           conn,
		Users:        userRepo,
		Sessions:     sessionRepo,
		Jobs:         jobRepo,
		Applications: applicationRepo,
		Auth:         authService,
		AIMatching:   aiProvider,
		Render:       renderer,
		UploadDir:    uploadDir,
	}

	authMW := middleware.NewAuthMiddleware(authService, userRepo)

	candidateOnly := middleware.RequireRole(models.RoleCandidate)
	recruiterOnly := middleware.RequireRole(models.RoleRecruiter)
	adminOnly := middleware.RequireRole(models.RoleSuperAdmin)

	mux := http.NewServeMux()

	// --- Static files & uploads ---
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))

	// --- Public pages ---
	mux.HandleFunc("GET /{$}", app.HomePage)
	mux.HandleFunc("GET /jobs", app.JobsListPage)
	mux.HandleFunc("GET /jobs/{id}", app.JobDetailPage)
	mux.HandleFunc("POST /jobs/{id}/apply", candidateOnly(app.JobApplySubmit))

	mux.HandleFunc("GET /login", app.LoginPage)
	mux.HandleFunc("POST /login", app.LoginSubmit)
	mux.HandleFunc("POST /logout", app.Logout)

	mux.HandleFunc("GET /register", app.RegisterChoicePage)
	mux.HandleFunc("GET /register/candidate", app.RegisterCandidatePage)
	mux.HandleFunc("POST /register/candidate", app.RegisterCandidateSubmit)
	mux.HandleFunc("GET /register/recruiter", app.RegisterRecruiterPage)
	mux.HandleFunc("POST /register/recruiter", app.RegisterRecruiterSubmit)

	// --- Candidate ---
	mux.HandleFunc("GET /dashboard", candidateOnly(app.CandidateDashboardPage))
	mux.HandleFunc("GET /applications", candidateOnly(app.CandidateApplicationsPage))
	mux.HandleFunc("GET /resume", candidateOnly(app.CandidateResumePage))
	mux.HandleFunc("POST /resume", candidateOnly(app.CandidateResumeSubmit))

	// --- Recruiter ---
	mux.HandleFunc("GET /recruiter/pending-approval", recruiterOnly(app.RecruiterPendingApprovalPage))
	mux.HandleFunc("GET /recruiter/dashboard", middleware.RequireApprovedRecruiter(app.RecruiterDashboardPage))
	mux.HandleFunc("GET /recruiter/jobs", middleware.RequireApprovedRecruiter(app.RecruiterJobsPage))
	mux.HandleFunc("GET /recruiter/jobs/new", middleware.RequireApprovedRecruiter(app.RecruiterJobNewPage))
	mux.HandleFunc("POST /recruiter/jobs/new", middleware.RequireApprovedRecruiter(app.RecruiterJobCreateSubmit))
	mux.HandleFunc("GET /recruiter/jobs/{id}/edit", middleware.RequireApprovedRecruiter(app.RecruiterJobEditPage))
	mux.HandleFunc("POST /recruiter/jobs/{id}/edit", middleware.RequireApprovedRecruiter(app.RecruiterJobEditSubmit))
	mux.HandleFunc("POST /recruiter/jobs/{id}/close", middleware.RequireApprovedRecruiter(app.RecruiterJobCloseSubmit))
	mux.HandleFunc("GET /recruiter/jobs/{id}/applicants", middleware.RequireApprovedRecruiter(app.RecruiterJobApplicantsPage))
	mux.HandleFunc("POST /recruiter/applications/{id}/stage", middleware.RequireApprovedRecruiter(app.RecruiterApplicationStageUpdate))
	mux.HandleFunc("POST /recruiter/applications/{id}/final-status", middleware.RequireApprovedRecruiter(app.RecruiterApplicationFinalStatusUpdate))

	// --- Super Admin ---
	mux.HandleFunc("GET /admin/dashboard", adminOnly(app.AdminDashboardPage))
	mux.HandleFunc("GET /admin/recruiters", adminOnly(app.AdminRecruitersPage))
	mux.HandleFunc("POST /admin/recruiters/{id}/approve", adminOnly(app.AdminRecruiterApproveSubmit))
	mux.HandleFunc("POST /admin/recruiters/{id}/reject", adminOnly(app.AdminRecruiterRejectSubmit))
	mux.HandleFunc("GET /admin/candidates", adminOnly(app.AdminCandidatesPage))
	mux.HandleFunc("GET /admin/jobs", adminOnly(app.AdminJobsPage))

	handler := authMW.LoadUser(mux)

	port := getenv("PORT", "8080")
	log.Printf("JOBHOO server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
