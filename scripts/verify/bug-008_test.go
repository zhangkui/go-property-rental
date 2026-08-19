package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
	"time"
)

type bug008BillingStore struct {
	bill        entity.Bill
	adjustments []entity.BillAdjustment
}

func (*bug008BillingStore) Generate(context.Context, entity.Bill, string) (entity.Bill, bool, error) {
	return entity.Bill{}, false, nil
}
func (*bug008BillingStore) List(context.Context, int, int, string, string) ([]entity.Bill, error) {
	return nil, nil
}
func (s *bug008BillingStore) Get(context.Context, string) (entity.Bill, error)       { return s.bill, nil }
func (*bug008BillingStore) Items(context.Context, string) ([]entity.BillItem, error) { return nil, nil }
func (s *bug008BillingStore) Adjust(_ context.Context, a entity.BillAdjustment) error {
	s.adjustments = append(s.adjustments, a)
	return nil
}
func (*bug008BillingStore) RecordPayment(context.Context, entity.Payment, []entity.PaymentAllocation) error {
	return nil
}
func TestBug008_BusinessRegression(t *testing.T) {
	due := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	paid := &bug008BillingStore{bill: entity.Bill{ID: "bill-1", Status: "paid", DueDate: due, Amount: 10000, Paid: 10000}}
	if err := (service.BillingService{Repo: paid}).ApplyLateFee(context.Background(), "bill-1", due.AddDate(0, 0, 10), 100, 30, "admin"); err != nil {
		t.Fatal(err)
	}
	if len(paid.adjustments) != 0 {
		t.Fatalf("paid bill received late-fee adjustments: %#v", paid.adjustments)
	}
	unpaid := &bug008BillingStore{bill: entity.Bill{ID: "bill-2", Status: "unpaid", DueDate: due, Amount: 10000}}
	if err := (service.BillingService{Repo: unpaid}).ApplyLateFee(context.Background(), "bill-2", due.AddDate(0, 0, 2), 100, 30, "admin"); err != nil {
		t.Fatal(err)
	}
	if len(unpaid.adjustments) != 1 || int64(unpaid.adjustments[0].Amount) != 200 {
		t.Fatalf("overdue unpaid bill did not receive expected fee: %#v", unpaid.adjustments)
	}
	boundary := &bug008BillingStore{bill: entity.Bill{ID: "bill-3", Status: "unpaid", DueDate: due, Amount: 10000}}
	_ = (service.BillingService{Repo: boundary}).ApplyLateFee(context.Background(), "bill-3", due, 100, 30, "admin")
	if len(boundary.adjustments) != 0 {
		t.Fatal("bill received fee exactly at due date")
	}
}
