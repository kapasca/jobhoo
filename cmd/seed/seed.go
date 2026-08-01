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
			INSERT INTO users (email, password_hash, role, full_name, email_verified)
			VALUES ($1, $2, 'recruiter', $3, true)
			RETURNING id
		`, email, passwordHash, fullName).Scan(&recruiterID)
		if err != nil {
			return nil, nil, fmt.Errorf("inserting recruiter %s: %w", email, err)
		}

		var companyID string
		// Leave the very last demo company 'pending' so there's always
		// something in the admin approval queue to test against; every
		// other demo company is auto-approved so the existing demo
		// recruiter accounts keep working exactly as they did before the
		// approval workflow existed.
		status := "approved"
		if i == numCompanies-1 {
			status = "pending"
		}
		var approvedAt *time.Time
		if status == "approved" {
			t := time.Now()
			approvedAt = &t
		}
		err = pool.QueryRow(ctx, `
			INSERT INTO companies (owner_id, name, industry, description, status, approved_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, recruiterID, c.Name, c.Industry, fmt.Sprintf("%s is hiring across %s.", c.Name, c.Industry), status, approvedAt).Scan(&companyID)
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

			location := seedLocations[rng.Intn(len(seedLocations))]
			workArrangement := "onsite"
			if rng.Intn(4) == 0 {
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
					company_id, created_by, title, description, country, state, employment_type,
					work_arrangement, category, salary_min, salary_max, salary_currency,
					must_have_skills, nice_to_have_skills, seniority, status, published_at
				) VALUES (
					$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, 'published', $16
				)
			`, companyID, recruiterID, fullTitle, description, location.Country, location.State, employmentType,
				workArrangement, string(cat.Value), salaryMin, salaryMax, countryCurrency(location.Country),
				mustHave, niceToHave, seniority, publishedAt,
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

var seedCurrencies = map[string]string{
	"Indonesia": "IDR", "Singapore": "SGD", "Malaysia": "MYR",
	"Thailand": "THB", "Vietnam": "VND", "Philippines": "PHP",
	"Timor-Leste": "USD", "Australia": "AUD", "New Zealand": "NZD",
}

func countryCurrency(country string) string {
	if c, ok := seedCurrencies[country]; ok {
		return c
	}
	return "USD"
}

var candidateNames = []string{
	"Alice Johnson", "Bob Smith", "Carol Williams", "David Brown", "Emily Davis",
	"Frank Miller", "Grace Wilson", "Henry Moore", "Iris Taylor", "Jack Anderson",
	"Kate Thomas", "Leo Jackson", "Mia White", "Nathan Harris", "Olivia Martin",
	"Peter Thompson", "Quinn Garcia", "Rachel Rodriguez", "Samuel Lee", "Tina Martinez",
}

var candidateSkills = [][]string{
	{"React", "JavaScript", "CSS", "HTML"},
	{"Python", "Data Analysis", "SQL", "Tableau"},
	{"Figma", "Design Systems", "Prototyping", "UI Design"},
	{"Go", "Kubernetes", "AWS", "Docker"},
	{"Product Management", "User Research", "Analytics", "Roadmapping"},
	{"Salesforce", "CRM", "Negotiation", "Lead Generation"},
	{"Java", "Spring Boot", "SQL", "REST APIs"},
	{"Machine Learning", "Python", "TensorFlow", "Data Science"},
	{"DevOps", "Terraform", "CI/CD", "Linux"},
	{"Product Design", "UX Research", "Figma", "Design Thinking"},
}

// seedCandidates creates numCandidates candidate accounts. About 70% have
// email verified, 30% are unverified (for testing admin dashboard filtering).
func seedCandidates(ctx context.Context, pool *pgxpool.Pool, rng *rand.Rand) ([]string, error) {
	passwordHash, err := auth.HashPassword("demo-password-123")
	if err != nil {
		return nil, fmt.Errorf("hashing demo password: %w", err)
	}

	candidateIDs := []string{}
	for i := 0; i < numCandidates; i++ {
		email := fmt.Sprintf("candidate%d%s", i+1, demoEmailDomain)
		fullName := candidateNames[i%len(candidateNames)]

		var candidateID string
		err := pool.QueryRow(ctx, `
			INSERT INTO users (email, password_hash, role, full_name, email_verified)
			VALUES ($1, $2, 'candidate', $3, true)
			RETURNING id
		`, email, passwordHash, fullName).Scan(&candidateID)
		if err != nil {
			return nil, fmt.Errorf("inserting candidate %s: %w", email, err)
		}

		// Create candidate profile with skills
		skills := candidateSkills[i%len(candidateSkills)]
		location := seedLocations[rng.Intn(len(seedLocations))]
		headline := "Experienced professional seeking new opportunities"

		_, err = pool.Exec(ctx, `
			INSERT INTO candidate_profiles (user_id, headline, skills, location)
			VALUES ($1, $2, $3, $4)
		`, candidateID, headline, skills, location.Country)
		if err != nil {
			return nil, fmt.Errorf("inserting candidate profile for %s: %w", email, err)
		}

		candidateIDs = append(candidateIDs, candidateID)
	}

	return candidateIDs, nil
}

// seedApplications creates applications from random candidates to random jobs,
// with varied stages (applied, screening, interview, offer, hired, rejected).
// Each candidate applies to 3-7 random jobs.
func seedApplications(ctx context.Context, pool *pgxpool.Pool, rng *rand.Rand, candidateIDs []string) error {
	stages := []string{"applied", "screening", "interview", "offer", "hired", "rejected"}
	stageWeights := []int{40, 25, 20, 10, 3, 2} // "applied" is most common

	// Get all job IDs
	rows, err := pool.Query(ctx, `SELECT id FROM jobs ORDER BY id`)
	if err != nil {
		return fmt.Errorf("querying jobs: %w", err)
	}
	defer rows.Close()

	jobIDs := []string{}
	for rows.Next() {
		var jobID string
		if err := rows.Scan(&jobID); err != nil {
			return err
		}
		jobIDs = append(jobIDs, jobID)
	}

	if len(jobIDs) == 0 {
		return fmt.Errorf("no jobs found to create applications for")
	}

	// Each candidate applies to 3-7 random jobs
	for _, candidateID := range candidateIDs {
		numApplications := 3 + rng.Intn(5) // 3-7
		for j := 0; j < numApplications; j++ {
			jobID := jobIDs[rng.Intn(len(jobIDs))]

			// Pick stage based on weights
			stageIdx := 0
			randVal := rng.Intn(100)
			cumulative := 0
			for i, weight := range stageWeights {
				cumulative += weight
				if randVal < cumulative {
					stageIdx = i
					break
				}
			}
			stage := stages[stageIdx]

			_, err := pool.Exec(ctx, `
				INSERT INTO applications (candidate_id, job_id, stage)
				VALUES ($1, $2, $3)
				ON CONFLICT (candidate_id, job_id) DO NOTHING
			`, candidateID, jobID, stage)
			if err != nil {
				return fmt.Errorf("inserting application from candidate %s to job %s: %w", candidateID, jobID, err)
			}
		}
	}

	return nil
}
