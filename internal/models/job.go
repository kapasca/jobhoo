package models

import "time"

type JobStatus string

const (
	JobOpen   JobStatus = "open"
	JobClosed JobStatus = "closed"
)

type EmploymentType string

const (
	FullTime   EmploymentType = "full_time"
	PartTime   EmploymentType = "part_time"
	Contract   EmploymentType = "contract"
	Internship EmploymentType = "internship"
	Freelance  EmploymentType = "freelance"
)

type WorkArrangement string

const (
	Onsite WorkArrangement = "onsite"
	Remote WorkArrangement = "remote"
	Hybrid WorkArrangement = "hybrid"
)

type Job struct {
	ID              int64
	RecruiterID     int64
	Title           string
	Position        string
	EmploymentType  EmploymentType
	WorkArrangement WorkArrangement
	Location        string
	SalaryMin       *int64
	SalaryMax       *int64
	Benefits        string
	Requirements    string
	Description     string
	ClosingDate     time.Time
	Status          JobStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time

	// Populated via joins for display purposes
	RecruiterCompany string
	ApplicantCount   int
}

func EmploymentTypeLabel(t EmploymentType) string {
	switch t {
	case FullTime:
		return "Full Time"
	case PartTime:
		return "Part Time"
	case Contract:
		return "Contract"
	case Internship:
		return "Internship"
	case Freelance:
		return "Freelance"
	default:
		return string(t)
	}
}

func WorkArrangementLabel(w WorkArrangement) string {
	switch w {
	case Onsite:
		return "On-site"
	case Remote:
		return "Remote"
	case Hybrid:
		return "Hybrid"
	default:
		return string(w)
	}
}
