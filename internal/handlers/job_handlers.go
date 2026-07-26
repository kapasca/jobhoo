package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"jobhoo/internal/middleware"
	"jobhoo/internal/repository"
)

func (a *App) JobsListPage(w http.ResponseWriter, r *http.Request) {
	const pageSize = 12
	page := parsePageParam(r, "page")
	total, err := a.Jobs.CountOpen()
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat lowongan")
		return
	}
	pagination := buildPagination(total, page, pageSize)

	jobs, err := a.Jobs.ListOpenPaginated(pageSize, paginationOffset(pagination.Page, pageSize))
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat lowongan")
		return
	}
	data := a.newPageData(r, "Lowongan")
	data["Jobs"] = jobs
	data["Pagination"] = pagination
	a.Render.Render(w, "jobs_list.html", data)
}

func jobIDFromPath(r *http.Request, prefix string) (int64, bool) {
	path := strings.TrimPrefix(r.URL.Path, prefix)
	path = strings.TrimSuffix(path, "/")
	// take the segment up to the next slash, if any
	if idx := strings.Index(path, "/"); idx != -1 {
		path = path[:idx]
	}
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (a *App) JobDetailPage(w http.ResponseWriter, r *http.Request) {
	id, ok := jobIDFromPath(r, "/jobs/")
	if !ok {
		a.renderError(w, r, http.StatusNotFound, "Lowongan tidak ditemukan")
		return
	}

	job, err := a.Jobs.GetByID(id)
	if err != nil {
		a.renderError(w, r, http.StatusNotFound, "Lowongan tidak ditemukan")
		return
	}

	data := a.newPageData(r, job.Title)
	data["Job"] = job

	user := middleware.CurrentUser(r)
	if user != nil {
		applied, _ := a.Applications.HasApplied(job.ID, user.ID)
		data["HasApplied"] = applied
	}

	a.Render.Render(w, "job_detail.html", data)
}

func (a *App) JobApplySubmit(w http.ResponseWriter, r *http.Request) {
	id, ok := jobIDFromPath(r, "/jobs/")
	if !ok {
		a.renderError(w, r, http.StatusNotFound, "Lowongan tidak ditemukan")
		return
	}

	user := middleware.CurrentUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	job, err := a.Jobs.GetByID(id)
	if err != nil {
		a.renderError(w, r, http.StatusNotFound, "Lowongan tidak ditemukan")
		return
	}
	if job.Status != "open" {
		http.Redirect(w, r, "/jobs/"+strconv.FormatInt(job.ID, 10), http.StatusSeeOther)
		return
	}

	already, _ := a.Applications.HasApplied(job.ID, user.ID)
	if already {
		http.Redirect(w, r, "/jobs/"+strconv.FormatInt(job.ID, 10), http.StatusSeeOther)
		return
	}

	profile, err := a.Users.GetCandidateProfile(user.ID)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Profil candidate tidak ditemukan")
		return
	}

	appID, err := a.Applications.Create(job.ID, user.ID, profile.ResumePath)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal mengirim lamaran")
		return
	}

	// Run AI matching synchronously against the mock provider so the score
	// is available immediately (swap for an async job queue with a real provider later).
	result, err := a.AIMatching.Analyze(profile.ResumePath, job.Description, job.Requirements)
	if err == nil {
		_ = a.Applications.SetAIMatchResult(appID, result.MatchScore, result.SkillMatch, result.ExperienceMatch, result.EducationMatch)
	}

	http.Redirect(w, r, "/applications", http.StatusSeeOther)
}

func (a *App) CandidateDashboardPage(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	profile, err := a.Users.GetCandidateProfile(user.ID)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Profil tidak ditemukan")
		return
	}

	applications, err := a.Applications.ListByCandidate(user.ID)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat lamaran")
		return
	}

	activeCount := 0
	for _, app := range applications {
		if app.FinalStatus == nil {
			activeCount++
		}
	}

	recent := applications
	if len(recent) > 5 {
		recent = recent[:5]
	}

	data := a.newPageData(r, "Dashboard")
	data["Profile"] = profile
	data["ActiveCount"] = activeCount
	data["TotalCount"] = len(applications)
	data["RecentApplications"] = recent
	a.Render.Render(w, "candidate_dashboard.html", data)
}

func (a *App) CandidateApplicationsPage(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	const pageSize = 12
	page := parsePageParam(r, "page")
	total, err := a.Applications.CountByCandidate(user.ID)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat lamaran")
		return
	}
	pagination := buildPagination(total, page, pageSize)

	applications, err := a.Applications.ListByCandidatePaginated(user.ID, pageSize, paginationOffset(pagination.Page, pageSize))
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat lamaran")
		return
	}

	data := a.newPageData(r, "Lamaran Saya")
	data["Applications"] = applications
	data["Pagination"] = pagination
	a.Render.Render(w, "candidate_applications.html", data)
}

func (a *App) CandidateResumePage(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	profile, err := a.Users.GetCandidateProfile(user.ID)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Profil tidak ditemukan")
		return
	}
	data := a.newPageData(r, "Resume Saya")
	data["Profile"] = profile
	a.Render.Render(w, "candidate_resume.html", data)
}

func (a *App) CandidateResumeSubmit(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		a.renderError(w, r, http.StatusBadRequest, "File terlalu besar")
		return
	}

	_, header, err := r.FormFile("resume")
	if err != nil {
		a.renderError(w, r, http.StatusBadRequest, "Resume PDF wajib diupload")
		return
	}

	resumePath, resumeFilename, err := a.saveUpload(header, "resumes")
	if err != nil {
		a.renderError(w, r, http.StatusBadRequest, "Gagal mengunggah resume: "+err.Error())
		return
	}

	if err := a.Users.UpdateCandidateResume(user.ID, resumePath, resumeFilename); err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memperbarui resume")
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

var _ = repository.ErrNotFound
