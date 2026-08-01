package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/middleware"
)

// CompanyApprovalQueue lists every company awaiting admin review.
func (h *Handlers) CompanyApprovalQueue(w http.ResponseWriter, r *http.Request) {
	companies, err := h.Companies.ListPendingWithOwner(r.Context())
	if err != nil {
		http.Error(w, "could not load pending companies", http.StatusInternalServerError)
		return
	}
	h.Render.Render(w, http.StatusOK, "admin-approvals.html", struct {
		BasePageData
		Companies []database.CompanyWithOwner
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
	// Notify company owner that their company was approved (best-effort)
	go func() {
		if c, err := h.Companies.GetByID(r.Context(), id); err == nil {
			if owner, err2 := h.Users.GetByID(r.Context(), c.OwnerID); err2 == nil {
				subj := "Your company is approved on JOBHOO"
				link := "/company/profile"
				text := "Your company has been approved and you can now post jobs: " + link
				html := "<p>Your company has been <strong>approved</strong>. <a href=\"" + link + "\">Go to profile</a></p>"
				_ = h.Email.Send(owner.Email, subj, html, text)
			}
		}
	}()
	http.Redirect(w, r, "/admin/approvals", http.StatusSeeOther)
}

// BlacklistCompany permanently blocks a company; the recruiter cannot re-apply.
func (h *Handlers) BlacklistCompany(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	admin := middleware.CurrentUser(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	reason := strings.TrimSpace(r.FormValue("reason"))
	if reason == "" {
		reason = "Your company has been permanently blocked from JOBHOO."
	}
	if err := h.Companies.Blacklist(r.Context(), id, admin.ID, reason); err != nil {
		http.Error(w, "could not blacklist company", http.StatusInternalServerError)
		return
	}
	go func() {
		if c, err := h.Companies.GetByID(r.Context(), id); err == nil {
			if owner, err2 := h.Users.GetByID(r.Context(), c.OwnerID); err2 == nil {
				subj := "Your company has been blocked on JOBHOO"
				text := reason
				html := "<p>Your company has been blocked: " + reason + "</p>"
				_ = h.Email.Send(owner.Email, subj, html, text)
			}
		}
	}()
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
	go func() {
		if c, err := h.Companies.GetByID(r.Context(), id); err == nil {
			if owner, err2 := h.Users.GetByID(r.Context(), c.OwnerID); err2 == nil {
				subj := "Your company application was rejected on JOBHOO"
				text := reason
				html := "<p>Your company application was <strong>rejected</strong>: " + reason + "</p>"
				_ = h.Email.Send(owner.Email, subj, html, text)
			}
		}
	}()
	http.Redirect(w, r, "/admin/approvals", http.StatusSeeOther)
}
