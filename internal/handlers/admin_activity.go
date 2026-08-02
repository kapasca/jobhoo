package handlers

import (
	"net/http"

	"github.com/jobhoo/jobhoo/internal/database"
)

const activityPageSize = 30

// AdminActivity shows a simple, newest-first feed of every AI call: the
// prompt sent, the response received, and who triggered it.
func (h *Handlers) AdminActivity(w http.ResponseWriter, r *http.Request) {
	pg := parsePage(r.URL.Query().Get("p"))
	logs, total, err := h.AICallLogs.List(r.Context(), activityPageSize, (pg-1)*activityPageSize)
	if err != nil {
		http.Error(w, "could not load activity log", http.StatusInternalServerError)
		return
	}

	h.Render.Render(w, http.StatusOK, "admin-activity.html", struct {
		BasePageData
		Logs       []database.AICallLogRow
		Page       int
		TotalPages int
	}{
		BasePageData: newBasePageData(r, "activity"),
		Logs:         logs,
		Page:         pg,
		TotalPages:   totalPages(total, activityPageSize),
	})
}
