package handlers

import (
	"net/http"
	"strconv"

	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/middleware"
	"github.com/jobhoo/jobhoo/internal/models"
)

const logPageSize = 20

func parsePage(q string) int {
	n, _ := strconv.Atoi(q)
	if n < 1 {
		return 1
	}
	return n
}

func totalPages(total, pageSize int) int {
	if total == 0 {
		return 1
	}
	p := (total + pageSize - 1) / pageSize
	if p < 1 {
		return 1
	}
	return p
}

type candidateDashboardData struct {
	BasePageData
	Applications []models.Application
	SavedJobs    []models.Job
	HasProfile   bool
}

func (h *Handlers) CandidateDashboard(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	applications, err := h.Applications.ListByCandidate(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "could not load applications", http.StatusInternalServerError)
		return
	}
	saved, err := h.SavedJobs.ListByCandidate(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "could not load saved jobs", http.StatusInternalServerError)
		return
	}
	// All jobs in the saved list are by definition saved.
	for i := range saved {
		saved[i].IsSaved = true
	}
	_, profileErr := h.Profiles.GetByUserID(r.Context(), user.ID)

	h.Render.Render(w, http.StatusOK, "candidate-dashboard.html", candidateDashboardData{
		BasePageData: newBasePageData(r, "dashboard"),
		Applications: applications,
		SavedJobs:    saved,
		HasProfile:   profileErr == nil,
	})
}

type adminDashboardData struct {
	BasePageData
	TotalUsers          int
	TotalCandidates     int
	TotalRecruiters     int
	TotalCompanies      int
	TotalJobs           int
	TotalApplications   int
	PendingCompanyCount int
	ActiveTab           string
	// users tab
	UserLogs       []database.UserRow
	UserPage       int
	UserTotalPages int
	// candidates tab
	CandidateLogs       []database.CandidateRegistration
	CandidatePage       int
	CandidateTotalPages int
	// recruiters tab (recruiter user list)
	RecruiterLogs       []database.CandidateRegistration
	RecruiterPage       int
	RecruiterTotalPages int
	// companies tab
	CompanyLogs       []database.CompanyWithOwner
	CompanyPage       int
	CompanyTotalPages int
	// jobs tab
	JobLogs       []database.AdminJobRow
	JobPage       int
	JobTotalPages int
	// applications tab
	AppLogs       []database.ApplicationLogRow
	AppPage       int
	AppTotalPages int
}

func (h *Handlers) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.Users.PlatformStats(r.Context())
	if err != nil {
		http.Error(w, "could not load platform stats", http.StatusInternalServerError)
		return
	}

	tab := r.URL.Query().Get("tab")
	data := adminDashboardData{
		BasePageData:        newBasePageData(r, "dashboard"),
		TotalUsers:          stats.TotalUsers,
		TotalCandidates:     stats.TotalCandidates,
		TotalRecruiters:     stats.TotalRecruiters,
		TotalCompanies:      stats.TotalCompanies,
		TotalJobs:           stats.TotalJobs,
		TotalApplications:   stats.TotalApplications,
		PendingCompanyCount: stats.PendingCompanyCount,
		ActiveTab:           tab,
	}

	switch tab {
	case "users":
		pg := parsePage(r.URL.Query().Get("up"))
		rows, total, _ := h.Users.ListAllUsers(r.Context(), logPageSize, (pg-1)*logPageSize)
		data.UserLogs = rows
		data.UserPage = pg
		data.UserTotalPages = totalPages(total, logPageSize)
	case "candidates":
		pg := parsePage(r.URL.Query().Get("kp"))
		rows, total, _ := h.Users.ListCandidateRegistrations(r.Context(), logPageSize, (pg-1)*logPageSize)
		data.CandidateLogs = rows
		data.CandidatePage = pg
		data.CandidateTotalPages = totalPages(total, logPageSize)
	case "recruiters":
		pg := parsePage(r.URL.Query().Get("rp"))
		rows, total, _ := h.Users.ListRecruiterRegistrations(r.Context(), logPageSize, (pg-1)*logPageSize)
		data.RecruiterLogs = rows
		data.RecruiterPage = pg
		data.RecruiterTotalPages = totalPages(total, logPageSize)
	case "companies":
		pg := parsePage(r.URL.Query().Get("cp"))
		rows, total, _ := h.Companies.ListAllRegistrations(r.Context(), logPageSize, (pg-1)*logPageSize)
		data.CompanyLogs = rows
		data.CompanyPage = pg
		data.CompanyTotalPages = totalPages(total, logPageSize)
	case "jobs":
		pg := parsePage(r.URL.Query().Get("jp"))
		rows, total, _ := h.Jobs.ListAllForAdmin(r.Context(), logPageSize, (pg-1)*logPageSize)
		data.JobLogs = rows
		data.JobPage = pg
		data.JobTotalPages = totalPages(total, logPageSize)
	case "applications":
		pg := parsePage(r.URL.Query().Get("ap"))
		rows, total, _ := h.Applications.ListAllForAdmin(r.Context(), logPageSize, (pg-1)*logPageSize)
		data.AppLogs = rows
		data.AppPage = pg
		data.AppTotalPages = totalPages(total, logPageSize)
	}

	h.Render.Render(w, http.StatusOK, "admin-dashboard.html", data)
}
