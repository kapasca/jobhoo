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
	"log"
	"math/rand"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

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

	log.Printf("creating %d recruiter accounts + companies...", numCompanies)
	companyIDs, recruiterIDs, err := seedCompanies(ctx, pool)
	if err != nil {
		log.Fatalf("company seed error: %v", err)
	}

	log.Printf("creating %d jobs across %d categories...", numJobs, len(jobCategories))
	if err := seedJobs(ctx, pool, rng, companyIDs, recruiterIDs); err != nil {
		log.Fatalf("job seed error: %v", err)
	}

	log.Println("done: 10 companies and 100 jobs across 5 categories seeded.")
}

// clearDemoData removes every row created by a previous seed run, identified
// by the reserved @jobhoo.demo email domain. Deleting users cascades to
// companies, jobs, applications, etc. via foreign keys, so this alone is
// enough to fully reset demo state.
func clearDemoData(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `DELETE FROM users WHERE email LIKE '%' || $1`, demoEmailDomain)
	return err
}
