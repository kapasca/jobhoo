// Command seed populates JOBHOO with realistic demo data for testing search,
// filtering, category classification, and pagination at a non-trivial scale:
// 10 recruiter accounts (one company each) and 100 jobs spread across the
// platform's 5 job categories.
//
// It is safe to re-run: it first deletes any existing rows tagged with the
// demo email domain (@jobhoo.demo), then reinserts fresh data, so it never
// touches real accounts and never silently duplicates on a second run.
//
// Usage:
//
//	go run ./cmd/seed
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/jobhoo/jobhoo/internal/auth"
	"github.com/jobhoo/jobhoo/internal/config"
	"github.com/jobhoo/jobhoo/internal/database"
)

const demoEmailDomain = "@jobhoo.demo"
const numCompanies = 10
const numJobs = 100

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	defer pool.Close()

	rng := rand.New(rand.NewSource(42)) // fixed seed: reproducible data across re-runs

	log.Println("clearing previous demo data...")
	if err := clearDemoData(ctx, pool); err != nil {
		log.Fatalf("clear error: %v", err)
	}

	log.Println("creating admin account...")
	if err := seedAdmin(ctx, pool); err != nil {
		log.Fatalf("admin seed error: %v", err)
	}

	log.Printf("creating %d recruiter accounts + companies...", numCompanies)
	companyIDs, recruiterIDs, err := seedCompanies(ctx, pool)
	if err != nil {
		log.Fatalf("company seed error: %v", err)
	}

	log.Printf("creating %d jobs across %d categories...", numJobs, len(jobCategories))
	if err := seedJobs(ctx, pool, rng, companyIDs, recruiterIDs); err != nil {
		log.Fatalf("job seed error: %v", err)
	}

	log.Println("done: 1 admin, 10 companies (1 left pending for the approval queue), and 100 jobs across 5 categories seeded.")
}

// seedAdmin creates a demo admin account. There's no signup path for the
// admin role by design (see internal/handlers/auth.go) — an admin account
// can only come from seeding or a direct database operation.
func seedAdmin(ctx context.Context, pool *pgxpool.Pool) error {
	passwordHash, err := auth.HashPassword("demo-password-123")
	if err != nil {
		return fmt.Errorf("hashing admin password: %w", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO users (email, password_hash, role, full_name)
		VALUES ($1, $2, 'admin', 'JOBHOO Admin')
	`, "admin"+demoEmailDomain, passwordHash)
	return err
}

// clearDemoData removes every row created by a previous seed run, identified
// by the reserved @jobhoo.demo email domain. Deleting users cascades to
// companies, jobs, applications, etc. via foreign keys, so this alone is
// enough to fully reset demo state.
func clearDemoData(ctx context.Context, pool *pgxpool.Pool) error {
	// Null out approved_by on demo companies before deleting the admin user
	// that set it — otherwise the FK on companies.approved_by fires first.
	if _, err := pool.Exec(ctx, `
		UPDATE companies SET approved_by = NULL
		WHERE owner_id IN (SELECT id FROM users WHERE email LIKE '%' || $1)
		   OR approved_by IN (SELECT id FROM users WHERE email LIKE '%' || $1)
	`, demoEmailDomain); err != nil {
		return err
	}
	_, err := pool.Exec(ctx, `DELETE FROM users WHERE email LIKE '%' || $1`, demoEmailDomain)
	return err
}
