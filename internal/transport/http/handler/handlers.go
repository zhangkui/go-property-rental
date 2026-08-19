package handler

import (
	"net/http"
	"time"

	"go-property-rental/internal/service"
)

type Handler struct {
	Auth          service.AuthService
	RBAC          service.RBACService
	Properties    service.PropertyService
	Tenants       service.TenantService
	Facilities    service.FacilityService
	Leases        service.LeaseService
	Billing       service.BillingService
	BillingPlans  service.BillingPlanService
	Deposits      service.DepositService
	WorkOrders    service.WorkOrderService
	Settlements   service.SettlementService
	Audits        service.AuditService
	Approvals     service.ApprovalService
	Notifications service.NotificationService
	Reports       service.ReportService
	Dashboard     service.DashboardService
}

func (h Handler) ListProperties(w http.ResponseWriter, r *http.Request) {
	page, size := pagination(r)
	items, err := h.Properties.List(r.Context(), page, size, r.URL.Query().Get("status"))
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	write(w, http.StatusOK, map[string]any{"items": items, "page": page, "page_size": size})
}

func (h Handler) CreateProperty(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Building      string `json:"building"`
		Room          string `json:"room"`
		Status        string `json:"status"`
		AvailableFrom string `json:"available_from"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	availableFrom := time.Time{}
	var err error
	if input.AvailableFrom != "" {
		availableFrom, err = parseDate(input.AvailableFrom)
		if err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
	}
	property, err := h.Properties.Create(r.Context(), input.Building, input.Room, input.Status, availableFrom)
	if err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	h.recordAudit(r, "property.create", "property", property.ID, property)
	write(w, http.StatusCreated, property)
}

func (h Handler) ChangePropertyStatus(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Status string `json:"status"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if err := h.Properties.ChangeStatus(r.Context(), r.PathValue("id"), input.Status); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	h.recordAudit(r, "property.status_changed", "property", r.PathValue("id"), input)
	write(w, http.StatusOK, input)
}
