package repository

import (
	"database/sql"

	"jobhoo/internal/models"
)

type ApplicationRepo struct {
	db *sql.DB
}

func NewApplicationRepo(db *sql.DB) *ApplicationRepo {
	return &ApplicationRepo{db: db}
}

func (r *ApplicationRepo) Create(jobID, candidateID int64, resumePath string) (int64, error) {
	var id int64
	err := r.db.QueryRow(
		`INSERT INTO applications (job_id, candidate_id, stage, resume_path) VALUES ($1, $2, 'applied', $3) RETURNING id`,
		jobID, candidateID, resumePath,
	).Scan(&id)
	return id, err
}

func (r *ApplicationRepo) HasApplied(jobID, candidateID int64) (bool, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM applications WHERE job_id = $1 AND candidate_id = $2`, jobID, candidateID).Scan(&n)
	return n > 0, err
}

func (r *ApplicationRepo) SetAIMatchResult(id int64, matchScore, skillMatch, experienceMatch, educationMatch int) error {
	_, err := r.db.Exec(
		`UPDATE applications SET match_score=$1, skill_match=$2, experience_match=$3, education_match=$4, updated_at=now() WHERE id=$5`,
		matchScore, skillMatch, experienceMatch, educationMatch, id,
	)
	return err
}

func (r *ApplicationRepo) UpdateStage(id int64, stage models.ApplicationStage) error {
	_, err := r.db.Exec(`UPDATE applications SET stage=$1, updated_at=now() WHERE id=$2`, stage, id)
	return err
}

func (r *ApplicationRepo) SetFinalStatus(id int64, status models.ApplicationFinalStatus) error {
	_, err := r.db.Exec(`UPDATE applications SET final_status=$1, updated_at=now() WHERE id=$2`, status, id)
	return err
}

func (r *ApplicationRepo) GetByID(id int64) (*models.Application, error) {
	a := &models.Application{}
	err := r.db.QueryRow(
		`SELECT a.id, a.job_id, a.candidate_id, a.stage, a.final_status, a.match_score, a.skill_match,
		        a.experience_match, a.education_match, a.resume_path, a.created_at, a.updated_at,
		        j.title, j.status
		 FROM applications a JOIN jobs j ON j.id = a.job_id
		 WHERE a.id = $1`, id,
	).Scan(&a.ID, &a.JobID, &a.CandidateID, &a.Stage, &a.FinalStatus, &a.MatchScore, &a.SkillMatch,
		&a.ExperienceMatch, &a.EducationMatch, &a.ResumePath, &a.CreatedAt, &a.UpdatedAt, &a.JobTitle, &a.JobStatus)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return a, err
}

// ListByCandidate returns a candidate's application history, most recent first.
func (r *ApplicationRepo) ListByCandidate(candidateID int64) ([]models.Application, error) {
	return r.ListByCandidatePaginated(candidateID, 1000, 0)
}

func (r *ApplicationRepo) ListByCandidatePaginated(candidateID int64, limit, offset int) ([]models.Application, error) {
	rows, err := r.db.Query(
		`SELECT a.id, a.job_id, a.candidate_id, a.stage, a.final_status, a.match_score, a.skill_match,
		        a.experience_match, a.education_match, a.resume_path, a.created_at, a.updated_at,
		        j.title, j.status
		 FROM applications a JOIN jobs j ON j.id = a.job_id
		 WHERE a.candidate_id = $1 ORDER BY a.created_at DESC
		 LIMIT $2 OFFSET $3`, candidateID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanApplications(rows)
}

func (r *ApplicationRepo) CountByCandidate(candidateID int64) (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM applications WHERE candidate_id = $1`, candidateID).Scan(&n)
	return n, err
}

// ListByJob returns all applicants for a job, used by the ATS kanban board.
func (r *ApplicationRepo) ListByJob(jobID int64) ([]models.Application, error) {
	rows, err := r.db.Query(
		`SELECT a.id, a.job_id, a.candidate_id, a.stage, a.final_status, a.match_score, a.skill_match,
		        a.experience_match, a.education_match, a.resume_path, a.created_at, a.updated_at,
		        j.title, j.status, u.email, cp.full_name
		 FROM applications a
		 JOIN jobs j ON j.id = a.job_id
		 JOIN users u ON u.id = a.candidate_id
		 LEFT JOIN candidate_profiles cp ON cp.user_id = a.candidate_id
		 WHERE a.job_id = $1 ORDER BY a.created_at DESC`, jobID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Application
	for rows.Next() {
		var a models.Application
		if err := rows.Scan(&a.ID, &a.JobID, &a.CandidateID, &a.Stage, &a.FinalStatus, &a.MatchScore, &a.SkillMatch,
			&a.ExperienceMatch, &a.EducationMatch, &a.ResumePath, &a.CreatedAt, &a.UpdatedAt, &a.JobTitle, &a.JobStatus,
			&a.CandidateEmail, &a.CandidateName); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ListByRecruiter returns recent applications across all of a recruiter's jobs.
func (r *ApplicationRepo) ListByRecruiter(recruiterID int64, limit int) ([]models.Application, error) {
	rows, err := r.db.Query(
		`SELECT a.id, a.job_id, a.candidate_id, a.stage, a.final_status, a.match_score, a.skill_match,
		        a.experience_match, a.education_match, a.resume_path, a.created_at, a.updated_at,
		        j.title, j.status, u.email, cp.full_name
		 FROM applications a
		 JOIN jobs j ON j.id = a.job_id
		 JOIN users u ON u.id = a.candidate_id
		 LEFT JOIN candidate_profiles cp ON cp.user_id = a.candidate_id
		 WHERE j.recruiter_id = $1
		 ORDER BY a.created_at DESC LIMIT $2`, recruiterID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Application
	for rows.Next() {
		var a models.Application
		if err := rows.Scan(&a.ID, &a.JobID, &a.CandidateID, &a.Stage, &a.FinalStatus, &a.MatchScore, &a.SkillMatch,
			&a.ExperienceMatch, &a.EducationMatch, &a.ResumePath, &a.CreatedAt, &a.UpdatedAt, &a.JobTitle, &a.JobStatus,
			&a.CandidateEmail, &a.CandidateName); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// CountApplicantsByJob returns applicant counts keyed by job id, for the recruiter job list.
func (r *ApplicationRepo) CountApplicantsByJob(recruiterID int64) (map[int64]int, error) {
	rows, err := r.db.Query(
		`SELECT j.id, COUNT(a.id) FROM jobs j
		 LEFT JOIN applications a ON a.job_id = j.id
		 WHERE j.recruiter_id = $1 GROUP BY j.id`, recruiterID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int64]int{}
	for rows.Next() {
		var id int64
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}
		out[id] = count
	}
	return out, rows.Err()
}

func (r *ApplicationRepo) CountByRecruiter(recruiterID int64) (int, error) {
	var n int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM applications a JOIN jobs j ON j.id = a.job_id WHERE j.recruiter_id = $1`, recruiterID,
	).Scan(&n)
	return n, err
}

// ApplicationsWithoutFinalStatus finds applications on closed jobs missing a final status,
// used to enforce the business rule that no application on a closed job may lack one.
func (r *ApplicationRepo) ApplicationsWithoutFinalStatus(jobID int64) ([]int64, error) {
	rows, err := r.db.Query(`SELECT id FROM applications WHERE job_id = $1 AND final_status IS NULL`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func scanApplications(rows *sql.Rows) ([]models.Application, error) {
	var out []models.Application
	for rows.Next() {
		var a models.Application
		if err := rows.Scan(&a.ID, &a.JobID, &a.CandidateID, &a.Stage, &a.FinalStatus, &a.MatchScore, &a.SkillMatch,
			&a.ExperienceMatch, &a.EducationMatch, &a.ResumePath, &a.CreatedAt, &a.UpdatedAt, &a.JobTitle, &a.JobStatus); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
