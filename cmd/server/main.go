package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/jobhoo/jobhoo/internal/ai"
	"github.com/jobhoo/jobhoo/internal/config"
	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/email"
	"github.com/jobhoo/jobhoo/internal/handlers"
	"github.com/jobhoo/jobhoo/internal/router"
)

func main() {
	_ = godotenv.Load() // no-op if .env is absent (e.g. in production containers)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	defer pool.Close()

	aiProvider, err := ai.New(cfg.AIProvider, cfg.AIAPIKey)
	if err != nil {
		log.Fatalf("ai provider error: %v", err)
	}
	log.Printf("AI provider active: %s", aiProvider.Name())
	log.Printf("Email provider active: %s", cfg.EmailProvider)

	renderer, err := handlers.NewRenderer("web/templates")
	if err != nil {
		log.Fatalf("template loading error: %v", err)
	}

	usersRepo := database.NewUsersRepo(pool)
	sessionsRepo := database.NewSessionsRepo(pool)
	jobsRepo := database.NewJobsRepo(pool)
	companiesRepo := database.NewCompaniesRepo(pool)
	applicationsRepo := database.NewApplicationsRepo(pool)
	savedJobsRepo := database.NewSavedJobsRepo(pool)
	profilesRepo := database.NewCandidateProfilesRepo(pool)
	tokensRepo := database.NewEmailTokensRepo(pool)
	emailLogsRepo := database.NewEmailLogsRepo(pool)

	// Initialize base email sender based on config
	var baseSender email.Sender
	switch cfg.EmailProvider {
	case "smtp":
		baseSender = email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.EmailFrom)
	default:
		baseSender = email.NewDevSender()
	}

	// Wrap with logging sender
	emailSender := email.NewLoggingSender(baseSender, emailLogsRepo)

	h := handlers.New(
		renderer, jobsRepo, usersRepo, sessionsRepo, companiesRepo,
		applicationsRepo, savedJobsRepo, profilesRepo, aiProvider,
		emailSender, tokensRepo,
	)
	mux := router.New(h, usersRepo, sessionsRepo, "web/static")

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("JOBHOO listening on :%s (env=%s)", cfg.Port, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
