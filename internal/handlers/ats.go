package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jobhoo/jobhoo/internal/ai"
	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/middleware"
	"github.com/jobhoo/jobhoo/internal/models"
)

type atsColumn struct {
	Stage        models.ApplicationStage
	Label        string
	Applications []atsApplicationRow
}

type atsApplicationRow struct {
	models.Application
	AIScore   float64
	AISummary string
	HasAIRank bool
	CSRFToken string
}

type atsBoardData struct {
	BasePageData
	Job      models.Job
	Columns  []atsColumn
	Rejected []atsApplicationRow
}

// jobOwnedByCurrentRecruiter loads a job and verifies the current recruiter's
// company owns it, so one recruiter can never view or edit another
// company's pipeline just by guessing a job ID.
func (h *Handlers) jobOwnedByCurrentRecruiter(w http.ResponseWriter, r *http.Request) (*models.Job, bool) {
	id := chi.URLParam(r, "id")
	job, err := h.Jobs.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			h.NotFoundPage(w, r)
			return nil, false
		}
		http.Error(w, "could not load job", http.StatusInternalServerError)
		return nil, false
	}

	company, ok := h.requireCompany(w, r)
	if !ok {
		return nil, false
	}
	if job.CompanyID != company.ID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return nil, false
	}
	return job, true
}

func (h *Handlers) ATSBoard(w http.ResponseWriter, r *http.Request) {
	job, ok := h.jobOwnedByCurrentRecruiter(w, r)
	if !ok {
		return
	}

	apps, err := h.Applications.ListByJob(r.Context(), job.ID)
	if err != nil {
		http.Error(w, "could not load applicants", http.StatusInternalServerError)
		return
	}

	csrfToken := middleware.CSRFToken(r)
	data := atsBoardData{BasePageData: newBasePageData(r, "dashboard"), Job: *job}
	for _, stage := range models.OrderedStages {
		col := atsColumn{Stage: stage, Label: stage.Label()}
		for _, a := range apps {
			if a.Stage == stage {
				col.Applications = append(col.Applications, atsApplicationRow{Application: a, CSRFToken: csrfToken})
			}
		}
		data.Columns = append(data.Columns, col)
	}
	for _, a := range apps {
		if a.Stage == models.StageRejected {
			data.Rejected = append(data.Rejected, atsApplicationRow{Application: a, CSRFToken: csrfToken})
		}
	}

	h.Render.Render(w, http.StatusOK, "ats-board.html", data)
}

// UpdateApplicationStage moves a single application to a new stage. It's an
// HTMX endpoint used by the stage-change buttons on the ATS board; it
// re-renders the whole board so counts and columns stay consistent.
func (h *Handlers) UpdateApplicationStage(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	newStage := models.ApplicationStage(r.FormValue("stage"))

	app, err := h.Applications.GetByID(r.Context(), appID)
	if err != nil {
		h.NotFoundPage(w, r)
		return
	}

	job, err := h.Jobs.GetByID(r.Context(), app.JobID)
	if err != nil {
		h.NotFoundPage(w, r)
		return
	}
	company, ok := h.requireCompany(w, r)
	if !ok {
		return
	}
	if job.CompanyID != company.ID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if err := h.Applications.UpdateStage(r.Context(), appID, newStage); err != nil {
		http.Error(w, "could not update stage", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/recruiter/jobs/"+job.ID+"/applicants", http.StatusSeeOther)
}

// RankCandidates asks the configured AI provider to score every non-rejected
// applicant against the job's requirements, then re-renders the board with
// scores attached. Ranking is advisory only: it never changes an
// application's stage or removes anyone from view.
func (h *Handlers) RankCandidates(w http.ResponseWriter, r *http.Request) {
	job, ok := h.jobOwnedByCurrentRecruiter(w, r)
	if !ok {
		return
	}

	apps, err := h.Applications.ListByJob(r.Context(), job.ID)
	if err != nil {
		http.Error(w, "could not load applicants", http.StatusInternalServerError)
		return
	}

	jobCtx := ai.JobContext{
		ID: job.ID, Title: job.Title, Description: job.Description,
		MustHave: job.MustHaveSkills, NiceToHave: job.NiceToHaveSkills, Seniority: job.Seniority,
	}
	candidateCtxs := make([]ai.CandidateContext, 0, len(apps))
	for _, a := range apps {
		candidateCtxs = append(candidateCtxs, ai.CandidateContext{
			ID: a.CandidateID, Name: a.CandidateName, Skills: a.CandidateSkills,
		})
	}

	rankings, err := h.AI.RankCandidates(r.Context(), jobCtx, candidateCtxs)
	scoreByCandidate := map[string]ai.CandidateRanking{}
	if err == nil {
		for _, rk := range rankings {
			scoreByCandidate[rk.CandidateID] = rk
		}
	}

	csrfToken := middleware.CSRFToken(r)
	data := atsBoardData{BasePageData: newBasePageData(r, "dashboard"), Job: *job}
	for _, stage := range models.OrderedStages {
		col := atsColumn{Stage: stage, Label: stage.Label()}
		for _, a := range apps {
			if a.Stage == stage {
				row := atsApplicationRow{Application: a, CSRFToken: csrfToken}
				if rk, found := scoreByCandidate[a.CandidateID]; found {
					row.AIScore, row.AISummary, row.HasAIRank = rk.Score, rk.Summary, true
				}
				col.Applications = append(col.Applications, row)
			}
		}
		data.Columns = append(data.Columns, col)
	}

	h.Render.RenderBlock(w, "ats-board.html", "ats-columns", data)
}
