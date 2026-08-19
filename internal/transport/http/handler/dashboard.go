package handler

import "net/http"

func (h Handler) DashboardSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.Dashboard.Summary(r.Context(), actor(r))
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	write(w, http.StatusOK, summary)
}
