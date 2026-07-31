package handlers

import (
	"net/http"
	"strconv"

	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/middleware"
	"github.com/jobhoo/jobhoo/internal/models"
)

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
	CompanyLogs         []database.CompanyWithOwner
	CompanyPage         int
	CompanyTotalPages   int
	CandidateLogs       []database.CandidateRegistration
	CandidatePage       int
	CandidateTotalPages int
}

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

func (h *Handlers) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.Users.PlatformStats(r.Context())
	if err != nil {
		http.Error(w, "could not load platform stats", http.StatusInternalServerError)
		return
	}

	cp := parsePage(r.URL.Query().Get("cp"))
	kp := parsePage(r.URL.Query().Get("kp"))

	companyLogs, companyTotal, _ := h.Companies.ListAllRegistrations(r.Context(), logPageSize, (cp-1)*logPageSize)
	candidateLogs, candidateTotal, _ := h.Users.ListCandidateRegistrations(r.Context(), logPageSize, (kp-1)*logPageSize)

	h.Render.Render(w, http.StatusOK, "admin-dashboard.html", adminDashboardData{
		BasePageData:        newBasePageData(r, "dashboard"),
		TotalUsers:          stats.TotalUsers,
		TotalCandidates:     stats.TotalCandidates,
		TotalRecruiters:     stats.TotalRecruiters,
		TotalCompanies:      stats.TotalCompanies,
		TotalJobs:           stats.TotalJobs,
		TotalApplications:   stats.TotalApplications,
		PendingCompanyCount: stats.PendingCompanyCount,
		CompanyLogs:         companyLogs,
		CompanyPage:         cp,
		CompanyTotalPages:   totalPages(companyTotal, logPageSize),
		CandidateLogs:       candidateLogs,
		CandidatePage:       kp,
		CandidateTotalPages: totalPages(candidateTotal, logPageSize),
	})
}
