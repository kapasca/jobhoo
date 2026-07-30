package handlers

import (
	"net/http"

	"github.com/jobhoo/jobhoo/internal/database"
)

type companiesPageData struct {
	BasePageData
	Companies []database.CompanyWithJobCount
}

// CompaniesDirectory is the public "Explore Companies" page — no auth
// required, since discovering who's hiring is part of job discovery just
// like browsing jobs.
func (h *Handlers) CompaniesDirectory(w http.ResponseWriter, r *http.Request) {
	companies, err := h.Companies.ListAll(r.Context())
	if err != nil {
		http.Error(w, "could not load companies", http.StatusInternalServerError)
		return
	}
	h.Render.Render(w, http.StatusOK, "companies.html", companiesPageData{
		BasePageData: newBasePageData(r, "companies"),
		Companies:    companies,
	})
}
