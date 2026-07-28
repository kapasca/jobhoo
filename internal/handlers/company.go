package handlers

import (
	"errors"
	"net/http"
	"strings"

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
}

func (h *Handlers) CompanySetupPage(w http.ResponseWriter, r *http.Request) {
	h.Render.Render(w, http.StatusOK, "company-setup.html", companySetupData{BasePageData: newBasePageData(r, "post")})
}

func (h *Handlers) CompanySetup(w http.ResponseWriter, r *http.Request) {
	user := middleware.CurrentUser(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
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

	if _, err := h.Companies.Create(r.Context(), user.ID, data.Name, data.Website, data.Description, data.Industry); err != nil {
		data.Error = "Something went wrong creating your company. Please try again."
		h.Render.Render(w, http.StatusBadRequest, "company-setup.html", data)
		return
	}

	http.Redirect(w, r, "/post-job", http.StatusSeeOther)
}

// requireCompany fetches the current recruiter's company. If they don't have
// one yet, it redirects to company setup and returns ok=false — callers
// must return immediately in that case, since the response is already
// written.
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
