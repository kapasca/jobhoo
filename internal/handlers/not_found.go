package handlers

import "net/http"

func (h *Handlers) NotFoundPage(w http.ResponseWriter, r *http.Request) {
	h.Render.Render(w, http.StatusNotFound, "404.html", newBasePageData(r, ""))
}
