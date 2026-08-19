package handler

import (
	"go-property-rental/internal/domain/entity"
	"net/http"
)

type leaseInput struct {
	PropertyID  string                 `json:"property_id"`
	TenantID    string                 `json:"tenant_id"`
	StartDate   string                 `json:"start_date"`
	EndDate     string                 `json:"end_date"`
	MonthlyRent int64                  `json:"monthly_rent"`
	Deposit     int64                  `json:"deposit"`
	Occupants   []entity.LeaseOccupant `json:"occupants"`
}

func (h Handler) ListLeases(w http.ResponseWriter, r *http.Request) {
	p, s := pagination(r)
	xs, e := h.Leases.List(r.Context(), p, s, r.URL.Query().Get("property_id"), r.URL.Query().Get("status"))
	if e != nil {
		fail(w, 500, e)
		return
	}
	write(w, 200, map[string]any{"items": xs})
}
func (h Handler) GetLease(w http.ResponseWriter, r *http.Request) {
	x, e := h.Leases.Get(r.Context(), r.PathValue("id"))
	if e != nil {
		fail(w, 404, e)
		return
	}
	versions, _ := h.Leases.Versions(r.Context(), x.ID)
	write(w, 200, map[string]any{"lease": x, "versions": versions})
}
func (h Handler) CreateLease(w http.ResponseWriter, r *http.Request) {
	var v leaseInput
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	start, e := parseDate(v.StartDate)
	if e != nil {
		fail(w, 400, e)
		return
	}
	end, e := parseDate(v.EndDate)
	if e != nil {
		fail(w, 400, e)
		return
	}
	x := entity.Lease{PropertyID: v.PropertyID, TenantID: v.TenantID, StartDate: start, EndDate: end, MonthlyRent: moneyValue(v.MonthlyRent), Deposit: moneyValue(v.Deposit)}
	x, e = h.Leases.Create(r.Context(), x, v.Occupants, actor(r))
	if e != nil {
		fail(w, 409, e)
		return
	}
	h.recordAudit(r, "lease.create", "lease", x.ID, x)
	write(w, 201, x)
}
func (h Handler) RenewLease(w http.ResponseWriter, r *http.Request) {
	var v struct {
		EndDate     string `json:"end_date"`
		MonthlyRent int64  `json:"monthly_rent"`
		Deposit     int64  `json:"deposit"`
	}
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	end, e := parseDate(v.EndDate)
	if e != nil {
		fail(w, 400, e)
		return
	}
	if e = h.Leases.Renew(r.Context(), r.PathValue("id"), end, v.MonthlyRent, v.Deposit, actor(r)); e != nil {
		fail(w, 409, e)
		return
	}
	h.recordAudit(r, "lease.renew", "lease", r.PathValue("id"), v)
	write(w, 200, map[string]string{"status": "renewed"})
}
func (h Handler) TransitionLease(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	permission := "lease:submit"
	if v.Status == "approved" || v.Status == "active" {
		permission = "lease:approve"
	}
	if v.Status == "terminated" || v.Status == "closed" {
		permission = "lease:terminate"
	}
	if e := requirePermission(r, permission); e != nil {
		fail(w, http.StatusForbidden, e)
		return
	}
	if e := h.Leases.Transition(r.Context(), r.PathValue("id"), v.Status, v.Reason, actor(r)); e != nil {
		fail(w, 409, e)
		return
	}
	h.recordAudit(r, "lease.transition", "lease", r.PathValue("id"), v)
	write(w, 200, v)
}
