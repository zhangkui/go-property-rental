package handler

import (
	"go-property-rental/internal/domain/entity"
	"net/http"
)

func (h Handler) CreateSettlement(w http.ResponseWriter, r *http.Request) {
	var v struct {
		LeaseID          string `json:"lease_id"`
		DepositDeduction int64  `json:"deposit_deduction"`
		Items            []struct {
			Kind        string `json:"kind"`
			Description string `json:"description"`
			Amount      int64  `json:"amount"`
		} `json:"items"`
		Readings []struct {
			Kind    string `json:"kind"`
			Reading int64  `json:"reading"`
		} `json:"readings"`
	}
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	items := make([]entity.SettlementItem, len(v.Items))
	for i, x := range v.Items {
		items[i] = entity.SettlementItem{Kind: x.Kind, Description: x.Description, Amount: moneyValue(x.Amount)}
	}
	readings := make([]entity.MeterReading, len(v.Readings))
	for i, x := range v.Readings {
		readings[i] = entity.MeterReading{Kind: x.Kind, Reading: moneyValue(x.Reading)}
	}
	x, e := h.Settlements.Create(r.Context(), v.LeaseID, v.DepositDeduction, items, readings, actor(r))
	if e != nil {
		fail(w, 409, e)
		return
	}
	h.recordAudit(r, "settlement.create", "settlement", x.ID, x)
	write(w, 201, x)
}
func (h Handler) GetSettlement(w http.ResponseWriter, r *http.Request) {
	x, items, readings, e := h.Settlements.Get(r.Context(), r.PathValue("id"))
	if e != nil {
		fail(w, 404, e)
		return
	}
	write(w, 200, map[string]any{"settlement": x, "items": items, "readings": readings})
}
func (h Handler) CompleteSettlement(w http.ResponseWriter, r *http.Request) {
	if e := h.Settlements.Complete(r.Context(), r.PathValue("id"), actor(r)); e != nil {
		fail(w, 409, e)
		return
	}
	h.recordAudit(r, "settlement.complete", "settlement", r.PathValue("id"), nil)
	write(w, 200, map[string]string{"status": "completed"})
}
