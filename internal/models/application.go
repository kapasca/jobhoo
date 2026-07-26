package models

import "time"

type ApplicationStage string

const (
	StageApplied        ApplicationStage = "applied"
	StageResumeReviewed ApplicationStage = "resume_reviewed"
	StageInterview      ApplicationStage = "interview"
	StageOffered        ApplicationStage = "offered"
	StageHired          ApplicationStage = "hired"
	StageRejected       ApplicationStage = "rejected"
)

// ATSStages defines the kanban column order shown to recruiters.
var ATSStages = []ApplicationStage{
	StageApplied,
	StageResumeReviewed,
	StageInterview,
	StageOffered,
	StageHired,
	StageRejected,
}

func StageLabel(s ApplicationStage) string {
	switch s {
	case StageApplied:
		return "Applied"
	case StageResumeReviewed:
		return "Resume Reviewed"
	case StageInterview:
		return "Interview"
	case StageOffered:
		return "Offered"
	case StageHired:
		return "Hired"
	case StageRejected:
		return "Rejected"
	default:
		return string(s)
	}
}

type ApplicationFinalStatus string

const (
	FinalProceedNextStage  ApplicationFinalStatus = "proceed_to_next_stage"
	FinalResumeNotReviewed ApplicationFinalStatus = "resume_not_reviewed"
	FinalNotMatched        ApplicationFinalStatus = "not_matched"
)

func FinalStatusLabel(s ApplicationFinalStatus) string {
	switch s {
	case FinalProceedNextStage:
		return "Proceed to Next Stage"
	case FinalResumeNotReviewed:
		return "Resume Not Reviewed"
	case FinalNotMatched:
		return "Not Matched"
	default:
		return string(s)
	}
}

type Application struct {
	ID              int64
	JobID           int64
	CandidateID     int64
	Stage           ApplicationStage
	FinalStatus     *ApplicationFinalStatus
	MatchScore      *int
	SkillMatch      *int
	ExperienceMatch *int
	EducationMatch  *int
	ResumePath      string
	CreatedAt       time.Time
	UpdatedAt       time.Time

	// Populated via joins for display purposes
	JobTitle        string
	JobStatus       JobStatus
	CandidateEmail  string
	CandidateName   string
}
