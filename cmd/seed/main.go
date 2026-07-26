// Command seed populates the database with enough dummy data to try every
// flow immediately: a super admin, an approved recruiter with sample jobs,
// and a candidate with a couple of applications.
//
// Usage: go run ./cmd/seed
package main

import (
	"log"
	"os"
	"time"

	"jobhoo/internal/db"
	"jobhoo/internal/models"
	"jobhoo/internal/repository"
	"jobhoo/internal/services/aimatching"
	authsvc "jobhoo/internal/services/auth"
)

func main() {
	conn, err := db.Connect()
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer conn.Close()

	migrationsDir := getenv("MIGRATIONS_DIR", "migrations")
	if err := db.Migrate(conn, migrationsDir); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	users := repository.NewUserRepo(conn)
	jobs := repository.NewJobRepo(conn)
	applications := repository.NewApplicationRepo(conn)
	ai := aimatching.NewProvider()

	uploadDir := getenv("UPLOAD_DIR", "uploads")
	os.MkdirAll(uploadDir+"/resumes", 0o755)
	os.MkdirAll(uploadDir+"/documents", 0o755)

	dummyResumePath := uploadDir + "/resumes/seed-resume.pdf"
	dummyDocPath := uploadDir + "/documents/seed-document.pdf"
	writeIfMissing(dummyResumePath, "%PDF-1.4 seed resume\n")
	writeIfMissing(dummyDocPath, "%PDF-1.4 seed supporting document\n")

	// --- Super Admin ---
	adminID := ensureUser(users, "admin@jobhoo.com", "admin12345", models.RoleSuperAdmin)
	log.Printf("super admin: admin@jobhoo.com / admin12345 (id=%d)", adminID)

	// --- Approved recruiter ---
	recruiterID := ensureUser(users, "recruiter@jobhoo.com", "recruiter123", models.RoleRecruiter)
	if _, err := users.GetRecruiterProfile(recruiterID); err == repository.ErrNotFound {
		_ = users.CreateRecruiterProfile(recruiterID, "PT Jobhoo Teknologi Indonesia", "/uploads/documents/seed-document.pdf", "seed-document.pdf")
		_ = users.UpdateRecruiterStatus(recruiterID, adminID, models.RecruiterApproved)
	}
	log.Printf("recruiter (approved): recruiter@jobhoo.com / recruiter123 (id=%d)", recruiterID)

	// --- Pending recruiter (to demo the approval flow) ---
	pendingRecruiterID := ensureUser(users, "pending-recruiter@jobhoo.com", "recruiter123", models.RoleRecruiter)
	if _, err := users.GetRecruiterProfile(pendingRecruiterID); err == repository.ErrNotFound {
		_ = users.CreateRecruiterProfile(pendingRecruiterID, "PT Kandidat Approval", "/uploads/documents/seed-document.pdf", "seed-document.pdf")
	}
	log.Printf("recruiter (pending): pending-recruiter@jobhoo.com / recruiter123 (id=%d)", pendingRecruiterID)

	// --- Candidate ---
	candidateID := ensureUser(users, "candidate@jobhoo.com", "candidate123", models.RoleCandidate)
	if _, err := users.GetCandidateProfile(candidateID); err == repository.ErrNotFound {
		_ = users.CreateCandidateProfile(candidateID, "Andi Wijaya", "/uploads/resumes/seed-resume.pdf", "seed-resume.pdf")
	}
	log.Printf("candidate: candidate@jobhoo.com / candidate123 (id=%d)", candidateID)

	// --- Sample jobs ---
	closing := time.Now().AddDate(0, 1, 0)
	salaryMin1, salaryMax1 := int64(8000000), int64(15000000)
	job1, err := createJobIfNotExists(jobs, recruiterID, "Backend Engineer (Go)", repository.JobInput{
		Title: "Backend Engineer (Go)", Position: "Backend Engineer",
		EmploymentType: models.FullTime, WorkArrangement: models.Remote,
		Location: "Jakarta (Remote)", SalaryMin: &salaryMin1, SalaryMax: &salaryMax1,
		Benefits:     "BPJS Kesehatan & Ketenagakerjaan, remote allowance, laptop.",
		Requirements: "Minimal 2 tahun pengalaman Go, familiar dengan PostgreSQL dan REST API.",
		Description:  "Membangun dan memelihara layanan backend untuk platform JOBHOO.",
		ClosingDate:  closing,
	})
	if err != nil {
		log.Printf("job1: %v", err)
	}

	salaryMin2, salaryMax2 := int64(6000000), int64(10000000)
	job2, err := createJobIfNotExists(jobs, recruiterID, "Frontend Developer (Next.js)", repository.JobInput{
		Title: "Frontend Developer (Next.js)", Position: "Frontend Developer",
		EmploymentType: models.FullTime, WorkArrangement: models.Hybrid,
		Location: "Yogyakarta", SalaryMin: &salaryMin2, SalaryMax: &salaryMax2,
		Benefits:     "BPJS, jam kerja fleksibel.",
		Requirements: "Pengalaman React/Next.js, TypeScript, Tailwind CSS.",
		Description:  "Mengembangkan antarmuka pengguna JOBHOO yang cepat dan minimalis.",
		ClosingDate:  closing,
	})
	if err != nil {
		log.Printf("job2: %v", err)
	}

	_, err = createJobIfNotExists(jobs, recruiterID, "UI/UX Designer", repository.JobInput{
		Title: "UI/UX Designer", Position: "Product Designer",
		EmploymentType: models.Contract, WorkArrangement: models.Remote,
		Location: "Remote (Indonesia)",
		Benefits:     "Fleksibel, proyek jangka pendek 3 bulan.",
		Requirements: "Portfolio kuat di Figma, pengalaman design system.",
		Description:  "Merancang pengalaman pengguna untuk fitur ATS Kanban JOBHOO.",
		ClosingDate:  closing,
	})
	if err != nil {
		log.Printf("job3: %v", err)
	}

	// --- Sample application with AI match result ---
	if job1 != 0 {
		applied, _ := applications.HasApplied(job1, candidateID)
		if !applied {
			appID, err := applications.Create(job1, candidateID, "/uploads/resumes/seed-resume.pdf")
			if err == nil {
				result, _ := ai.Analyze("/uploads/resumes/seed-resume.pdf", "Membangun layanan backend", "Go, PostgreSQL, REST API")
				_ = applications.SetAIMatchResult(appID, result.MatchScore, result.SkillMatch, result.ExperienceMatch, result.EducationMatch)
				_ = applications.UpdateStage(appID, models.StageResumeReviewed)
			}
		}
	}
	if job2 != 0 {
		applied, _ := applications.HasApplied(job2, candidateID)
		if !applied {
			appID, err := applications.Create(job2, candidateID, "/uploads/resumes/seed-resume.pdf")
			if err == nil {
				result, _ := ai.Analyze("/uploads/resumes/seed-resume.pdf", "Mengembangkan antarmuka", "React, Next.js, TypeScript")
				_ = applications.SetAIMatchResult(appID, result.MatchScore, result.SkillMatch, result.ExperienceMatch, result.EducationMatch)
			}
		}
	}

	log.Println("seed complete")
}

func ensureUser(users *repository.UserRepo, email, password string, role models.UserRole) int64 {
	if u, err := users.GetByEmail(email); err == nil {
		return u.ID
	}
	hash, err := authsvc.HashPassword(password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}
	id, err := users.CreateUser(email, hash, role)
	if err != nil {
		log.Fatalf("create user %s: %v", email, err)
	}
	return id
}

func createJobIfNotExists(jobs *repository.JobRepo, recruiterID int64, title string, input repository.JobInput) (int64, error) {
	existing, err := jobs.ListByRecruiter(recruiterID)
	if err == nil {
		for _, j := range existing {
			if j.Title == title {
				return j.ID, nil
			}
		}
	}
	return jobs.Create(recruiterID, input)
}

func writeIfMissing(path, content string) {
	if _, err := os.Stat(path); err == nil {
		return
	}
	_ = os.WriteFile(path, []byte(content), 0o644)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
