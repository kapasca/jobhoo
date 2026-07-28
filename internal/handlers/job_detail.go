package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/middleware"
	"github.com/jobhoo/jobhoo/internal/models"
)

type jobDetailData struct {
	BasePageData
	Job        models.Job
	IsSaved    bool
	HasApplied bool
	CanApply   bool
	ApplyError string
	ApplySent  bool
}

func (h *Handlers) JobDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, err := h.Jobs.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			h.NotFoundPage(w, r)
			return
		}
		http.Error(w, "could not load job", http.StatusInternalServerError)
		return
	}

	data := jobDetailData{BasePageData: newBasePageData(r, "jobs"), Job: *job}

	if user := middleware.CurrentUser(r); user != nil && user.Role == models.RoleCandidate {
		data.CanApply = true
		if saved, err := h.SavedJobs.IsSaved(r.Context(), user.ID, job.ID); err == nil {
			data.IsSaved = saved
		}
		if applied, err := h.Applications.HasApplied(r.Context(), job.ID, user.ID); err == nil {
			data.HasApplied = applied
		}
	}

	h.Render.Render(w, http.StatusOK, "job-detail.html", data)
}

func (h *Handlers) ApplyToJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user := middleware.CurrentUser(r) // guaranteed non-nil: route requires auth+candidate role

	job, err := h.Jobs.GetByID(r.Context(), id)
	if err != nil {
		h.NotFoundPage(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	coverNote := r.FormValue("cover_note")

	data := jobDetailData{BasePageData: newBasePageData(r, "jobs"), Job: *job, CanApply: true}

	_, err = h.Applications.Create(r.Context(), job.ID, user.ID, coverNote)
	if err != nil {
		if errors.Is(err, database.ErrAlreadyApplied) {
			data.HasApplied = true
		} else {
			data.ApplyError = "Something went wrong submitting your application. Please try again."
		}
		h.Render.Render(w, http.StatusOK, "job-detail.html", data)
		return
	}

	data.HasApplied = true
	data.ApplySent = true
	h.Render.Render(w, http.StatusOK, "job-detail.html", data)
}

// SaveJob is an HTMX endpoint: it toggles the bookmark and returns just the
// updated bookmark button markup so it can swap in place.
func (h *Handlers) SaveJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user := middleware.CurrentUser(r)

	saved, err := h.SavedJobs.Toggle(r.Context(), user.ID, id)
	if err != nil {
		http.Error(w, "could not update saved jobs", http.StatusInternalServerError)
		return
	}

	label := "Save job"
	activeClass := ""
	fill := "none"
	if saved {
		label = "Saved"
		activeClass = "is-saved"
		fill = "currentColor"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<button class="job-card__bookmark ` + activeClass + `" title="` + label + `" aria-label="` + label + `"
		hx-post="/jobs/` + id + `/save" hx-swap="outerHTML">
		<svg width="20" height="20" viewBox="0 0 24 24" fill="` + fill + `" stroke="currentColor" stroke-width="2">
			<path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"/>
		</svg>
	</button>`))
}
