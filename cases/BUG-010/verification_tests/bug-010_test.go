package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
)

type bug010BillingStore struct {
	payments    []entity.Payment
	allocations [][]entity.PaymentAllocation
}

func (*bug010BillingStore) Generate(context.Context, entity.Bill, string) (entity.Bill, bool, error) {
	return entity.Bill{}, false, nil
}
func (*bug010BillingStore) List(context.Context, int, int, string, string) ([]entity.Bill, error) {
	return nil, nil
}
func (*bug010BillingStore) Get(context.Context, string) (entity.Bill, error) {
	return entity.Bill{}, nil
}
func (*bug010BillingStore) Items(context.Context, string) ([]entity.BillItem, error) { return nil, nil }
func (*bug010BillingStore) Adjust(context.Context, entity.BillAdjustment) error      { return nil }
func (s *bug010BillingStore) RecordPayment(_ context.Context, p entity.Payment, a []entity.PaymentAllocation) error {
	s.payments = append(s.payments, p)
	s.allocations = append(s.allocations, append([]entity.PaymentAllocation(nil), a...))
	return nil
}
func TestBug010_BusinessRegression(t *testing.T) {
	s := &bug010BillingStore{}
	svc := service.BillingService{Repo: s}
	alloc := []entity.PaymentAllocation{{BillID: "bill-1", Amount: 5000}}
	first, err := svc.Pay(context.Background(), "external-ref", "payer", 5000, append([]entity.PaymentAllocation(nil), alloc...))
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.Pay(context.Background(), "external-ref", "payer", 5000, append([]entity.PaymentAllocation(nil), alloc...))
	if err != nil {
		t.Fatal(err)
	}
	if first.Reference != "external-ref" || second.Reference != "external-ref" {
		t.Fatalf("client reference was rewritten: %q %q", first.Reference, second.Reference)
	}
	if len(s.payments) != 2 || s.payments[0].Reference != s.payments[1].Reference {
		t.Fatalf("duplicate calls did not preserve the same business reference: %#v", s.payments)
	}
	if s.allocations[0][0].PaymentID == "" || s.allocations[0][0].PaymentID != first.ID {
		t.Fatalf("allocation not linked to payment: %#v", s.allocations[0])
	}
}
