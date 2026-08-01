package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/models"
)

type adminUserDetailData struct {
	BasePageData
	User models.User
}

func (h *Handlers) AdminUserDetail(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	user, err := h.Users.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "could not load user", http.StatusInternalServerError)
		return
	}

	data := adminUserDetailData{
		BasePageData: newBasePageData(r, ""),
		User:         *user,
	}
	h.Render.RenderBlock(w, "admin-user-modal.html", "admin-user-modal", data)
}

type adminCandidateDetailData struct {
	BasePageData
	User    models.User
	Profile *models.CandidateProfile
}

func (h *Handlers) AdminCandidateDetail(w http.ResponseWriter, r *http.Request) {
	candidateID := chi.URLParam(r, "candidateID")
	user, err := h.Users.GetByID(r.Context(), candidateID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "could not load candidate", http.StatusInternalServerError)
		return
	}

	profile, _ := h.Profiles.GetByUserID(r.Context(), candidateID) // OK if profile doesn't exist

	data := adminCandidateDetailData{
		BasePageData: newBasePageData(r, ""),
		User:         *user,
		Profile:      profile,
	}
	h.Render.RenderBlock(w, "admin-candidate-modal.html", "admin-candidate-modal", data)
}

type adminRecruiterDetailData struct {
	BasePageData
	User    models.User
	Company *models.Company
}

func (h *Handlers) AdminRecruiterDetail(w http.ResponseWriter, r *http.Request) {
	recruiterID := chi.URLParam(r, "recruiterID")
	user, err := h.Users.GetByID(r.Context(), recruiterID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "could not load recruiter", http.StatusInternalServerError)
		return
	}

	company, _ := h.Companies.GetByOwnerID(r.Context(), recruiterID) // OK if no company

	data := adminRecruiterDetailData{
		BasePageData: newBasePageData(r, ""),
		User:         *user,
		Company:      company,
	}
	h.Render.RenderBlock(w, "admin-recruiter-modal.html", "admin-recruiter-modal", data)
}

type adminCompanyDetailData struct {
	BasePageData
	Company models.Company
	Owner   models.User
}

func (h *Handlers) AdminCompanyDetail(w http.ResponseWriter, r *http.Request) {
	companyID := chi.URLParam(r, "companyID")
	company, err := h.Companies.GetByID(r.Context(), companyID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "could not load company", http.StatusInternalServerError)
		return
	}

	owner, _ := h.Users.GetByID(r.Context(), company.OwnerID)

	data := adminCompanyDetailData{
		BasePageData: newBasePageData(r, ""),
		Company:      *company,
		Owner:        *owner,
	}
	h.Render.RenderBlock(w, "admin-company-modal.html", "admin-company-modal", data)
}

type adminJobDetailData struct {
	BasePageData
	Job     models.Job
	Company models.Company
	Owner   models.User
}

func (h *Handlers) AdminJobDetail(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")
	job, err := h.Jobs.GetByID(r.Context(), jobID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "could not load job", http.StatusInternalServerError)
		return
	}

	company, _ := h.Companies.GetByID(r.Context(), job.CompanyID)
	owner, _ := h.Users.GetByID(r.Context(), company.OwnerID)

	data := adminJobDetailData{
		BasePageData: newBasePageData(r, ""),
		Job:          *job,
		Company:      *company,
		Owner:        *owner,
	}
	h.Render.RenderBlock(w, "admin-job-modal.html", "admin-job-modal", data)
}
