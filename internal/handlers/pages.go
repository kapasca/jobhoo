package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/jobhoo/jobhoo/internal/database"
	"github.com/jobhoo/jobhoo/internal/middleware"
	"github.com/jobhoo/jobhoo/internal/models"
)

const jobsPerPage = 30

// BasePageData carries the fields every page needs for shared chrome (nav
// login state, active nav highlighting). Every page-specific data struct
// embeds this so base.html can rely on it being present everywhere.
type BasePageData struct {
	ActiveNav     string
	CurrentUser   *models.User
	DashboardPath string
	CSRFToken     string
}

func newBasePageData(r *http.Request, activeNav string) BasePageData {
	user := middleware.CurrentUser(r)
	path := ""
	if user != nil {
		switch user.Role {
		case models.RoleRecruiter:
			path = "/dashboard/recruiter"
		case models.RoleAdmin:
			path = "/dashboard/admin"
		default:
			path = "/dashboard/candidate"
		}
	}
	return BasePageData{ActiveNav: activeNav, CurrentUser: user, DashboardPath: path, CSRFToken: middleware.CSRFToken(r)}
}

type pageData struct {
	BasePageData
	Jobs any
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	result, err := h.Jobs.ListPublished(r.Context(), database.JobListFilter{Limit: 9})
	if err != nil {
		log.Printf("Home handler error: %v", err)
		http.Error(w, "could not load jobs", http.StatusInternalServerError)
		return
	}
	log.Printf("Home handler: loaded %d jobs (total: %d)", len(result.Jobs), result.Total)
	h.Render.Render(w, http.StatusOK, "home.html", pageData{BasePageData: newBasePageData(r, "home"), Jobs: result.Jobs})
}

// jobsPageData carries everything the jobs listing page needs: the current
// page of results, the active search/category filters (so the form and chips
// reflect what was actually applied), and pagination bounds.
type jobsPageData struct {
	BasePageData
	Jobs       []models.Job
	Categories []jobCategoryOption
	Search     string
	Category   string
	Page       int
	TotalPages int
	Total      int
	HasPrev    bool
	HasNext    bool
	PrevPage   int
	NextPage   int
}

type jobCategoryOption struct {
	Value    string
	Label    string
	IsActive bool
}

func (h *Handlers) JobsIndex(w http.ResponseWriter, r *http.Request) {
	h.renderJobsPage(w, r, "jobs.html")
}

// JobsSearch is an HTMX partial endpoint: it returns only the results
// fragment (job grid + pagination) so it can swap into #job-results without
// a full page reload.
func (h *Handlers) JobsSearch(w http.ResponseWriter, r *http.Request) {
	h.renderJobsPage(w, r, "jobs.html", "job-results")
}

func (h *Handlers) renderJobsPage(w http.ResponseWriter, r *http.Request, page string, blockName ...string) {
	search := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")
	pageNum, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if pageNum < 1 {
		pageNum = 1
	}

	result, err := h.Jobs.ListPublished(r.Context(), database.JobListFilter{
		Search:   search,
		Category: models.JobCategory(category),
		Limit:    jobsPerPage,
		Offset:   (pageNum - 1) * jobsPerPage,
	})
	if err != nil {
		http.Error(w, "could not load jobs", http.StatusInternalServerError)
		return
	}

	totalPages := (result.Total + jobsPerPage - 1) / jobsPerPage
	if totalPages < 1 {
		totalPages = 1
	}

	categoryOptions := make([]jobCategoryOption, 0, len(models.JobCategories)+1)
	categoryOptions = append(categoryOptions, jobCategoryOption{Value: "", Label: "All categories", IsActive: category == ""})
	for _, c := range models.JobCategories {
		categoryOptions = append(categoryOptions, jobCategoryOption{
			Value: string(c.Value), Label: c.Label, IsActive: string(c.Value) == category,
		})
	}

	data := jobsPageData{
		BasePageData: newBasePageData(r, "jobs"),
		Jobs:         result.Jobs,
		Categories:   categoryOptions,
		Search:       search,
		Category:     category,
		Page:         pageNum,
		TotalPages:   totalPages,
		Total:        result.Total,
		HasPrev:      pageNum > 1,
		HasNext:      pageNum < totalPages,
		PrevPage:     pageNum - 1,
		NextPage:     pageNum + 1,
	}

	if len(blockName) == 1 {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl := h.Render.cache[page]
		if tmpl == nil {
			http.Error(w, "template not found", http.StatusInternalServerError)
			return
		}
		if err := tmpl.ExecuteTemplate(w, blockName[0], data); err != nil {
			http.Error(w, "render error", http.StatusInternalServerError)
		}
		return
	}

	h.Render.Render(w, http.StatusOK, page, data)
}
