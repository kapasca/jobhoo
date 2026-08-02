package handlers

import (
	"bytes"
	"context"
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
	"github.com/ledongthuc/pdf"
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
	ResumeText    string // legacy plain text only, for candidates who haven't re-uploaded since this feature shipped
	ResumeFileURL string
	Location      string
	Skills        string // comma-joined for the textarea
	Saved         bool
	Error         string
	Warning       string
}

// handleResumeUpload saves the uploaded "resume_file" field, validating type
// (PDF, DOCX, TXT) and returning the public URL, the raw file bytes (for AI
// extraction), and the detected extension. Returns ("", nil, "", nil) if no
// file was uploaded.
func handleResumeUpload(r *http.Request) (url string, data []byte, ext string, err error) {
	file, header, ferr := r.FormFile("resume_file")
	if ferr == http.ErrMissingFile {
		return "", nil, "", nil
	}
	if ferr != nil {
		return "", nil, "", fmt.Errorf("reading upload: %w", ferr)
	}
	defer file.Close()

	data, err = io.ReadAll(file)
	if err != nil {
		return "", nil, "", fmt.Errorf("reading file: %w", err)
	}

	mimeType := http.DetectContentType(data)
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
		return "", nil, "", fmt.Errorf("unsupported file type — please upload a PDF, DOCX, or TXT file")
	}

	var rnd [16]byte
	if _, err := rand.Read(rnd[:]); err != nil {
		return "", nil, "", fmt.Errorf("generating filename: %w", err)
	}
	filename := hex.EncodeToString(rnd[:]) + outExt

	if err := os.MkdirAll(resumeUploadsDir, 0o755); err != nil {
		return "", nil, "", fmt.Errorf("creating upload directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(resumeUploadsDir, filename), data, 0o644); err != nil {
		return "", nil, "", fmt.Errorf("saving file: %w", err)
	}
	return resumeURLPrefix + filename, data, outExt, nil
}

// extractResumeFile routes an uploaded resume to the right AI extraction
// method based on its file extension: PDFs are converted to text first (since
// gpt-5-nano vision only supports images), TXT is parsed as raw text, and DOCX
// is unzipped/stripped to plain text first.
func (h *Handlers) extractResumeFile(ctx context.Context, data []byte, ext string) (ai.ResumeExtraction, error) {
	switch ext {
	case ".pdf":
		// Extract text from PDF, then use text-based extraction
		// (gpt-5-nano vision doesn't support PDF MIME types)
		text, err := extractPDFText(data)
		if err != nil {
			return ai.ResumeExtraction{}, fmt.Errorf("could not extract text from PDF: %w", err)
		}
		if text == "" {
			return ai.ResumeExtraction{}, fmt.Errorf("PDF appears to be empty or unreadable")
		}
		return h.AI.ExtractResumeText(ctx, text)
	case ".txt":
		return h.AI.ExtractResumeText(ctx, string(data))
	case ".docx":
		text, err := ai.ExtractDocxText(data)
		if err != nil {
			return ai.ResumeExtraction{}, err
		}
		return h.AI.ExtractResumeText(ctx, text)
	default:
		return ai.ResumeExtraction{}, fmt.Errorf("unsupported resume file type %q", ext)
	}
}

// extractPDFText extracts text content from a PDF file using the ledongthuc/pdf library.
// It returns the concatenated text from all pages.
func extractPDFText(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("could not read PDF: %w", err)
	}

	var text strings.Builder
	pageCount := reader.NumPage()
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		p := reader.Page(pageNum)
		content, err := p.GetPlainText(nil)
		if err != nil {
			// Log but continue — some pages might be unreadable
			continue
		}
		text.WriteString(content)
		text.WriteString("\n")
	}
	return text.String(), nil
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
		if extraction, ok := ai.ParseStoredResumeText(profile.ResumeText); ok {
			data.ResumeText = ai.FormatResumeExtractionText(extraction)
		} else {
			data.ResumeText = profile.ResumeText
		}
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
		Location:     strings.TrimSpace(r.FormValue("location")),
		Skills:       r.FormValue("skills"),
	}

	// Preserve existing resume data — only replaced when a new file is uploaded.
	existing, existErr := h.Profiles.GetByUserID(r.Context(), user.ID)
	storedResumeText := ""
	if existErr == nil {
		data.ResumeFileURL = existing.ResumeFileURL
		storedResumeText = existing.ResumeText
		if extraction, ok := ai.ParseStoredResumeText(existing.ResumeText); ok {
			data.ResumeText = ai.FormatResumeExtractionText(extraction)
		} else {
			data.ResumeText = existing.ResumeText
		}
	}

	uploadedURL, fileData, ext, err := handleResumeUpload(r)
	if err != nil {
		data.Error = err.Error()
		h.Render.Render(w, http.StatusBadRequest, "profile.html", data)
		return
	}
	if uploadedURL != "" {
		data.ResumeFileURL = uploadedURL

		extraction, extractErr := h.extractResumeFile(withAIUser(r.Context(), user), fileData, ext)
		if extractErr != nil {
			// Keep the previously stored resume text so an AI hiccup doesn't
			// wipe out prior extraction results; just warn the candidate.
			data.Warning = "Your resume file was uploaded, but we couldn't read its content automatically: " + extractErr.Error()
			data.ResumeText = ""
			if legacyExtraction, ok := ai.ParseStoredResumeText(storedResumeText); ok {
				data.ResumeText = ai.FormatResumeExtractionText(legacyExtraction)
			}
		} else {
			storedResumeText = ai.MarshalResumeExtraction(extraction)
			data.ResumeText = ai.FormatResumeExtractionText(extraction)
		}
	}

	skills := splitSkills(data.Skills)
	if err := h.Profiles.Upsert(r.Context(), user.ID, data.Headline, storedResumeText, data.ResumeFileURL, data.Location, skills); err != nil {
		http.Error(w, "could not save profile", http.StatusInternalServerError)
		return
	}

	data.Saved = true
	h.Render.Render(w, http.StatusOK, "profile.html", data)
}

// resumeAnalysisData feeds the "How AI reads your resume" modal, showing the
// candidate exactly what the AI extracted from their uploaded file instead
// of a risky "suggest improvements" feature that could encourage fake data.
type resumeAnalysisData struct {
	Error      string
	Extraction ai.ResumeExtraction
	HasData    bool
}

// ResumeAnalysis is an HTMX endpoint: it shows the candidate how the AI
// parsed their uploaded resume (experience, education, skills, achievements,
// candidate story), using the cached extraction stored at upload time —
// falling back to a live extraction if the file predates this feature.
func (h *Handlers) ResumeAnalysis(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	profile, err := h.Profiles.GetByUserID(r.Context(), user.ID)
	if err != nil {
		h.Render.RenderBlock(w, "profile.html", "resume-analysis-modal", resumeAnalysisData{Error: "No resume found. Please upload one first."})
		return
	}

	if extraction, ok := ai.ParseStoredResumeText(profile.ResumeText); ok {
		h.Render.RenderBlock(w, "profile.html", "resume-analysis-modal", resumeAnalysisData{Extraction: extraction, HasData: true})
		return
	}

	if profile.ResumeFileURL == "" {
		h.Render.RenderBlock(w, "profile.html", "resume-analysis-modal", resumeAnalysisData{Error: "No resume file found. Please upload one first."})
		return
	}

	// Legacy profile (resume predates AI extraction, or a prior extraction
	// failed) — re-read the stored file from disk and extract live.
	ext := strings.ToLower(filepath.Ext(profile.ResumeFileURL))
	filePath := filepath.Join(resumeUploadsDir, filepath.Base(profile.ResumeFileURL))
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		h.Render.RenderBlock(w, "profile.html", "resume-analysis-modal", resumeAnalysisData{Error: "Couldn't read your stored resume file. Please re-upload it."})
		return
	}

	extraction, err := h.extractResumeFile(withAIUser(r.Context(), user), fileData, ext)
	if err != nil {
		h.Render.RenderBlock(w, "profile.html", "resume-analysis-modal", resumeAnalysisData{Error: "AI couldn't analyze your resume: " + err.Error()})
		return
	}

	// Cache the result so future views don't re-trigger an AI call.
	_ = h.Profiles.Upsert(r.Context(), user.ID, profile.Headline, ai.MarshalResumeExtraction(extraction), profile.ResumeFileURL, profile.Location, profile.Skills)

	h.Render.RenderBlock(w, "profile.html", "resume-analysis-modal", resumeAnalysisData{Extraction: extraction, HasData: true})
}

// jobRecommendationRow pairs a job with the AI's reasoning for candidate
// dashboard display.
type jobRecommendationRow struct {
	models.Job
	Reason string
	Score  float64
}

// RecommendedJobs is an HTMX endpoint on the candidate dashboard that asks
// the AI provider to suggest jobs, primarily based on the candidate's parsed
// resume content (with their profile skills list as a secondary signal).
func (h *Handlers) RecommendedJobs(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	profile, err := h.Profiles.GetByUserID(r.Context(), user.ID)
	if err != nil {
		h.Render.RenderBlock(w, "candidate-dashboard.html", "recommendations-empty", nil)
		return
	}

	result, err := h.Jobs.ListPublished(r.Context(), database.JobListFilter{Limit: 50})
	if err != nil {
		h.Render.RenderBlock(w, "candidate-dashboard.html", "recommendations-error", map[string]string{"error": "Could not load open jobs. Please try again."})
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

	resumeText := ""
	if extraction, ok := ai.ParseStoredResumeText(profile.ResumeText); ok {
		resumeText = ai.FormatResumeExtractionText(extraction)
	} else {
		resumeText = profile.ResumeText
	}

	recs, err := h.AI.RecommendJobs(withAIUser(r.Context(), user), ai.CandidateContext{ID: user.ID, ResumeText: resumeText, Skills: profile.Skills}, jobCtxs)
	if err != nil {
		h.Render.RenderBlock(w, "candidate-dashboard.html", "recommendations-error", map[string]string{"error": "AI recommendation failed: " + err.Error()})
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
