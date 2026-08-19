package handler

import "net/http"

func (h Handler) DepositTransactions(w http.ResponseWriter, r *http.Request) {
	xs, e := h.Deposits.List(r.Context(), r.PathValue("leaseID"))
	if e != nil {
		fail(w, 500, e)
		return
	}
	balance, e := h.Deposits.Balance(r.Context(), r.PathValue("leaseID"))
	if e != nil {
		fail(w, 500, e)
		return
	}
	write(w, 200, map[string]any{"items": xs, "balance": balance})
}
func (h Handler) CreateDepositTransaction(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Kind      string `json:"kind"`
		Reference string `json:"reference"`
		Reason    string `json:"reason"`
		Amount    int64  `json:"amount"`
	}
	if e := decode(r, &v); e != nil {
		fail(w, 400, e)
		return
	}
	permission := "deposit:collect"
	if v.Kind == "deduct" {
		permission = "deposit:deduct"
	}
	if v.Kind == "refund" {
		permission = "deposit:refund"
	}
	if e := requirePermission(r, permission); e != nil {
		fail(w, http.StatusForbidden, e)
		return
	}
	x, e := h.Deposits.Transaction(r.Context(), r.PathValue("leaseID"), v.Kind, v.Reference, v.Reason, actor(r), v.Amount)
	if e != nil {
		fail(w, 409, e)
		return
	}
	h.recordAudit(r, "deposit."+v.Kind, "lease", r.PathValue("leaseID"), x)
	write(w, 201, x)
}
