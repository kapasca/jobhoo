package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jobhoo/jobhoo/internal/auth"
)

// seedCompanies creates one recruiter user + one company per entry in
// companyNames, returning their IDs in matching order so seedJobs can pick a
// random (company, its own recruiter) pair for each job.
func seedCompanies(ctx context.Context, pool *pgxpool.Pool) (companyIDs []string, recruiterIDs []string, err error) {
	passwordHash, err := auth.HashPassword("demo-password-123")
	if err != nil {
		return nil, nil, fmt.Errorf("hashing demo password: %w", err)
	}

	for i, c := range companyNames {
		if i >= numCompanies {
			break
		}

		email := fmt.Sprintf("recruiter%d%s", i+1, demoEmailDomain)
		fullName := fmt.Sprintf("%s Recruiting Team", c.Name)

		var recruiterID string
		err := pool.QueryRow(ctx, `
			INSERT INTO users (email, password_hash, role, full_name)
			VALUES ($1, $2, 'recruiter', $3)
			RETURNING id
		`, email, passwordHash, fullName).Scan(&recruiterID)
		if err != nil {
			return nil, nil, fmt.Errorf("inserting recruiter %s: %w", email, err)
		}

		var companyID string
		err = pool.QueryRow(ctx, `
			INSERT INTO companies (owner_id, name, industry, description)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, recruiterID, c.Name, c.Industry, fmt.Sprintf("%s is hiring across %s.", c.Name, c.Industry)).Scan(&companyID)
		if err != nil {
			return nil, nil, fmt.Errorf("inserting company %s: %w", c.Name, err)
		}

		companyIDs = append(companyIDs, companyID)
		recruiterIDs = append(recruiterIDs, recruiterID)
	}

	return companyIDs, recruiterIDs, nil
}

// seedJobs creates numJobs jobs distributed evenly across every category
// (numJobs/len(jobCategories) each), each assigned to a random company, with
// varied location/employment/salary/skills/publish date so search, filter,
// sort, and pagination all have real variety to exercise.
func seedJobs(ctx context.Context, pool *pgxpool.Pool, rng *rand.Rand, companyIDs, recruiterIDs []string) error {
	perCategory := numJobs / len(jobCategories)
	now := time.Now()

	for _, cat := range jobCategories {
		for i := 0; i < perCategory; i++ {
			companyIdx := rng.Intn(len(companyIDs))
			companyID := companyIDs[companyIdx]
			recruiterID := recruiterIDs[companyIdx]

			title := cat.Titles[rng.Intn(len(cat.Titles))]
			seniority := seniorities[rng.Intn(len(seniorities))]
			fullTitle := fmt.Sprintf("%s %s", seniority, title)

			location := locations[rng.Intn(len(locations))]
			workArrangement := "onsite"
			if location == "Remote" {
				workArrangement = "remote"
			} else if rng.Intn(3) == 0 {
				workArrangement = "hybrid"
			}

			employmentTypes := []string{"full_time", "full_time", "full_time", "contract", "part_time"}
			employmentType := employmentTypes[rng.Intn(len(employmentTypes))]

			baseSalary := 70_000 + rng.Intn(60_000)
			seniorityMultiplier := map[string]float64{"Junior": 0.75, "Mid": 1.0, "Senior": 1.35, "Lead": 1.7}[seniority]
			salaryMin := int(float64(baseSalary) * seniorityMultiplier)
			salaryMax := salaryMin + 15_000 + rng.Intn(25_000)

			mustHave := pickSkills(rng, cat.MustHave, 2+rng.Intn(2))
			niceToHave := pickSkills(rng, cat.NiceToHave, 1+rng.Intn(3))

			// Spread publish dates across the last 120 days so "posted Xd/mo
			// ago" and recency-sort both show real variety.
			publishedAt := now.Add(-time.Duration(rng.Intn(120*24)) * time.Hour)

			description := fmt.Sprintf(
				"We're looking for a %s %s to join our %s team. You'll work closely with cross-functional partners to ship impactful work.",
				seniority, title, cat.Industry,
			)

			_, err := pool.Exec(ctx, `
				INSERT INTO jobs (
					company_id, created_by, title, description, location, employment_type,
					work_arrangement, category, salary_min, salary_max, must_have_skills,
					nice_to_have_skills, seniority, status, published_at
				) VALUES (
					$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, 'published', $14
				)
			`, companyID, recruiterID, fullTitle, description, location, employmentType,
				workArrangement, string(cat.Value), salaryMin, salaryMax, mustHave,
				niceToHave, seniority, publishedAt,
			)
			if err != nil {
				return fmt.Errorf("inserting job %q: %w", fullTitle, err)
			}
		}
	}

	return nil
}

// pickSkills returns up to n distinct random skills from pool.
func pickSkills(rng *rand.Rand, pool []string, n int) []string {
	if n > len(pool) {
		n = len(pool)
	}
	indices := rng.Perm(len(pool))[:n]
	out := make([]string, 0, n)
	for _, idx := range indices {
		out = append(out, pool[idx])
	}
	return out
}
