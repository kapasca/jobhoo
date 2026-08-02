package handlers

import (
	"encoding/json"
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
	AIScore            float64
	AISummary          string
	HasAIRank          bool
	CSRFToken          string
	ResumeExtractionJS string // JSON-encoded ai.ResumeExtraction for the candidate detail modal, empty if none parsed
}

type atsBoardData struct {
	BasePageData
	Job       models.Job
	Columns   []atsColumn
	Rejected  []atsApplicationRow
	RankError string
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

	// Load any previously computed AI scores so they persist across page
	// reloads instead of disappearing until "Rank candidates" is clicked again.
	scores := map[string]aiScoreEntry{}
	if insights, err := h.AIInsights.ListByJob(r.Context(), job.ID); err == nil {
		for candidateID, insight := range insights {
			if insight.Summary != "" || insight.Score != 0 {
				scores[candidateID] = aiScoreEntry{Score: insight.Score, Summary: insight.Summary}
			}
		}
	}

	csrfToken := middleware.CSRFToken(r)
	data := atsBoardData{BasePageData: newBasePageData(r, "dashboard"), Job: *job}
	data.Columns, data.Rejected = buildATSColumns(apps, scores, csrfToken)

	h.Render.Render(w, http.StatusOK, "ats-board.html", data)
}

// aiScoreEntry is a cached or freshly-computed AI ranking result for one
// candidate, used to populate atsApplicationRow.
type aiScoreEntry struct {
	Score   float64
	Summary string
}

// buildATSColumns groups applications into ATS board columns by stage, and a
// separate rejected list, attaching any available AI scores along the way.
func buildATSColumns(apps []models.Application, scores map[string]aiScoreEntry, csrfToken string) (columns []atsColumn, rejected []atsApplicationRow) {
	for _, stage := range models.OrderedStages {
		col := atsColumn{Stage: stage, Label: stage.Label()}
		for _, a := range apps {
			if a.Stage != stage {
				continue
			}
			row := newATSApplicationRow(a, scores, csrfToken)
			col.Applications = append(col.Applications, row)
		}
		columns = append(columns, col)
	}
	for _, a := range apps {
		if a.Stage == models.StageRejected {
			rejected = append(rejected, newATSApplicationRow(a, scores, csrfToken))
		}
	}
	return columns, rejected
}

// newATSApplicationRow builds a single card's data, attaching any cached AI
// score and the candidate's parsed resume extraction (if any) so the detail
// modal can show what the AI actually read from their resume.
func newATSApplicationRow(a models.Application, scores map[string]aiScoreEntry, csrfToken string) atsApplicationRow {
	row := atsApplicationRow{Application: a, CSRFToken: csrfToken}
	if entry, found := scores[a.CandidateID]; found {
		row.AIScore, row.AISummary, row.HasAIRank = entry.Score, entry.Summary, true
	}
	if extraction, ok := ai.ParseStoredResumeText(a.CandidateResumeText); ok {
		if data, err := json.Marshal(extraction); err == nil {
			row.ResumeExtractionJS = string(data)
		}
	}
	return row
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

// candidateResumeText picks the best available text representation of a
// candidate's resume for AI prompts: the structured extraction if present
// (the normal case), otherwise whatever raw text is stored (legacy data).
func candidateResumeText(raw string) string {
	if extraction, ok := ai.ParseStoredResumeText(raw); ok {
		return ai.FormatResumeExtractionText(extraction)
	}
	return raw
}

// RankCandidates asks the configured AI provider to score every non-rejected
// applicant against the job's requirements, persists each score so it
// survives page reloads, then re-renders the board. Ranking is advisory
// only: it never changes an application's stage or removes anyone from view.
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

	// Start from cached scores so a failed AI call still shows prior results
	// instead of blanking out the whole board.
	scores := map[string]aiScoreEntry{}
	if insights, err := h.AIInsights.ListByJob(r.Context(), job.ID); err == nil {
		for candidateID, insight := range insights {
			if insight.Summary != "" || insight.Score != 0 {
				scores[candidateID] = aiScoreEntry{Score: insight.Score, Summary: insight.Summary}
			}
		}
	}

	jobCtx := ai.JobContext{
		ID: job.ID, Title: job.Title, Description: job.Description,
		MustHave: job.MustHaveSkills, NiceToHave: job.NiceToHaveSkills, Seniority: job.Seniority,
	}
	candidateCtxs := make([]ai.CandidateContext, 0, len(apps))
	for _, a := range apps {
		candidateCtxs = append(candidateCtxs, ai.CandidateContext{
			ID: a.CandidateID, Name: a.CandidateName,
			ResumeText: candidateResumeText(a.CandidateResumeText), Skills: a.CandidateSkills,
		})
	}

	rankings, err := h.AI.RankCandidates(withAIUser(r.Context(), middleware.CurrentUser(r)), jobCtx, candidateCtxs)
	var rankErr string
	if err != nil {
		rankErr = "AI ranking failed: " + err.Error() + " — showing previously computed scores, if any."
	} else {
		for _, rk := range rankings {
			scores[rk.CandidateID] = aiScoreEntry{Score: rk.Score, Summary: rk.Summary}
			_ = h.AIInsights.UpsertRanking(r.Context(), job.ID, rk.CandidateID, h.AI.Name(), rk.Score, rk.Summary)
		}
	}

	csrfToken := middleware.CSRFToken(r)
	data := atsBoardData{BasePageData: newBasePageData(r, "dashboard"), Job: *job, RankError: rankErr}
	data.Columns, data.Rejected = buildATSColumns(apps, scores, csrfToken)

	h.Render.RenderBlock(w, "ats-board.html", "ats-columns", data)
}

// explainMatchData feeds the "Explain Match" result fragment shown in the
// candidate detail card on the ATS board.
type explainMatchData struct {
	Error       string
	OverallNote string
	Strengths   []string
	Gaps        []string
	Score       float64
	HasScore    bool
}

// ExplainMatch produces (and caches) an AI explanation of why a specific
// candidate does or doesn't fit a specific job, shown behind an "Explain
// Match" button on the candidate detail card.
func (h *Handlers) ExplainMatch(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")

	app, err := h.Applications.GetByID(r.Context(), appID)
	if err != nil {
		h.Render.RenderBlock(w, "ats-board.html", "explain-match-result", explainMatchData{Error: "Application not found."})
		return
	}
	job, err := h.Jobs.GetByID(r.Context(), app.JobID)
	if err != nil {
		h.Render.RenderBlock(w, "ats-board.html", "explain-match-result", explainMatchData{Error: "Job not found."})
		return
	}
	company, ok := h.requireCompany(w, r)
	if !ok {
		return
	}
	if job.CompanyID != company.ID {
		h.Render.RenderBlock(w, "ats-board.html", "explain-match-result", explainMatchData{Error: "You don't have access to this application."})
		return
	}

	// Serve from cache if we already have both a score and an explanation
	// for this pair — avoids re-billing the AI gateway on repeat views.
	if cached, err := h.AIInsights.Get(r.Context(), job.ID, app.CandidateID); err == nil && (len(cached.Strengths) > 0 || len(cached.Gaps) > 0 || cached.OverallNote != "") {
		h.Render.RenderBlock(w, "ats-board.html", "explain-match-result", explainMatchData{
			OverallNote: cached.OverallNote, Strengths: cached.Strengths, Gaps: cached.Gaps,
			Score: cached.Score, HasScore: cached.Score != 0,
		})
		return
	}

	candidate, err := h.Users.GetByID(r.Context(), app.CandidateID)
	if err != nil {
		h.Render.RenderBlock(w, "ats-board.html", "explain-match-result", explainMatchData{Error: "Candidate not found."})
		return
	}
	profile, err := h.Profiles.GetByUserID(r.Context(), app.CandidateID)
	resumeText, skills := "", []string(nil)
	if err == nil {
		resumeText = candidateResumeText(profile.ResumeText)
		skills = profile.Skills
	}

	jobCtx := ai.JobContext{
		ID: job.ID, Title: job.Title, Description: job.Description,
		MustHave: job.MustHaveSkills, NiceToHave: job.NiceToHaveSkills, Seniority: job.Seniority,
	}
	candidateCtx := ai.CandidateContext{ID: candidate.ID, Name: candidate.FullName, ResumeText: resumeText, Skills: skills}

	explanation, err := h.AI.ExplainMatch(withAIUser(r.Context(), middleware.CurrentUser(r)), jobCtx, candidateCtx)
	if err != nil {
		h.Render.RenderBlock(w, "ats-board.html", "explain-match-result", explainMatchData{Error: "AI explanation failed: " + err.Error()})
		return
	}

	_ = h.AIInsights.UpsertExplanation(r.Context(), job.ID, app.CandidateID, h.AI.Name(), explanation.Strengths, explanation.Gaps, explanation.OverallNote)

	var score float64
	var hasScore bool
	if insight, err := h.AIInsights.Get(r.Context(), job.ID, app.CandidateID); err == nil {
		score, hasScore = insight.Score, insight.Score != 0
	}

	h.Render.RenderBlock(w, "ats-board.html", "explain-match-result", explainMatchData{
		OverallNote: explanation.OverallNote, Strengths: explanation.Strengths, Gaps: explanation.Gaps,
		Score: score, HasScore: hasScore,
	})
}
