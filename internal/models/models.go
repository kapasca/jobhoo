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
	CreatedAt    time.Time
}

type Company struct {
	ID          string
	OwnerID     string
	Name        string
	LogoURL     string
	Website     string
	Description string
	Industry    string
	CreatedAt   time.Time
}

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
	Hybrid WorkArrangement = "hybrid"
	Remote WorkArrangement = "remote"
)

type JobStatus string

const (
	JobDraft     JobStatus = "draft"
	JobPublished JobStatus = "published"
	JobClosed    JobStatus = "closed"
	JobArchived  JobStatus = "archived"
)

type JobCategory string

const (
	CategoryEngineeringProduct JobCategory = "engineering_product"
	CategoryDesignCreative     JobCategory = "design_creative"
	CategorySalesMarketing     JobCategory = "sales_marketing"
	CategoryDataAnalytics      JobCategory = "data_analytics"
	CategoryOperationsSupport  JobCategory = "operations_support"
)

// JobCategories lists every category in display order, with a human label
// for use in filter chips / nav. Kept here (not in the DB) so the display
// label and canonical order are versioned with the code, not the data.
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

func (c JobCategory) Label() string {
	for _, entry := range JobCategories {
		if entry.Value == c {
			return entry.Label
		}
	}
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
	Location         string
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
	PublishedAt      *time.Time
	CreatedAt        time.Time
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
