package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) FreezeUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Users.FreezeUser(r.Context(), id); err != nil {
		http.Error(w, "could not freeze user", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/dashboard/admin?tab=users", http.StatusSeeOther)
}

func (h *Handlers) UnfreezeUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Users.UnfreezeUser(r.Context(), id); err != nil {
		http.Error(w, "could not unfreeze user", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/dashboard/admin?tab=users", http.StatusSeeOther)
}

func (h *Handlers) FreezeJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Jobs.FreezeJob(r.Context(), id); err != nil {
		http.Error(w, "could not freeze job", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/dashboard/admin?tab=jobs", http.StatusSeeOther)
}

func (h *Handlers) UnfreezeJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Jobs.UnfreezeJob(r.Context(), id); err != nil {
		http.Error(w, "could not unfreeze job", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/dashboard/admin?tab=jobs", http.StatusSeeOther)
}
