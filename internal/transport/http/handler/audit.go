package handler

import "net/http"

func (h Handler) ListAudits(w http.ResponseWriter, r *http.Request) {
	p, s := pagination(r)
	xs, e := h.Audits.List(r.Context(), p, s, r.URL.Query().Get("resource"), r.URL.Query().Get("actor_id"))
	if e != nil {
		fail(w, 500, e)
		return
	}
	write(w, 200, map[string]any{"items": xs})
}
