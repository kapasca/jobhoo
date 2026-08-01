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

	"github.com/go-chi/chi/v5"
	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/middleware"
	"github.com/jobhoo/jobhoo/internal/models"
)

type companySetupData struct {
	BasePageData
	Error       string
	Name        string
	Website     string
	Industry    string
	Description string
	LogoURL     string
	IsEdit      bool
	Incomplete  bool // true when redirected here because profile is incomplete
}

func (h *Handlers) CompanySetupPage(w http.ResponseWriter, r *http.Request) {
	h.Render.Render(w, http.StatusOK, "company-setup.html", companySetupData{BasePageData: newBasePageData(r, "post")})
}

func (h *Handlers) CompanySetup(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	r.Body = http.MaxBytesReader(w, r.Body, logoMaxBytes+4096)
	if err := r.ParseMultipartForm(logoMaxBytes); err != nil {
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}

	data := companySetupData{
		BasePageData: newBasePageData(r, "post"),
		Name:         strings.TrimSpace(r.FormValue("name")),
		Website:      strings.TrimSpace(r.FormValue("website")),
		Industry:     strings.TrimSpace(r.FormValue("industry")),
		Description:  strings.TrimSpace(r.FormValue("description")),
	}

	if data.Name == "" {
		data.Error = "Company name is required."
		h.Render.Render(w, http.StatusBadRequest, "company-setup.html", data)
		return
	}

	if err := validateWebsite(data.Website); err != nil {
		data.Error = err.Error()
		h.Render.Render(w, http.StatusBadRequest, "company-setup.html", data)
		return
	}

	data.Website = normalizeWebsiteURL(data.Website)

	logoURL, err := handleLogoUpload(r)
	if err != nil {
		data.Error = err.Error()
		h.Render.Render(w, http.StatusBadRequest, "company-setup.html", data)
		return
	}

	if _, err := h.Companies.Create(r.Context(), user.ID, data.Name, data.Website, data.Description, data.Industry, logoURL); err != nil {
		data.Error = "Something went wrong creating your company. Please try again."
		h.Render.Render(w, http.StatusBadRequest, "company-setup.html", data)
		return
	}

	// New companies start 'pending' — send them to their dashboard, which
	// shows the pending-approval banner, rather than straight to /post-job
	// (which they can't use yet; see requireApprovedCompany).
	http.Redirect(w, r, "/dashboard/recruiter", http.StatusSeeOther)
}

// CompanyProfilePage / CompanyProfileUpdate let a recruiter edit their
// company's public-facing profile at any time — including while pending —
// since a company under review will often need to fix its profile before
// an admin approves it.
func (h *Handlers) CompanyProfilePage(w http.ResponseWriter, r *http.Request) {
	company, ok := h.requireCompany(w, r)
	if !ok {
		return
	}
	h.Render.Render(w, http.StatusOK, "company-setup.html", companySetupData{
		BasePageData: newBasePageData(r, "profile"),
		Name:         company.Name,
		Website:      company.Website,
		Industry:     company.Industry,
		Description:  company.Description,
		LogoURL:      company.LogoURL,
		IsEdit:       true,
		Incomplete:   r.URL.Query().Get("incomplete") == "1",
	})
}

// CompanyPublicRedirect sends the recruiter straight to their company’s
// public listing page without needing to know their own company ID.
func (h *Handlers) CompanyPublicRedirect(w http.ResponseWriter, r *http.Request) {
	company, ok := h.requireCompany(w, r)
	if !ok {
		return
	}
	http.Redirect(w, r, "/companies/"+company.ID, http.StatusSeeOther)
}

const (
	logoUploadsDir = "web/static/uploads/logos"
	logoURLPrefix  = "/static/uploads/logos/"
	logoMaxBytes   = 2 << 20 // 2 MB
)

// handleLogoUpload reads the "logo_file" multipart field, validates its MIME
// type (JPEG/PNG/GIF/WebP only), and writes it to logoUploadsDir with a
// cryptographically-random filename. Returns the public URL path, or "" if
// no file was uploaded, or an error if the file was invalid or unwritable.
func handleLogoUpload(r *http.Request) (string, error) {
	file, _, err := r.FormFile("logo_file")
	if err == http.ErrMissingFile {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading upload: %w", err)
	}
	defer file.Close()

	// Detect MIME type from actual file header, not the client-supplied name.
	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("reading file header: %w", err)
	}
	mimeType := http.DetectContentType(header[:n])

	var ext string
	switch mimeType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	case "image/webp":
		ext = ".webp"
	default:
		return "", fmt.Errorf("unsupported file type — please upload a JPEG, PNG, GIF, or WebP image")
	}

	var rndBytes [16]byte
	if _, err := rand.Read(rndBytes[:]); err != nil {
		return "", fmt.Errorf("generating filename: %w", err)
	}
	filename := hex.EncodeToString(rndBytes[:]) + ext

	if err := os.MkdirAll(logoUploadsDir, 0o755); err != nil {
		return "", fmt.Errorf("creating upload directory: %w", err)
	}
	dst, err := os.Create(filepath.Join(logoUploadsDir, filename))
	if err != nil {
		return "", fmt.Errorf("saving file: %w", err)
	}
	defer dst.Close()

	if _, err := dst.Write(header[:n]); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return logoURLPrefix + filename, nil
}

// validateWebsite checks if website contains at least one dot (.) as domain separator.
// Returns nil if valid, or an error message if invalid.
func validateWebsite(website string) error {
	if website == "" {
		// website is optional
		return nil
	}
	if !strings.Contains(website, ".") {
		return fmt.Errorf("website must contain a valid domain name (with a dot)")
	}
	return nil
}

// normalizeWebsiteURL ensures website URL has a protocol (http:// or https://).
// If the URL doesn't start with http:// or https://, adds https:// as default.
// Returns the normalized URL, or empty string if input is empty.
func normalizeWebsiteURL(website string) string {
	if website == "" {
		return ""
	}
	if strings.HasPrefix(website, "http://") || strings.HasPrefix(website, "https://") {
		return website
	}
	return "https://" + website
}

func (h *Handlers) CompanyProfileUpdate(w http.ResponseWriter, r *http.Request) {
	company, ok := h.requireCompany(w, r)
	if !ok {
		return
	}
	// Cap request body at 2MB logo + small form fields overhead.
	r.Body = http.MaxBytesReader(w, r.Body, logoMaxBytes+4096)
	if err := r.ParseMultipartForm(logoMaxBytes); err != nil {
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}

	data := companySetupData{
		BasePageData: newBasePageData(r, "dashboard"),
		Name:         strings.TrimSpace(r.FormValue("name")),
		Website:      strings.TrimSpace(r.FormValue("website")),
		Industry:     strings.TrimSpace(r.FormValue("industry")),
		Description:  strings.TrimSpace(r.FormValue("description")),
		LogoURL:      company.LogoURL, // keep existing logo by default
		IsEdit:       true,
	}

	if data.Name == "" {
		data.Error = "Company name is required."
		h.Render.Render(w, http.StatusBadRequest, "company-setup.html", data)
		return
	}

	if err := validateWebsite(data.Website); err != nil {
		data.Error = err.Error()
		h.Render.Render(w, http.StatusBadRequest, "company-setup.html", data)
		return
	}

	data.Website = normalizeWebsiteURL(data.Website)

	uploadedURL, err := handleLogoUpload(r)
	if err != nil {
		data.Error = err.Error()
		h.Render.Render(w, http.StatusBadRequest, "company-setup.html", data)
		return
	}
	if uploadedURL != "" {
		data.LogoURL = uploadedURL
	}

	if err := h.Companies.Update(r.Context(), company.ID, data.Name, data.Website, data.Description, data.Industry, data.LogoURL); err != nil {
		data.Error = "Something went wrong saving your profile. Please try again."
		h.Render.Render(w, http.StatusBadRequest, "company-setup.html", data)
		return
	}

	http.Redirect(w, r, "/dashboard/recruiter", http.StatusSeeOther)
}

// requireCompany fetches the current recruiter's company. If they don't have
// one yet, it redirects to company setup and returns ok=false — callers
// must return immediately in that case, since the response is already
// written. Does NOT check approval status — use requireApprovedCompany for
// actions (like posting a job) that must wait for admin review.
func (h *Handlers) requireCompany(w http.ResponseWriter, r *http.Request) (company *models.Company, ok bool) {
	user := middleware.CurrentUser(r)
	company, err := h.Companies.GetByOwnerID(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.Redirect(w, r, "/company/setup", http.StatusSeeOther)
			return nil, false
		}
		http.Error(w, "could not load company", http.StatusInternalServerError)
		return nil, false
	}
	return company, true
}

// requireApprovedCompany is requireCompany plus an approval-status gate, for
// actions that must wait on admin review (posting a job). Recruiters with a
// pending/rejected company are redirected to their dashboard, which shows
// the relevant status banner and explains why the action is unavailable.
func (h *Handlers) requireApprovedCompany(w http.ResponseWriter, r *http.Request) (company *models.Company, ok bool) {
	company, ok = h.requireCompany(w, r)
	if !ok {
		return nil, false
	}
	if company.Status != models.CompanyApproved {
		http.Redirect(w, r, "/dashboard/recruiter", http.StatusSeeOther)
		return nil, false
	}
	if !company.IsProfileComplete() {
		http.Redirect(w, r, "/company/profile?incomplete=1", http.StatusSeeOther)
		return nil, false
	}
	return company, true
}

// CompanyResubmit lets a recruiter with a rejected company resubmit for
// admin review after fixing their profile. Blacklisted companies are
// blocked — they cannot resubmit.
func (h *Handlers) CompanyResubmit(w http.ResponseWriter, r *http.Request) {
	company, ok := h.requireCompany(w, r)
	if !ok {
		return
	}
	if company.Status != models.CompanyRejected {
		http.Redirect(w, r, "/dashboard/recruiter", http.StatusSeeOther)
		return
	}
	if err := h.Companies.Resubmit(r.Context(), company.ID); err != nil {
		http.Error(w, "could not resubmit company", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/dashboard/recruiter", http.StatusSeeOther)
}

// CompanyPublicDetail is the public company profile page — no auth
// required. Shows the company's info plus its currently open jobs.
// Pending/rejected companies 404 here (same as an unapproved job wouldn't
// appear in search) so a company can't be "previewed" publicly before
// admin approval.
func (h *Handlers) CompanyPublicDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	company, err := h.Companies.GetByID(r.Context(), id)
	if err != nil || company.Status != models.CompanyApproved {
		h.NotFoundPage(w, r)
		return
	}

	jobs, err := h.Jobs.ListByCompany(r.Context(), company.ID)
	if err != nil {
		http.Error(w, "could not load jobs", http.StatusInternalServerError)
		return
	}
	// Public visitors only see currently-open jobs, not drafts/closed/
	// archived/scheduled-future ones that ListByCompany also returns for
	// the recruiter's own dashboard.
	var openJobs []models.Job
	for _, j := range jobs {
		if j.Status == models.JobPublished && !j.IsScheduled() && !j.IsExpired() {
			openJobs = append(openJobs, j)
		}
	}

	h.Render.Render(w, http.StatusOK, "company-detail.html", struct {
		BasePageData
		Company models.Company
		Jobs    []models.Job
	}{
		BasePageData: newBasePageData(r, "companies"),
		Company:      *company,
		Jobs:         openJobs,
	})
}
