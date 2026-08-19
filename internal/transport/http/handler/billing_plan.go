package handler

import (
	"net/http"
	"time"

	"go-property-rental/internal/domain/entity"
)

type billingPlanInput struct {
	LeaseID         string                   `json:"lease_id"`
	BillingDay      int                      `json:"billing_day"`
	DueDays         int                      `json:"due_days"`
	AdvanceDays     int                      `json:"advance_days"`
	NextPeriodStart string                   `json:"next_period_start"`
	Status          string                   `json:"status"`
	Items           []entity.BillingPlanItem `json:"items"`
}

func (h Handler) ListBillingPlans(w http.ResponseWriter, r *http.Request) {
	page, size := pagination(r)
	items, err := h.BillingPlans.ListPlans(r.Context(), page, size, r.URL.Query().Get("status"))
	if err != nil {
		fail(w, 500, err)
		return
	}
	write(w, 200, map[string]any{"items": items, "page": page, "page_size": size})
}
func (h Handler) GetBillingPlan(w http.ResponseWriter, r *http.Request) {
	item, err := h.BillingPlans.GetPlan(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, 404, err)
		return
	}
	write(w, 200, item)
}
func (h Handler) SaveBillingPlan(w http.ResponseWriter, r *http.Request) {
	var input billingPlanInput
	if err := decode(r, &input); err != nil {
		fail(w, 400, err)
		return
	}
	plan := entity.BillingPlan{ID: r.PathValue("id"), LeaseID: input.LeaseID, BillingDay: input.BillingDay, DueDays: input.DueDays, AdvanceDays: input.AdvanceDays, Status: input.Status}
	if input.NextPeriodStart != "" {
		value, err := parseDate(input.NextPeriodStart)
		if err != nil {
			fail(w, 400, err)
			return
		}
		plan.NextPeriodStart = value
	}
	saved, err := h.BillingPlans.SavePlan(r.Context(), plan, input.Items)
	if err != nil {
		fail(w, 400, err)
		return
	}
	h.recordAudit(r, "billing_plan.save", "billing_plan", saved.ID, saved)
	write(w, 201, saved)
}
func (h Handler) SetBillingPlanStatus(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Status string `json:"status"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, 400, err)
		return
	}
	if err := h.BillingPlans.SetStatus(r.Context(), r.PathValue("id"), input.Status); err != nil {
		fail(w, 400, err)
		return
	}
	h.recordAudit(r, "billing_plan.status", "billing_plan", r.PathValue("id"), input)
	write(w, 200, input)
}
func (h Handler) PreviewBillingRun(w http.ResponseWriter, r *http.Request) {
	var input struct {
		At string `json:"at"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, 400, err)
		return
	}
	at := time.Now().UTC()
	if input.At != "" {
		value, err := parseDate(input.At)
		if err != nil {
			fail(w, 400, err)
			return
		}
		at = value
	}
	preview, err := h.BillingPlans.Preview(r.Context(), at)
	if err != nil {
		fail(w, 500, err)
		return
	}
	write(w, 200, preview)
}
func (h Handler) ExecuteBillingRun(w http.ResponseWriter, r *http.Request) {
	run, err := h.BillingPlans.RunDue(r.Context(), "manual", actor(r), time.Now().UTC())
	if err != nil {
		fail(w, 409, err)
		return
	}
	h.recordAudit(r, "billing_run.execute", "billing_run", run.ID, run)
	write(w, 201, run)
}
func (h Handler) ListBillingRuns(w http.ResponseWriter, r *http.Request) {
	page, size := pagination(r)
	items, err := h.BillingPlans.ListRuns(r.Context(), page, size)
	if err != nil {
		fail(w, 500, err)
		return
	}
	write(w, 200, map[string]any{"items": items, "page": page, "page_size": size})
}
func (h Handler) GetBillingRun(w http.ResponseWriter, r *http.Request) {
	run, items, err := h.BillingPlans.GetRun(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, 404, err)
		return
	}
	write(w, 200, map[string]any{"run": run, "items": items})
}
func (h Handler) RetryBillingRunItem(w http.ResponseWriter, r *http.Request) {
	if err := h.BillingPlans.Retry(r.Context(), r.PathValue("itemID")); err != nil {
		fail(w, 409, err)
		return
	}
	h.recordAudit(r, "billing_run.retry", "billing_run_item", r.PathValue("itemID"), nil)
	write(w, 200, map[string]string{"status": "succeeded"})
}
