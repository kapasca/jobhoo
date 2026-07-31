package handlers

import (
	"net/http"

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
}

func (h *Handlers) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.Users.PlatformStats(r.Context())
	if err != nil {
		http.Error(w, "could not load platform stats", http.StatusInternalServerError)
		return
	}
	h.Render.Render(w, http.StatusOK, "admin-dashboard.html", adminDashboardData{
		BasePageData:        newBasePageData(r, "dashboard"),
		TotalUsers:          stats.TotalUsers,
		TotalCandidates:     stats.TotalCandidates,
		TotalRecruiters:     stats.TotalRecruiters,
		TotalCompanies:      stats.TotalCompanies,
		TotalJobs:           stats.TotalJobs,
		TotalApplications:   stats.TotalApplications,
		PendingCompanyCount: stats.PendingCompanyCount,
	})
}
