package handler

import (
	"go-property-rental/internal/domain/entity"
	"net/http"
)

func (h Handler) ListTenants(w http.ResponseWriter, r *http.Request) {
	p, s := pagination(r)
	xs, e := h.Tenants.List(r.Context(), p, s, r.URL.Query().Get("status"))
	if e != nil {
		fail(w, 500, e)
		return
	}
	write(w, 200, map[string]any{"items": xs})
}
func (h Handler) GetTenant(w http.ResponseWriter, r *http.Request) {
	x, e := h.Tenants.Get(r.Context(), r.PathValue("id"))
	if e != nil {
		fail(w, 404, e)
		return
	}
	h.recordAudit(r, "tenant.update", "tenant", x.ID, x)
	write(w, 200, x)
}
func (h Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var x entity.Tenant
	if e := decode(r, &x); e != nil {
		fail(w, 400, e)
		return
	}
	x, e := h.Tenants.Create(r.Context(), x)
	if e != nil {
		fail(w, 400, e)
		return
	}
	h.recordAudit(r, "tenant.create", "tenant", x.ID, x)
	write(w, 201, x)
}
func (h Handler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	var x entity.Tenant
	if e := decode(r, &x); e != nil {
		fail(w, 400, e)
		return
	}
	x.ID = r.PathValue("id")
	if e := h.Tenants.Update(r.Context(), x); e != nil {
		fail(w, 400, e)
		return
	}
	write(w, 200, x)
}
func (h Handler) SetTenantStatus(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Status string `json:"status"`
	}
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	if e := h.Tenants.SetStatus(r.Context(), r.PathValue("id"), v.Status); e != nil {
		fail(w, 400, e)
		return
	}
	write(w, 200, v)
}
