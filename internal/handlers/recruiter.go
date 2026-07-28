package handlers

import (
	"net/http"
	"strconv"
	"strings"

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
}

func (h *Handlers) PostJobPage(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireCompany(w, r); !ok {
		return
	}
	h.Render.Render(w, http.StatusOK, "post-job.html", postJobData{
		BasePageData: newBasePageData(r, "post"),
		Categories:   models.JobCategories,
	})
}

func (h *Handlers) PostJob(w http.ResponseWriter, r *http.Request) {
	company, ok := h.requireCompany(w, r)
	if !ok {
		return
	}
	user := middleware.CurrentUser(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	data := postJobData{BasePageData: newBasePageData(r, "post"), Categories: models.JobCategories}

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	if title == "" || description == "" {
		data.Error = "Title and description are required."
		h.Render.Render(w, http.StatusBadRequest, "post-job.html", data)
		return
	}

	salaryMin := parseOptionalInt(r.FormValue("salary_min"))
	salaryMax := parseOptionalInt(r.FormValue("salary_max"))

	_, err := h.Jobs.Create(r.Context(), database.CreateJobInput{
		CompanyID:        company.ID,
		CreatedBy:        user.ID,
		Title:            title,
		Description:      description,
		Location:         strings.TrimSpace(r.FormValue("location")),
		EmploymentType:   models.EmploymentType(r.FormValue("employment_type")),
		WorkArrangement:  models.WorkArrangement(r.FormValue("work_arrangement")),
		Category:         models.JobCategory(r.FormValue("category")),
		Seniority:        strings.TrimSpace(r.FormValue("seniority")),
		SalaryMin:        salaryMin,
		SalaryMax:        salaryMax,
		MustHaveSkills:   splitSkills(r.FormValue("must_have_skills")),
		NiceToHaveSkills: splitSkills(r.FormValue("nice_to_have_skills")),
	})
	if err != nil {
		data.Error = "Something went wrong publishing this job. Please try again."
		h.Render.Render(w, http.StatusBadRequest, "post-job.html", data)
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
