package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/middleware"
	"github.com/jobhoo/jobhoo/internal/models"
)

type postJobData struct {
	BasePageData
	Error      string
	Categories []struct {
		Value models.JobCategory
		Label string
	}
	Countries []models.CountryEntry
	// IsEdit + fields below are populated when rendering the edit form with
	// an existing job's current values; PostJobPage leaves them zero.
	IsEdit           bool
	JobID            string
	Title            string
	Description      string
	Country          string
	State            string
	EmploymentType   string
	WorkArrangement  string
	Category         string
	Seniority        string
	Currency         string
	SalaryMin        string
	SalaryMax        string
	MustHaveSkills   string
	NiceToHaveSkills string
	OpensAt          string
	ClosesAt         string
}

func newPostJobData(r *http.Request, nav string) postJobData {
	return postJobData{
		BasePageData: newBasePageData(r, nav),
		Categories:   models.JobCategories,
		Countries:    models.Countries,
	}
}

func (h *Handlers) PostJobPage(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireApprovedCompany(w, r); !ok {
		return
	}
	h.Render.Render(w, http.StatusOK, "post-job.html", newPostJobData(r, "post"))
}

func (h *Handlers) PostJob(w http.ResponseWriter, r *http.Request) {
	company, ok := h.requireApprovedCompany(w, r)
	if !ok {
		return
	}
	user := middleware.CurrentUser(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	data := newPostJobData(r, "post")

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	if title == "" || description == "" {
		data.Error = "Title and description are required."
		h.Render.Render(w, http.StatusBadRequest, "post-job.html", data)
		return
	}

	opensAt, closesAt, err := parseScheduleWindow(r.FormValue("opens_at"), r.FormValue("closes_at"))
	if err != nil {
		data.Error = err.Error()
		h.Render.Render(w, http.StatusBadRequest, "post-job.html", data)
		return
	}

	_, err = h.Jobs.Create(r.Context(), database.CreateJobInput{
		CompanyID:        company.ID,
		CreatedBy:        user.ID,
		Title:            title,
		Description:      description,
		Country:          strings.TrimSpace(r.FormValue("country")),
		State:            strings.TrimSpace(r.FormValue("state")),
		EmploymentType:   models.EmploymentType(r.FormValue("employment_type")),
		WorkArrangement:  models.WorkArrangement(r.FormValue("work_arrangement")),
		Category:         models.JobCategory(r.FormValue("category")),
		Seniority:        senioritOrDefault(r.FormValue("seniority")),
		SalaryMin:        parseOptionalInt(r.FormValue("salary_min")),
		SalaryMax:        parseOptionalInt(r.FormValue("salary_max")),
		SalaryCurrency:   strings.TrimSpace(r.FormValue("salary_currency")),
		MustHaveSkills:   splitSkills(r.FormValue("must_have_skills")),
		NiceToHaveSkills: splitSkills(r.FormValue("nice_to_have_skills")),
		OpensAt:          opensAt,
		ClosesAt:         closesAt,
	})
	if err != nil {
		data.Error = "Something went wrong publishing this job. Please try again."
		h.Render.Render(w, http.StatusBadRequest, "post-job.html", data)
		return
	}

	http.Redirect(w, r, "/dashboard/recruiter", http.StatusSeeOther)
}

// jobOwnedByRequestingRecruiter loads a job and verifies the current
// recruiter's company owns it — shared by edit/close/archive/reopen so one
// recruiter can never manage another company's job by guessing its ID.
func (h *Handlers) jobOwnedByRequestingRecruiter(w http.ResponseWriter, r *http.Request) (*models.Job, bool) {
	id := chi.URLParam(r, "id")
	job, err := h.Jobs.GetByID(r.Context(), id)
	if err != nil {
		h.NotFoundPage(w, r)
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

func (h *Handlers) EditJobPage(w http.ResponseWriter, r *http.Request) {
	job, ok := h.jobOwnedByRequestingRecruiter(w, r)
	if !ok {
		return
	}
	h.Render.Render(w, http.StatusOK, "post-job.html", jobToPostJobData(r, job))
}

func (h *Handlers) EditJob(w http.ResponseWriter, r *http.Request) {
	job, ok := h.jobOwnedByRequestingRecruiter(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	data := jobToPostJobData(r, job)
	data.Title = strings.TrimSpace(r.FormValue("title"))
	data.Description = strings.TrimSpace(r.FormValue("description"))
	if data.Title == "" || data.Description == "" {
		data.Error = "Title and description are required."
		h.Render.Render(w, http.StatusBadRequest, "post-job.html", data)
		return
	}

	opensAt, closesAt, err := parseScheduleWindow(r.FormValue("opens_at"), r.FormValue("closes_at"))
	if err != nil {
		data.Error = err.Error()
		h.Render.Render(w, http.StatusBadRequest, "post-job.html", data)
		return
	}

	err = h.Jobs.Update(r.Context(), job.ID, database.UpdateJobInput{
		Title:            data.Title,
		Description:      data.Description,
		Country:          strings.TrimSpace(r.FormValue("country")),
		State:            strings.TrimSpace(r.FormValue("state")),
		EmploymentType:   models.EmploymentType(r.FormValue("employment_type")),
		WorkArrangement:  models.WorkArrangement(r.FormValue("work_arrangement")),
		Category:         models.JobCategory(r.FormValue("category")),
		Seniority:        senioritOrDefault(r.FormValue("seniority")),
		SalaryMin:        parseOptionalInt(r.FormValue("salary_min")),
		SalaryMax:        parseOptionalInt(r.FormValue("salary_max")),
		SalaryCurrency:   strings.TrimSpace(r.FormValue("salary_currency")),
		MustHaveSkills:   splitSkills(r.FormValue("must_have_skills")),
		NiceToHaveSkills: splitSkills(r.FormValue("nice_to_have_skills")),
		OpensAt:          opensAt,
		ClosesAt:         closesAt,
	})
	if err != nil {
		data.Error = "Something went wrong saving this job. Please try again."
		h.Render.Render(w, http.StatusBadRequest, "post-job.html", data)
		return
	}

	http.Redirect(w, r, "/dashboard/recruiter", http.StatusSeeOther)
}

// CloseJob, ArchiveJob, and ReopenJob are simple one-click status changes
// from the recruiter dashboard job list.
func (h *Handlers) CloseJob(w http.ResponseWriter, r *http.Request) {
	h.setJobStatus(w, r, models.JobClosed)
}
func (h *Handlers) ArchiveJob(w http.ResponseWriter, r *http.Request) {
	job, ok := h.jobOwnedByRequestingRecruiter(w, r)
	if !ok {
		return
	}
	// Always ensure a published job is closed before it is archived.
	if job.Status == models.JobPublished {
		if err := h.Jobs.UpdateStatus(r.Context(), job.ID, models.JobClosed); err != nil {
			http.Error(w, "could not close job", http.StatusInternalServerError)
			return
		}
	}
	if err := h.Jobs.UpdateStatus(r.Context(), job.ID, models.JobArchived); err != nil {
		http.Error(w, "could not archive job", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/dashboard/recruiter", http.StatusSeeOther)
}
func (h *Handlers) ReopenJob(w http.ResponseWriter, r *http.Request) {
	h.setJobStatus(w, r, models.JobPublished)
}

func (h *Handlers) setJobStatus(w http.ResponseWriter, r *http.Request, status models.JobStatus) {
	job, ok := h.jobOwnedByRequestingRecruiter(w, r)
	if !ok {
		return
	}
	if err := h.Jobs.UpdateStatus(r.Context(), job.ID, status); err != nil {
		http.Error(w, "could not update job status", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/dashboard/recruiter", http.StatusSeeOther)
}

type recruiterJobRow struct {
	models.Job
	ApplicantCount int
}

type recruiterDashboardData struct {
	BasePageData
	Company *models.Company
	Jobs    []recruiterJobRow
}

func (h *Handlers) RecruiterDashboard(w http.ResponseWriter, r *http.Request) {
	company, ok := h.requireCompany(w, r)
	if !ok {
		return
	}

	jobs, err := h.Jobs.ListByCompany(r.Context(), company.ID)
	if err != nil {
		http.Error(w, "could not load jobs", http.StatusInternalServerError)
		return
	}

	rows := make([]recruiterJobRow, 0, len(jobs))
	for _, j := range jobs {
		count, _ := h.Applications.CountByJob(r.Context(), j.ID)
		rows = append(rows, recruiterJobRow{Job: j, ApplicantCount: count})
	}

	h.Render.Render(w, http.StatusOK, "recruiter-dashboard.html", recruiterDashboardData{
		BasePageData: newBasePageData(r, "dashboard"),
		Company:      company,
		Jobs:         rows,
	})
}

func parseOptionalInt(s string) *int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &n
}

func splitSkills(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// datetimeLocalLayout matches the value format of an <input type="datetime-local">.
const datetimeLocalLayout = "2006-01-02T15:04"

// parseScheduleWindow parses the post-job form's optional opens_at/closes_at
// fields and validates that closes_at (if both are set) is after opens_at.
func parseScheduleWindow(opensRaw, closesRaw string) (opensAt, closesAt *time.Time, err error) {
	opensAt, err = parseOptionalDatetimeLocal(opensRaw)
	if err != nil {
		return nil, nil, errFieldError("Invalid open date/time.")
	}
	closesAt, err = parseOptionalDatetimeLocal(closesRaw)
	if err != nil {
		return nil, nil, errFieldError("Invalid close date/time.")
	}
	if opensAt != nil && closesAt != nil && !closesAt.After(*opensAt) {
		return nil, nil, errFieldError("Close date must be after the open date.")
	}
	return opensAt, closesAt, nil
}

func parseOptionalDatetimeLocal(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation(datetimeLocalLayout, raw, time.Local)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

type errFieldError string

func (e errFieldError) Error() string { return string(e) }

// jobToPostJobData pre-fills the shared post-job form template with an
// existing job's values, for the edit-job page.
func jobToPostJobData(r *http.Request, job *models.Job) postJobData {
	data := newPostJobData(r, "post")
	data.IsEdit = true
	data.JobID = job.ID
	data.Title = job.Title
	data.Description = job.Description
	data.Country = job.Country
	data.State = job.State
	data.EmploymentType = string(job.EmploymentType)
	data.WorkArrangement = string(job.WorkArrangement)
	data.Category = string(job.Category)
	data.Seniority = job.Seniority
	data.Currency = job.SalaryCurrency
	data.MustHaveSkills = strings.Join(job.MustHaveSkills, ", ")
	data.NiceToHaveSkills = strings.Join(job.NiceToHaveSkills, ", ")
	if job.SalaryMin != nil {
		data.SalaryMin = strconv.Itoa(*job.SalaryMin)
	}
	if job.SalaryMax != nil {
		data.SalaryMax = strconv.Itoa(*job.SalaryMax)
	}
	if job.OpensAt != nil {
		data.OpensAt = job.OpensAt.Local().Format(datetimeLocalLayout)
	}
	if job.ClosesAt != nil {
		data.ClosesAt = job.ClosesAt.Local().Format(datetimeLocalLayout)
	}
	return data
}

func senioritOrDefault(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "General"
	}
	return s
}
