package handler

import (
	"go-property-rental/internal/domain/entity"
	"net/http"
)

func (h Handler) ListWorkOrders(w http.ResponseWriter, r *http.Request) {
	p, s := pagination(r)
	xs, e := h.WorkOrders.List(r.Context(), p, s, r.URL.Query().Get("property_id"), r.URL.Query().Get("status"))
	if e != nil {
		fail(w, 500, e)
		return
	}
	write(w, 200, map[string]any{"items": xs})
}
func (h Handler) GetWorkOrder(w http.ResponseWriter, r *http.Request) {
	x, ms, e := h.WorkOrders.Get(r.Context(), r.PathValue("id"))
	if e != nil {
		fail(w, 404, e)
		return
	}
	write(w, 200, map[string]any{"work_order": x, "materials": ms})
}
func (h Handler) CreateWorkOrder(w http.ResponseWriter, r *http.Request) {
	var x entity.WorkOrder
	if e := decode(r, &x); e != nil {
		fail(w, 400, e)
		return
	}
	x, e := h.WorkOrders.Create(r.Context(), x)
	if e != nil {
		fail(w, 400, e)
		return
	}
	h.recordAudit(r, "work_order.create", "work_order", x.ID, x)
	write(w, 201, x)
}
func (h Handler) AssignWorkOrder(w http.ResponseWriter, r *http.Request) {
	var v struct {
		AssigneeID string `json:"assignee_id"`
	}
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	if e := h.WorkOrders.Assign(r.Context(), r.PathValue("id"), v.AssigneeID, actor(r)); e != nil {
		fail(w, 409, e)
		return
	}
	h.recordAudit(r, "work_order.action", "work_order", r.PathValue("id"), v)
	write(w, 200, v)
}
func (h Handler) AddWorkOrderMaterial(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Name     string `json:"name"`
		Quantity int64  `json:"quantity"`
		UnitCost int64  `json:"unit_cost"`
	}
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	if e := h.WorkOrders.AddMaterial(r.Context(), r.PathValue("id"), v.Name, v.Quantity, v.UnitCost); e != nil {
		fail(w, 409, e)
		return
	}
	h.recordAudit(r, "work_order.material", "work_order", r.PathValue("id"), v)
	write(w, 201, v)
}
func (h Handler) TransitionWorkOrder(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	if e := h.WorkOrders.Transition(r.Context(), r.PathValue("id"), v.Status, v.Reason, actor(r)); e != nil {
		fail(w, 409, e)
		return
	}
	write(w, 200, v)
}
func (h Handler) ConfirmWorkOrder(w http.ResponseWriter, r *http.Request) {
	if e := h.WorkOrders.Confirm(r.Context(), r.PathValue("id")); e != nil {
		fail(w, 409, e)
		return
	}
	h.recordAudit(r, "work_order.confirm", "work_order", r.PathValue("id"), nil)
	write(w, 200, map[string]string{"status": "closed"})
}
