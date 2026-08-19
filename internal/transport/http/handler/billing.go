package handler

import (
	"go-property-rental/internal/domain/entity"
	"net/http"
	"time"
)

func (h Handler) ListBills(w http.ResponseWriter, r *http.Request) {
	p, s := pagination(r)
	xs, e := h.Billing.List(r.Context(), p, s, r.URL.Query().Get("lease_id"), r.URL.Query().Get("status"))
	if e != nil {
		fail(w, 500, e)
		return
	}
	write(w, 200, map[string]any{"items": xs})
}
func (h Handler) GetBill(w http.ResponseWriter, r *http.Request) {
	x, e := h.Billing.Get(r.Context(), r.PathValue("id"))
	if e != nil {
		fail(w, 404, e)
		return
	}
	items, e := h.Billing.Items(r.Context(), x.ID)
	if e != nil {
		fail(w, 500, e)
		return
	}
	write(w, 200, map[string]any{"bill": x, "items": items})
}
func (h Handler) GenerateBill(w http.ResponseWriter, r *http.Request) {
	var v struct {
		LeaseID        string `json:"lease_id"`
		PeriodStart    string `json:"period_start"`
		IdempotencyKey string `json:"idempotency_key"`
	}
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	start, e := parseDate(v.PeriodStart)
	if e != nil {
		fail(w, 400, e)
		return
	}
	x, created, e := h.Billing.GenerateMonthly(r.Context(), v.LeaseID, start, v.IdempotencyKey)
	if e != nil {
		fail(w, 409, e)
		return
	}
	status := 200
	if created {
		status = 201
	}
	h.recordAudit(r, "bill.generate", "bill", x.ID, v)
	write(w, status, map[string]any{"bill": x, "created": created})
}
func (h Handler) AdjustBill(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Kind   string `json:"kind"`
		Amount int64  `json:"amount"`
		Reason string `json:"reason"`
	}
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	var e error
	if v.Kind == "penalty" {
		e = h.Billing.AddPenalty(r.Context(), r.PathValue("id"), v.Amount, v.Reason, actor(r))
	} else {
		e = h.Billing.Discount(r.Context(), r.PathValue("id"), v.Amount, v.Reason, actor(r))
	}
	if e != nil {
		fail(w, 409, e)
		return
	}
	h.recordAudit(r, "bill.adjust", "bill", r.PathValue("id"), v)
	write(w, 200, v)
}
func (h Handler) ApplyLateFee(w http.ResponseWriter, r *http.Request) {
	var v struct {
		At               string `json:"at"`
		DailyBasisPoints int    `json:"daily_basis_points"`
		MaxDays          int    `json:"max_days"`
	}
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	at := time.Now().UTC()
	if v.At != "" {
		var e error
		at, e = parseDate(v.At)
		if e != nil {
			fail(w, 400, e)
			return
		}
	}
	if e := h.Billing.ApplyLateFee(r.Context(), r.PathValue("id"), at, v.DailyBasisPoints, v.MaxDays, actor(r)); e != nil {
		fail(w, 409, e)
		return
	}
	write(w, 200, map[string]string{"status": "applied"})
}
func (h Handler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Reference   string `json:"reference"`
		Payer       string `json:"payer"`
		Amount      int64  `json:"amount"`
		Allocations []struct {
			BillID string `json:"bill_id"`
			Amount int64  `json:"amount"`
		} `json:"allocations"`
	}
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	as := make([]entity.PaymentAllocation, len(v.Allocations))
	for i, a := range v.Allocations {
		as[i] = entity.PaymentAllocation{BillID: a.BillID, Amount: moneyValue(a.Amount)}
	}
	x, e := h.Billing.Pay(r.Context(), v.Reference, v.Payer, v.Amount, as)
	if e != nil {
		fail(w, 409, e)
		return
	}
	h.recordAudit(r, "payment.post", "payment", x.ID, v)
	write(w, 201, x)
}
