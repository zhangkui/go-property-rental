package handler

import "net/http"

func (h Handler) ListFacilities(w http.ResponseWriter, r *http.Request) {
	xs, e := h.Facilities.List(r.Context())
	if e != nil {
		fail(w, 500, e)
		return
	}
	write(w, 200, map[string]any{"items": xs})
}
func (h Handler) CreateFacility(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Name string `json:"name"`
	}
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	x, e := h.Facilities.Create(r.Context(), v.Name)
	if e != nil {
		fail(w, 400, e)
		return
	}
	h.recordAudit(r, "facility.create", "facility", x.ID, x)
	write(w, 201, x)
}
func (h Handler) PropertyFacilities(w http.ResponseWriter, r *http.Request) {
	xs, e := h.Facilities.ForProperty(r.Context(), r.PathValue("id"))
	if e != nil {
		fail(w, 500, e)
		return
	}
	write(w, 200, map[string]any{"items": xs})
}
func (h Handler) ReplacePropertyFacilities(w http.ResponseWriter, r *http.Request) {
	var v struct {
		FacilityIDs []string `json:"facility_ids"`
	}
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	if e := h.Facilities.Replace(r.Context(), r.PathValue("id"), v.FacilityIDs); e != nil {
		fail(w, 400, e)
		return
	}
	h.recordAudit(r, "property.facilities.replace", "property", r.PathValue("id"), v)
	write(w, 200, v)
}
