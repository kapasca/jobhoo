package models

import "time"

type UserRole string

const (
	RoleCandidate   UserRole = "candidate"
	RoleRecruiter   UserRole = "recruiter"
	RoleSuperAdmin  UserRole = "super_admin"
)

type RecruiterStatus string

const (
	RecruiterPending  RecruiterStatus = "pending"
	RecruiterApproved RecruiterStatus = "approved"
	RecruiterRejected RecruiterStatus = "rejected"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         UserRole
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CandidateProfile struct {
	ID             int64
	UserID         int64
	FullName       string
	ResumePath     string
	ResumeFilename string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type RecruiterProfile struct {
	ID               int64
	UserID           int64
	CompanyName      string
	DocumentPath     string
	DocumentFilename string
	Status           RecruiterStatus
	ReviewedBy       *int64
	ReviewedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// AuthUser is the combined view used once a user is authenticated,
// convenient for templates and middleware checks.
type AuthUser struct {
	ID    int64
	Email string
	Role  UserRole
	// RecruiterStatus is only meaningful when Role == RoleRecruiter
	RecruiterStatus RecruiterStatus
}
