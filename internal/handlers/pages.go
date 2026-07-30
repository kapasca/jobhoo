package handlers

import (
	"log"
	"net/http"
	"net/url"
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
	Jobs         []models.Job
	Search       string
	Category     string
	Location     string
	Sort         string
	Arrangements map[string]bool // e.g. Arrangements["remote"] == true if that checkbox is active
	Employments  map[string]bool
	Page         int
	TotalPages   int
	Total        int
	From         int
	To           int
	HasPrev      bool
	HasNext      bool
	PrevPage     int
	NextPage     int
	// FilterQuery is the current filters pre-encoded as a URL query string
	// fragment (no leading "?"/"&"), so Prev/Next pagination links can carry
	// every active filter forward instead of only q/category.
	FilterQuery string
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
	q := r.URL.Query()
	search := q.Get("q")
	category := q.Get("category")
	location := q.Get("location")
	sort := q.Get("sort")
	arrangements := q["arrangement"] // multi-value checkboxes, e.g. ?arrangement=remote&arrangement=hybrid
	employments := q["employment"]
	pageNum, _ := strconv.Atoi(q.Get("page"))
	if pageNum < 1 {
		pageNum = 1
	}

	result, err := h.Jobs.ListPublished(r.Context(), database.JobListFilter{
		Search:           search,
		Category:         models.JobCategory(category),
		Location:         location,
		Sort:             sort,
		WorkArrangements: arrangements,
		EmploymentTypes:  employments,
		Limit:            jobsPerPage,
		Offset:           (pageNum - 1) * jobsPerPage,
	})
	if err != nil {
		http.Error(w, "could not load jobs", http.StatusInternalServerError)
		return
	}

	totalPages := (result.Total + jobsPerPage - 1) / jobsPerPage
	if totalPages < 1 {
		totalPages = 1
	}

	arrangementSet := toSet(arrangements)
	employmentSet := toSet(employments)

	filterQuery := url.Values{}
	if search != "" {
		filterQuery.Set("q", search)
	}
	if category != "" {
		filterQuery.Set("category", category)
	}
	if location != "" {
		filterQuery.Set("location", location)
	}
	if sort != "" {
		filterQuery.Set("sort", sort)
	}
	for _, a := range arrangements {
		filterQuery.Add("arrangement", a)
	}
	for _, e := range employments {
		filterQuery.Add("employment", e)
	}

	from := (pageNum-1)*jobsPerPage + 1
	to := from + len(result.Jobs) - 1
	if result.Total == 0 {
		from = 0
		to = 0
	}

	data := jobsPageData{
		BasePageData: newBasePageData(r, "jobs"),
		Jobs:         result.Jobs,
		Search:       search,
		Category:     category,
		Location:     location,
		Sort:         sort,
		Arrangements: arrangementSet,
		Employments:  employmentSet,
		Page:         pageNum,
		TotalPages:   totalPages,
		Total:        result.Total,
		From:         from,
		To:           to,
		HasPrev:      pageNum > 1,
		HasNext:      pageNum < totalPages,
		PrevPage:     pageNum - 1,
		NextPage:     pageNum + 1,
		FilterQuery:  filterQuery.Encode(),
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

func toSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, v := range values {
		set[v] = true
	}
	return set
}
