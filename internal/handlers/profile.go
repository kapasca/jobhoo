package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jobhoo/jobhoo/internal/ai"
	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/middleware"
	"github.com/jobhoo/jobhoo/internal/models"
)

const (
	resumeUploadsDir = "web/static/uploads/resumes"
	resumeURLPrefix  = "/static/uploads/resumes/"
	resumeMaxBytes   = 5 << 20 // 5 MB
)

type profileData struct {
	BasePageData
	Headline      string
	ResumeText    string
	ResumeFileURL string
	Location      string
	Skills        string // comma-joined for the textarea
	Saved         bool
	Error         string
	Suggestions   []string
}

// handleResumeUpload saves the uploaded "resume_file" field, validating type
// (PDF, DOCX, TXT) and returning the public URL path. Returns "" if no file.
func handleResumeUpload(r *http.Request) (string, error) {
	file, header, err := r.FormFile("resume_file")
	if err == http.ErrMissingFile {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading upload: %w", err)
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("reading file header: %w", err)
	}
	mimeType := http.DetectContentType(buf[:n])
	origExt := strings.ToLower(filepath.Ext(header.Filename))

	var outExt string
	switch {
	case mimeType == "application/pdf":
		outExt = ".pdf"
	case mimeType == "text/plain":
		outExt = ".txt"
	case origExt == ".docx" && (mimeType == "application/zip" || mimeType == "application/octet-stream"):
		outExt = ".docx"
	default:
		return "", fmt.Errorf("unsupported file type — please upload a PDF, DOCX, or TXT file")
	}

	var rnd [16]byte
	if _, err := rand.Read(rnd[:]); err != nil {
		return "", fmt.Errorf("generating filename: %w", err)
	}
	filename := hex.EncodeToString(rnd[:]) + outExt

	if err := os.MkdirAll(resumeUploadsDir, 0o755); err != nil {
		return "", fmt.Errorf("creating upload directory: %w", err)
	}
	dst, err := os.Create(filepath.Join(resumeUploadsDir, filename))
	if err != nil {
		return "", fmt.Errorf("saving file: %w", err)
	}
	defer dst.Close()

	if _, err := dst.Write(buf[:n]); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}
	return resumeURLPrefix + filename, nil
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
		data.ResumeFileURL = profile.ResumeFileURL
		data.Location = profile.Location
		data.Skills = strings.Join(profile.Skills, ", ")
	}

	h.Render.Render(w, http.StatusOK, "profile.html", data)
}

func (h *Handlers) ProfileUpdate(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	r.Body = http.MaxBytesReader(w, r.Body, resumeMaxBytes+4096)
	if err := r.ParseMultipartForm(resumeMaxBytes); err != nil {
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}

	data := profileData{
		BasePageData: newBasePageData(r, "profile"),
		Headline:     strings.TrimSpace(r.FormValue("headline")),
		ResumeText:   strings.TrimSpace(r.FormValue("resume_text")),
		Location:     strings.TrimSpace(r.FormValue("location")),
		Skills:       r.FormValue("skills"),
	}

	// Preserve existing file URL — only replaced when a new file is uploaded.
	if existing, err := h.Profiles.GetByUserID(r.Context(), user.ID); err == nil {
		data.ResumeFileURL = existing.ResumeFileURL
	}

	uploadedURL, err := handleResumeUpload(r)
	if err != nil {
		data.Error = err.Error()
		h.Render.Render(w, http.StatusBadRequest, "profile.html", data)
		return
	}
	if uploadedURL != "" {
		data.ResumeFileURL = uploadedURL
	}

	skills := splitSkills(data.Skills)
	if err := h.Profiles.Upsert(r.Context(), user.ID, data.Headline, data.ResumeText, data.ResumeFileURL, data.Location, skills); err != nil {
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
