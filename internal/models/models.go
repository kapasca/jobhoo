// Package models defines JOBHOO's core domain types. These mirror the
// database schema (internal/database/migrations) and are the shared
// vocabulary between the database layer, handlers, and templates.
package models

import "time"

type UserRole string

const (
	RoleCandidate UserRole = "candidate"
	RoleRecruiter UserRole = "recruiter"
	RoleAdmin     UserRole = "admin"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string `json:"-"`
	Role         UserRole
	FullName     string
	AvatarURL    string
	IsFrozen     bool
	CreatedAt    time.Time
}

type CompanyStatus string

const (
	CompanyPending     CompanyStatus = "pending"
	CompanyApproved    CompanyStatus = "approved"
	CompanyRejected    CompanyStatus = "rejected"
	CompanyBlacklisted CompanyStatus = "blacklisted"
)

type Company struct {
	ID              string
	OwnerID         string
	Name            string
	LogoURL         string
	Website         string
	Description     string
	Industry        string
	Status          CompanyStatus
	ApprovedAt      *time.Time
	ApprovedBy      string
	RejectionReason string
	CreatedAt       time.Time
}

// IsProfileComplete reports whether the company has filled in the fields
// required before job posting is permitted (industry + description).
func (c Company) IsProfileComplete() bool {
	return c.Industry != "" && c.Description != ""
}

type EmploymentType string

const (
	FullTime   EmploymentType = "full_time"
	PartTime   EmploymentType = "part_time"
	Contract   EmploymentType = "contract"
	Internship EmploymentType = "internship"
	Freelance  EmploymentType = "freelance"
)

func (e EmploymentType) String() string {
	switch e {
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
		return string(e)
	}
}

type WorkArrangement string

const (
	Onsite WorkArrangement = "onsite"
	Hybrid WorkArrangement = "hybrid"
	Remote WorkArrangement = "remote"
)

func (w WorkArrangement) String() string {
	switch w {
	case Onsite:
		return "Onsite"
	case Hybrid:
		return "Hybrid"
	case Remote:
		return "Remote"
	default:
		return string(w)
	}
}

type JobStatus string

const (
	JobDraft     JobStatus = "draft"
	JobPublished JobStatus = "published"
	JobClosed    JobStatus = "closed"
	JobArchived  JobStatus = "archived"
)

type JobCategory string

const (
	CategoryEngineeringProduct JobCategory = "Engineering & Product"
	CategoryDesignCreative     JobCategory = "Design & Creative"
	CategorySalesMarketing     JobCategory = "Sales & Marketing"
	CategoryDataAnalytics      JobCategory = "Data & Analytics"
	CategoryOperationsSupport  JobCategory = "Operations & Support"
)

// JobCategories lists every standard category in display order.
var JobCategories = []struct {
	Value JobCategory
	Label string
}{
	{CategoryEngineeringProduct, "Engineering & Product"},
	{CategoryDesignCreative, "Design & Creative"},
	{CategorySalesMarketing, "Sales & Marketing"},
	{CategoryDataAnalytics, "Data & Analytics"},
	{CategoryOperationsSupport, "Operations & Support"},
}

// Label returns the human-readable category name; for custom categories it
// just returns the stored string as-is.
func (c JobCategory) Label() string {
	return string(c)
}

type Job struct {
	ID               string
	CompanyID        string
	CompanyName      string  // denormalized for card rendering
	CompanyLogoURL   *string // denormalized for card rendering; nullable since logo_url can be NULL in companies table
	CreatedBy        string
	Title            string
	Description      string
	Country          string
	State            string
	EmploymentType   EmploymentType
	WorkArrangement  WorkArrangement
	Category         JobCategory
	SalaryMin        *int
	SalaryMax        *int
	SalaryCurrency   string
	MustHaveSkills   []string
	NiceToHaveSkills []string
	Seniority        string
	Status           JobStatus
	IsSaved          bool       // true when the current user has bookmarked this job
	IsFrozen         bool       // true when an admin has frozen this job
	OpensAt          *time.Time // if set, hidden from public listing until this time
	ClosesAt         *time.Time // if set, hidden from public listing after this time
	PublishedAt      *time.Time
	CreatedAt        time.Time
}

// IsScheduled reports whether the job has an opens_at in the future — used
// by the recruiter dashboard to show "Scheduled" instead of "Published".
func (j Job) IsScheduled() bool {
	return j.OpensAt != nil && j.OpensAt.After(time.Now())
}

// Location returns a formatted display string combining state and country.
func (j Job) Location() string {
	if j.State == "" || j.State == j.Country {
		return j.Country
	}
	return j.State + ", " + j.Country
}

// IsExpired reports whether the job's closes_at has passed.
func (j Job) IsExpired() bool {
	return j.ClosesAt != nil && j.ClosesAt.Before(time.Now())
}

// PostedAgo returns a short human string like "3d ago" for job cards.
func (j Job) PostedAgo() string {
	if j.PublishedAt == nil {
		return "Draft"
	}
	d := time.Since(*j.PublishedAt)
	switch {
	case d < time.Hour:
		return "Just now"
	case d < 24*time.Hour:
		return itoa(int(d.Hours())) + "h ago"
	case d < 30*24*time.Hour:
		return itoa(int(d.Hours()/24)) + "d ago"
	default:
		return itoa(int(d.Hours()/24/30)) + "mo ago"
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}

type ApplicationStage string

const (
	StageApplied   ApplicationStage = "applied"
	StageScreening ApplicationStage = "screening"
	StageInterview ApplicationStage = "interview"
	StageOffer     ApplicationStage = "offer"
	StageHired     ApplicationStage = "hired"
	StageRejected  ApplicationStage = "rejected"
)

// OrderedStages defines the canonical left-to-right pipeline order used by
// the ATS board. Rejected is tracked but shown separately, not as a column
// candidates "graduate" into.
var OrderedStages = []ApplicationStage{
	StageApplied, StageScreening, StageInterview, StageOffer, StageHired,
}

// StageLabels gives a human-readable label for each stage, used by the ATS
// board columns and application status badges.
var StageLabels = map[ApplicationStage]string{
	StageApplied:   "Applied",
	StageScreening: "Screening",
	StageInterview: "Interview",
	StageOffer:     "Offer",
	StageHired:     "Hired",
	StageRejected:  "Rejected",
}

func (s ApplicationStage) Label() string {
	if label, ok := StageLabels[s]; ok {
		return label
	}
	return string(s)
}

type Application struct {
	ID          string
	JobID       string
	CandidateID string
	Stage       ApplicationStage
	CoverNote   string
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Populated via joins for display purposes.
	CandidateName   string
	CandidateEmail  string
	CandidateSkills []string
	JobTitle        string
	CompanyName     string
	Location        string
}

type CandidateProfile struct {
	UserID        string
	Headline      string
	ResumeText    string
	ResumeFileURL string
	Skills        []string
	Location      string
	UpdatedAt     time.Time
}
