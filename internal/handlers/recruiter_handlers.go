package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"jobhoo/internal/middleware"
	"jobhoo/internal/models"
	"jobhoo/internal/repository"
)

// extractIDSegment pulls the numeric id out of a path like
// "/recruiter/jobs/42/edit" given prefix "/recruiter/jobs/" -> 42
func extractIDSegment(path, prefix string) (int64, string, bool) {
	rest := strings.TrimPrefix(path, prefix)
	rest = strings.TrimSuffix(rest, "/")
	parts := strings.SplitN(rest, "/", 2)
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", false
	}
	suffix := ""
	if len(parts) > 1 {
		suffix = parts[1]
	}
	return id, suffix, true
}

func (a *App) RecruiterDashboardPage(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	activeJobs, _ := a.Jobs.CountByRecruiterAndStatus(user.ID, models.JobOpen)
	closedJobs, _ := a.Jobs.CountByRecruiterAndStatus(user.ID, models.JobClosed)
	totalApplicants, _ := a.Applications.CountByRecruiter(user.ID)
	recent, _ := a.Applications.ListByRecruiter(user.ID, 8)

	data := a.newPageData(r, "Dashboard Recruiter")
	data["ActiveJobs"] = activeJobs
	data["ClosedJobs"] = closedJobs
	data["TotalApplicants"] = totalApplicants
	data["RecentApplications"] = recent
	a.Render.Render(w, "recruiter_dashboard.html", data)
}

func (a *App) RecruiterJobsPage(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	const pageSize = 12
	page := parsePageParam(r, "page")
	total, err := a.Jobs.CountByRecruiter(user.ID)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat lowongan")
		return
	}
	pagination := buildPagination(total, page, pageSize)

	jobs, err := a.Jobs.ListByRecruiterPaginated(user.ID, pageSize, paginationOffset(pagination.Page, pageSize))
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat lowongan")
		return
	}

	counts, _ := a.Applications.CountApplicantsByJob(user.ID)
	for i := range jobs {
		jobs[i].ApplicantCount = counts[jobs[i].ID]
	}

	data := a.newPageData(r, "Lowongan Saya")
	data["Jobs"] = jobs
	data["Pagination"] = pagination
	a.Render.Render(w, "recruiter_jobs.html", data)
}

func (a *App) RecruiterJobNewPage(w http.ResponseWriter, r *http.Request) {
	data := a.newPageData(r, "Buat Lowongan")
	data["IsEdit"] = false
	data["FormAction"] = "/recruiter/jobs/new"
	a.Render.Render(w, "recruiter_job_form.html", data)
}

func parseJobInput(r *http.Request) (repository.JobInput, error) {
	closingDate, err := time.Parse("2006-01-02", r.FormValue("closing_date"))
	if err != nil {
		return repository.JobInput{}, err
	}

	var salaryMin, salaryMax *int64
	if v := strings.TrimSpace(r.FormValue("salary_min")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			salaryMin = &n
		}
	}
	if v := strings.TrimSpace(r.FormValue("salary_max")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			salaryMax = &n
		}
	}

	return repository.JobInput{
		Title:           strings.TrimSpace(r.FormValue("title")),
		Position:        strings.TrimSpace(r.FormValue("position")),
		EmploymentType:  models.EmploymentType(r.FormValue("employment_type")),
		WorkArrangement: models.WorkArrangement(r.FormValue("work_arrangement")),
		Location:        strings.TrimSpace(r.FormValue("location")),
		SalaryMin:       salaryMin,
		SalaryMax:       salaryMax,
		Benefits:        strings.TrimSpace(r.FormValue("benefits")),
		Requirements:    strings.TrimSpace(r.FormValue("requirements")),
		Description:     strings.TrimSpace(r.FormValue("description")),
		ClosingDate:     closingDate,
	}, nil
}

func (a *App) RecruiterJobCreateSubmit(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	if err := r.ParseForm(); err != nil {
		a.renderError(w, r, http.StatusBadRequest, "Form tidak valid")
		return
	}

	input, err := parseJobInput(r)
	if err != nil {
		data := a.newPageData(r, "Buat Lowongan")
		data["IsEdit"] = false
		data["FormAction"] = "/recruiter/jobs/new"
		data["Error"] = "Tanggal closing date tidak valid."
		a.Render.Render(w, "recruiter_job_form.html", data)
		return
	}

	id, err := a.Jobs.Create(user.ID, input)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal membuat lowongan")
		return
	}

	http.Redirect(w, r, "/recruiter/jobs/"+strconv.FormatInt(id, 10)+"/applicants", http.StatusSeeOther)
}

func (a *App) RecruiterJobEditPage(w http.ResponseWriter, r *http.Request) {
	id, _, ok := extractIDSegment(r.URL.Path, "/recruiter/jobs/")
	if !ok {
		a.renderError(w, r, http.StatusNotFound, "Lowongan tidak ditemukan")
		return
	}

	job, err := a.Jobs.GetByID(id)
	if err != nil {
		a.renderError(w, r, http.StatusNotFound, "Lowongan tidak ditemukan")
		return
	}

	user := middleware.CurrentUser(r)
	if job.RecruiterID != user.ID {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}

	data := a.newPageData(r, "Edit Lowongan")
	data["IsEdit"] = true
	data["Job"] = job
	data["FormAction"] = "/recruiter/jobs/" + strconv.FormatInt(job.ID, 10) + "/edit"
	a.Render.Render(w, "recruiter_job_form.html", data)
}

func (a *App) RecruiterJobEditSubmit(w http.ResponseWriter, r *http.Request) {
	id, _, ok := extractIDSegment(r.URL.Path, "/recruiter/jobs/")
	if !ok {
		a.renderError(w, r, http.StatusNotFound, "Lowongan tidak ditemukan")
		return
	}

	job, err := a.Jobs.GetByID(id)
	if err != nil {
		a.renderError(w, r, http.StatusNotFound, "Lowongan tidak ditemukan")
		return
	}
	user := middleware.CurrentUser(r)
	if job.RecruiterID != user.ID {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		a.renderError(w, r, http.StatusBadRequest, "Form tidak valid")
		return
	}

	input, err := parseJobInput(r)
	if err != nil {
		data := a.newPageData(r, "Edit Lowongan")
		data["IsEdit"] = true
		data["Job"] = job
		data["FormAction"] = "/recruiter/jobs/" + strconv.FormatInt(job.ID, 10) + "/edit"
		data["Error"] = "Tanggal closing date tidak valid."
		a.Render.Render(w, "recruiter_job_form.html", data)
		return
	}

	if err := a.Jobs.Update(job.ID, input); err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memperbarui lowongan")
		return
	}

	http.Redirect(w, r, "/recruiter/jobs", http.StatusSeeOther)
}

func (a *App) RecruiterJobCloseSubmit(w http.ResponseWriter, r *http.Request) {
	id, _, ok := extractIDSegment(r.URL.Path, "/recruiter/jobs/")
	if !ok {
		a.renderError(w, r, http.StatusNotFound, "Lowongan tidak ditemukan")
		return
	}

	job, err := a.Jobs.GetByID(id)
	if err != nil {
		a.renderError(w, r, http.StatusNotFound, "Lowongan tidak ditemukan")
		return
	}
	user := middleware.CurrentUser(r)
	if job.RecruiterID != user.ID {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}

	if err := a.Jobs.Close(job.ID); err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal menutup lowongan")
		return
	}

	http.Redirect(w, r, "/recruiter/jobs/"+strconv.FormatInt(job.ID, 10)+"/applicants", http.StatusSeeOther)
}

// StageColumn groups applications by ATS stage for the kanban board template.
type StageColumn struct {
	Stage        models.ApplicationStage
	Applications []models.Application
}

func (a *App) RecruiterJobApplicantsPage(w http.ResponseWriter, r *http.Request) {
	id, _, ok := extractIDSegment(r.URL.Path, "/recruiter/jobs/")
	if !ok {
		a.renderError(w, r, http.StatusNotFound, "Lowongan tidak ditemukan")
		return
	}

	job, err := a.Jobs.GetByID(id)
	if err != nil {
		a.renderError(w, r, http.StatusNotFound, "Lowongan tidak ditemukan")
		return
	}
	user := middleware.CurrentUser(r)
	if job.RecruiterID != user.ID {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}

	applicants, err := a.Applications.ListByJob(job.ID)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat pelamar")
		return
	}

	data := a.newPageData(r, job.Title+" · Pelamar")
	data["Job"] = job
	data["Applicants"] = applicants
	data["StageColumns"] = groupByStage(applicants)
	a.Render.Render(w, "recruiter_ats.html", data)
}

func groupByStage(applicants []models.Application) []StageColumn {
	columns := make([]StageColumn, 0, len(models.ATSStages))
	for _, stage := range models.ATSStages {
		col := StageColumn{Stage: stage}
		for _, app := range applicants {
			if app.Stage == stage {
				col.Applications = append(col.Applications, app)
			}
		}
		columns = append(columns, col)
	}
	return columns
}

// RecruiterApplicationStageUpdate handles the htmx select-driven stage change
// on the ATS board and re-renders just the board fragment.
func (a *App) RecruiterApplicationStageUpdate(w http.ResponseWriter, r *http.Request) {
	id, _, ok := extractIDSegment(r.URL.Path, "/recruiter/applications/")
	if !ok {
		a.renderError(w, r, http.StatusNotFound, "Lamaran tidak ditemukan")
		return
	}

	app, err := a.Applications.GetByID(id)
	if err != nil {
		a.renderError(w, r, http.StatusNotFound, "Lamaran tidak ditemukan")
		return
	}

	job, err := a.Jobs.GetByID(app.JobID)
	if err != nil {
		a.renderError(w, r, http.StatusNotFound, "Lowongan tidak ditemukan")
		return
	}
	user := middleware.CurrentUser(r)
	if job.RecruiterID != user.ID {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}

	newStage := models.ApplicationStage(r.FormValue("stage"))
	if err := a.Applications.UpdateStage(app.ID, newStage); err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memperbarui status")
		return
	}

	a.renderATSBoard(w, r, job.ID)
}

// RecruiterApplicationFinalStatusUpdate is used once a job is closed: every
// application must end up with one of the three required final statuses.
func (a *App) RecruiterApplicationFinalStatusUpdate(w http.ResponseWriter, r *http.Request) {
	id, _, ok := extractIDSegment(r.URL.Path, "/recruiter/applications/")
	if !ok {
		a.renderError(w, r, http.StatusNotFound, "Lamaran tidak ditemukan")
		return
	}

	app, err := a.Applications.GetByID(id)
	if err != nil {
		a.renderError(w, r, http.StatusNotFound, "Lamaran tidak ditemukan")
		return
	}

	job, err := a.Jobs.GetByID(app.JobID)
	if err != nil {
		a.renderError(w, r, http.StatusNotFound, "Lowongan tidak ditemukan")
		return
	}
	user := middleware.CurrentUser(r)
	if job.RecruiterID != user.ID {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}

	status := models.ApplicationFinalStatus(r.FormValue("final_status"))
	if status != "" {
		if err := a.Applications.SetFinalStatus(app.ID, status); err != nil {
			a.renderError(w, r, http.StatusInternalServerError, "Gagal memperbarui status akhir")
			return
		}
	}

	a.renderATSBoard(w, r, job.ID)
}

func (a *App) renderATSBoard(w http.ResponseWriter, r *http.Request, jobID int64) {
	job, err := a.Jobs.GetByID(jobID)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat lowongan")
		return
	}
	applicants, err := a.Applications.ListByJob(jobID)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat pelamar")
		return
	}

	data := PageData{
		"Job":          job,
		"Applicants":   applicants,
		"StageColumns": groupByStage(applicants),
	}
	a.Render.RenderPartial(w, "recruiter_ats.html", "ats_board", data)
}
