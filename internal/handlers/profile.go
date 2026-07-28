package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/jobhoo/jobhoo/internal/ai"
	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/middleware"
	"github.com/jobhoo/jobhoo/internal/models"
)

type profileData struct {
	BasePageData
	Headline    string
	ResumeText  string
	Location    string
	Skills      string // comma-joined for the textarea
	Saved       bool
	Suggestions []string
}

func (h *Handlers) ProfilePage(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	data := profileData{BasePageData: newBasePageData(r, "profile")}

	profile, err := h.Profiles.GetByUserID(r.Context(), user.ID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		http.Error(w, "could not load profile", http.StatusInternalServerError)
		return
	}
	if profile != nil {
		data.Headline = profile.Headline
		data.ResumeText = profile.ResumeText
		data.Location = profile.Location
		data.Skills = strings.Join(profile.Skills, ", ")
	}

	h.Render.Render(w, http.StatusOK, "profile.html", data)
}

func (h *Handlers) ProfileUpdate(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	data := profileData{
		BasePageData: newBasePageData(r, "profile"),
		Headline:     strings.TrimSpace(r.FormValue("headline")),
		ResumeText:   strings.TrimSpace(r.FormValue("resume_text")),
		Location:     strings.TrimSpace(r.FormValue("location")),
		Skills:       r.FormValue("skills"),
	}

	skills := splitSkills(data.Skills)
	if err := h.Profiles.Upsert(r.Context(), user.ID, data.Headline, data.ResumeText, data.Location, skills); err != nil {
		http.Error(w, "could not save profile", http.StatusInternalServerError)
		return
	}

	data.Saved = true
	h.Render.Render(w, http.StatusOK, "profile.html", data)
}

// ResumeSuggestions is an HTMX endpoint: it asks the AI provider for resume
// improvement suggestions based on the candidate's saved resume text and
// returns just the suggestions list fragment.
func (h *Handlers) ResumeSuggestions(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	profile, err := h.Profiles.GetByUserID(r.Context(), user.ID)
	if err != nil || profile.ResumeText == "" {
		h.Render.RenderBlock(w, "profile.html", "suggestions-empty", nil)
		return
	}

	suggestions, err := h.AI.SuggestResumeImprovements(r.Context(), profile.ResumeText, nil)
	if err != nil {
		http.Error(w, "could not generate suggestions", http.StatusInternalServerError)
		return
	}

	h.Render.RenderBlock(w, "profile.html", "suggestions-list", suggestions)
}

// jobRecommendationRow pairs a job with the AI's reasoning for candidate
// dashboard display.
type jobRecommendationRow struct {
	models.Job
	Reason string
	Score  float64
}

// RecommendedJobs is an HTMX endpoint on the candidate dashboard that asks
// the AI provider to suggest jobs based on the candidate's profile skills.
func (h *Handlers) RecommendedJobs(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	profile, err := h.Profiles.GetByUserID(r.Context(), user.ID)
	if err != nil {
		h.Render.RenderBlock(w, "candidate-dashboard.html", "recommendations-empty", nil)
		return
	}

	result, err := h.Jobs.ListPublished(r.Context(), database.JobListFilter{Limit: 50})
	if err != nil {
		http.Error(w, "could not load jobs", http.StatusInternalServerError)
		return
	}

	jobCtxs := make([]ai.JobContext, 0, len(result.Jobs))
	jobByID := map[string]models.Job{}
	for _, j := range result.Jobs {
		jobCtxs = append(jobCtxs, ai.JobContext{
			ID: j.ID, Title: j.Title, MustHave: j.MustHaveSkills, NiceToHave: j.NiceToHaveSkills,
		})
		jobByID[j.ID] = j
	}

	recs, err := h.AI.RecommendJobs(r.Context(), ai.CandidateContext{ID: user.ID, Skills: profile.Skills}, jobCtxs)
	if err != nil {
		http.Error(w, "could not generate recommendations", http.StatusInternalServerError)
		return
	}

	rows := make([]jobRecommendationRow, 0, 5)
	for i, rec := range recs {
		if i >= 5 {
			break
		}
		if job, ok := jobByID[rec.JobID]; ok {
			rows = append(rows, jobRecommendationRow{Job: job, Reason: rec.Reason, Score: rec.Score})
		}
	}

	h.Render.RenderBlock(w, "candidate-dashboard.html", "recommendations-list", rows)
}
