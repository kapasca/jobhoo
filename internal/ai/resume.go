package ai

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"
)

// ResumeExtraction is the structured, factual breakdown JOBHOO's AI produces
// from a candidate's uploaded resume file. It is the PRIMARY source of truth
// for candidate profiling: every AI feature that compares a candidate to a
// job (ranking, match explanations, recommendations) is built from this,
// not from what the candidate manually typed into a form field. The
// candidate's separately entered "skills" profile field is only ever a
// secondary, lower-weight cross-check against what's here.
type ResumeExtraction struct {
	Headline       string   `json:"headline"`
	Experience     []string `json:"experience"`
	Education      []string `json:"education"`
	Skills         []string `json:"skills"`
	Achievements   []string `json:"achievements"`
	CandidateStory string   `json:"candidate_story"`
}

// resumeExtractionMarker distinguishes an AI-generated extraction (stored as
// JSON) from resume text a candidate may have typed manually before this
// feature existed — both live in the same candidate_profiles.resume_text
// column, so no database schema change was needed to add this feature.
const resumeExtractionMarker = "jobhoo_resume_extraction_v1"

type storedResumeExtraction struct {
	Marker string `json:"_marker"`
	ResumeExtraction
}

// MarshalResumeExtraction serializes an extraction for storage in the
// candidate_profiles.resume_text column.
func MarshalResumeExtraction(e ResumeExtraction) string {
	data, err := json.Marshal(storedResumeExtraction{Marker: resumeExtractionMarker, ResumeExtraction: e})
	if err != nil {
		return ""
	}
	return string(data)
}

// ParseStoredResumeText attempts to read a stored resume_text value back into
// a structured ResumeExtraction. ok is false when the value predates this
// feature (plain candidate-typed text) or is empty — callers should fall
// back to treating raw as plain text in that case.
func ParseStoredResumeText(raw string) (ResumeExtraction, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw[0] != '{' {
		return ResumeExtraction{}, false
	}
	var wrapped storedResumeExtraction
	if err := json.Unmarshal([]byte(raw), &wrapped); err != nil || wrapped.Marker != resumeExtractionMarker {
		return ResumeExtraction{}, false
	}
	return wrapped.ResumeExtraction, true
}

// FormatResumeExtractionText renders a structured extraction as labeled
// plain text, suitable for embedding as resume context in other AI prompts
// (ranking, match explanations, recommendations).
func FormatResumeExtractionText(e ResumeExtraction) string {
	var b strings.Builder
	if e.Headline != "" {
		fmt.Fprintf(&b, "Headline: %s\n", e.Headline)
	}
	writeResumeList(&b, "Experience", e.Experience)
	writeResumeList(&b, "Education", e.Education)
	if len(e.Skills) > 0 {
		fmt.Fprintf(&b, "Skills (from resume): %s\n", strings.Join(e.Skills, ", "))
	}
	writeResumeList(&b, "Achievements", e.Achievements)
	if e.CandidateStory != "" {
		fmt.Fprintf(&b, "Candidate Story: %s\n", e.CandidateStory)
	}
	return strings.TrimSpace(b.String())
}

func writeResumeList(b *strings.Builder, label string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "%s:\n", label)
	for _, item := range items {
		fmt.Fprintf(b, "- %s\n", item)
	}
}

var xmlTagPattern = regexp.MustCompile(`<[^>]+>`)

// ExtractDocxText pulls the plain text body out of a .docx file (a ZIP
// archive containing word/document.xml) without needing an external
// dependency. It's intentionally simple — strip XML tags and decode
// entities — good enough to hand to the AI for structured extraction, not
// meant to preserve formatting.
func ExtractDocxText(data []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("reading docx as zip: %w", err)
	}
	var raw []byte
	for _, f := range zr.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("opening document.xml: %w", err)
		}
		raw, err = io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return "", fmt.Errorf("reading document.xml: %w", err)
		}
		break
	}
	if raw == nil {
		return "", errors.New("docx file has no word/document.xml — is it a valid Word document?")
	}

	text := string(raw)
	text = strings.ReplaceAll(text, "</w:p>", "\n")
	text = strings.ReplaceAll(text, "<w:br/>", "\n")
	text = strings.ReplaceAll(text, "<w:tab/>", "\t")
	text = xmlTagPattern.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("no readable text found in docx file")
	}
	return text, nil
}
