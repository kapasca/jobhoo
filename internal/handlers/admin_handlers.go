package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"jobhoo/internal/middleware"
	"jobhoo/internal/models"
)

func (a *App) AdminDashboardPage(w http.ResponseWriter, r *http.Request) {
	pendingCount, _ := a.Users.CountRecruitersByStatus(models.RecruiterPending)
	totalRecruiters, _ := a.Users.CountRecruiters()
	candidates, _ := a.Users.CountCandidates()
	jobsCount, err := a.Jobs.CountAll()
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat data")
		return
	}

	data := a.newPageData(r, "Dashboard Admin")
	data["PendingRecruiters"] = pendingCount
	data["TotalRecruiters"] = totalRecruiters
	data["TotalCandidates"] = candidates
	data["TotalJobs"] = jobsCount
	a.Render.Render(w, "admin_dashboard.html", data)
}

func (a *App) AdminRecruitersPage(w http.ResponseWriter, r *http.Request) {
	const pendingPageSize = 10
	const allPageSize = 15
	pendingPage := parsePageParam(r, "pending_page")
	allPage := parsePageParam(r, "all_page")

	pendingTotal, err := a.Users.CountRecruitersByStatus(models.RecruiterPending)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat data")
		return
	}
	allTotal, err := a.Users.CountRecruiters()
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat data")
		return
	}

	pendingPager := buildPagination(pendingTotal, pendingPage, pendingPageSize)
	allPager := buildPagination(allTotal, allPage, allPageSize)

	pending, err := a.Users.ListRecruitersByStatusPaginated(models.RecruiterPending, pendingPageSize, paginationOffset(pendingPager.Page, pendingPageSize))
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat data")
		return
	}
	all, err := a.Users.ListAllRecruitersPaginated(allPageSize, paginationOffset(allPager.Page, allPageSize))
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat data")
		return
	}

	data := a.newPageData(r, "Approval Recruiter")
	data["Pending"] = pending
	data["All"] = all
	data["PendingPagination"] = pendingPager
	data["AllPagination"] = allPager
	a.Render.Render(w, "admin_recruiters.html", data)
}

func (a *App) AdminRecruiterApproveSubmit(w http.ResponseWriter, r *http.Request) {
	a.adminSetRecruiterStatus(w, r, "/admin/recruiters/", "/approve", models.RecruiterApproved)
}

func (a *App) AdminRecruiterRejectSubmit(w http.ResponseWriter, r *http.Request) {
	a.adminSetRecruiterStatus(w, r, "/admin/recruiters/", "/reject", models.RecruiterRejected)
}

func (a *App) adminSetRecruiterStatus(w http.ResponseWriter, r *http.Request, prefix, suffix string, status models.RecruiterStatus) {
	path := strings.TrimPrefix(r.URL.Path, prefix)
	path = strings.TrimSuffix(path, suffix)
	userID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		a.renderError(w, r, http.StatusNotFound, "Recruiter tidak ditemukan")
		return
	}

	admin := middleware.CurrentUser(r)
	if err := a.Users.UpdateRecruiterStatus(userID, admin.ID, status); err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memperbarui status recruiter")
		return
	}

	http.Redirect(w, r, "/admin/recruiters", http.StatusSeeOther)
}

func (a *App) AdminCandidatesPage(w http.ResponseWriter, r *http.Request) {
	const pageSize = 20
	page := parsePageParam(r, "page")
	total, err := a.Users.CountCandidates()
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat data")
		return
	}
	pagination := buildPagination(total, page, pageSize)

	candidates, err := a.Users.ListCandidatesPaginated(pageSize, paginationOffset(pagination.Page, pageSize))
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat data")
		return
	}
	data := a.newPageData(r, "Candidates")
	data["Candidates"] = candidates
	data["Pagination"] = pagination
	a.Render.Render(w, "admin_candidates.html", data)
}

func (a *App) AdminJobsPage(w http.ResponseWriter, r *http.Request) {
	const pageSize = 20
	page := parsePageParam(r, "page")
	total, err := a.Jobs.CountAll()
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat data")
		return
	}
	pagination := buildPagination(total, page, pageSize)

	jobs, err := a.Jobs.ListAllPaginated(pageSize, paginationOffset(pagination.Page, pageSize))
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat data")
		return
	}
	data := a.newPageData(r, "Semua Lowongan")
	data["Jobs"] = jobs
	data["Pagination"] = pagination
	a.Render.Render(w, "admin_jobs.html", data)
}
