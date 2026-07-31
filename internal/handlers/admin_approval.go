package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jobhoo/jobhoo/internal/middleware"
	"github.com/jobhoo/jobhoo/internal/models"
)

// CompanyApprovalQueue lists every company awaiting admin review.
func (h *Handlers) CompanyApprovalQueue(w http.ResponseWriter, r *http.Request) {
	companies, err := h.Companies.ListPending(r.Context())
	if err != nil {
		http.Error(w, "could not load pending companies", http.StatusInternalServerError)
		return
	}
	h.Render.Render(w, http.StatusOK, "admin-approvals.html", struct {
		BasePageData
		Companies []models.Company
	}{
		BasePageData: newBasePageData(r, "dashboard"),
		Companies:    companies,
	})
}

// ApproveCompany approves a pending company, letting its recruiter start
// posting jobs immediately.
func (h *Handlers) ApproveCompany(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	admin := middleware.CurrentUser(r)
	if err := h.Companies.Approve(r.Context(), id, admin.ID); err != nil {
		http.Error(w, "could not approve company", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/approvals", http.StatusSeeOther)
}

// RejectCompany rejects a pending company with a reason the recruiter will
// see on their dashboard.
func (h *Handlers) RejectCompany(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	admin := middleware.CurrentUser(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	reason := strings.TrimSpace(r.FormValue("reason"))
	if reason == "" {
		reason = "Did not meet JOBHOO's listing criteria."
	}
	if err := h.Companies.Reject(r.Context(), id, admin.ID, reason); err != nil {
		http.Error(w, "could not reject company", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/approvals", http.StatusSeeOther)
}
