package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
	"time"
)

type bug007BillingStore struct{ keys []string }

func (s *bug007BillingStore) Generate(_ context.Context, b entity.Bill, key string) (entity.Bill, bool, error) {
	s.keys = append(s.keys, key)
	return b, true, nil
}
func (*bug007BillingStore) List(context.Context, int, int, string, string) ([]entity.Bill, error) {
	return nil, nil
}
func (*bug007BillingStore) Get(context.Context, string) (entity.Bill, error) {
	return entity.Bill{}, nil
}
func (*bug007BillingStore) Items(context.Context, string) ([]entity.BillItem, error) { return nil, nil }
func (*bug007BillingStore) Adjust(context.Context, entity.BillAdjustment) error      { return nil }
func (*bug007BillingStore) RecordPayment(context.Context, entity.Payment, []entity.PaymentAllocation) error {
	return nil
}

type bug007LeaseStore struct{ lease entity.Lease }

func (*bug007LeaseStore) List(context.Context, int, int, string, string) ([]entity.Lease, error) {
	return nil, nil
}
func (s *bug007LeaseStore) Get(context.Context, string) (entity.Lease, error) { return s.lease, nil }
func (*bug007LeaseStore) Create(context.Context, entity.Lease, entity.LeaseVersion, []entity.LeaseOccupant, string) error {
	return nil
}
func (*bug007LeaseStore) Renew(context.Context, entity.LeaseVersion, time.Time, string) error {
	return nil
}
func (*bug007LeaseStore) Versions(context.Context, string) ([]entity.LeaseVersion, error) {
	return nil, nil
}
func (*bug007LeaseStore) ChangeStatus(context.Context, string, string, string, string) error {
	return nil
}
func TestBug007_BusinessRegression(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	billing := &bug007BillingStore{}
	leases := &bug007LeaseStore{lease: entity.Lease{ID: "lease-1", Status: "active", StartDate: start.AddDate(0, -1, 0), EndDate: start.AddDate(0, 2, 0), MonthlyRent: 88000}}
	svc := service.BillingService{Repo: billing, Leases: leases}
	for i := 0; i < 2; i++ {
		if _, _, err := svc.GenerateMonthly(context.Background(), "lease-1", start, "request-key"); err != nil {
			t.Fatal(err)
		}
	}
	if len(billing.keys) != 2 || billing.keys[0] != "request-key" || billing.keys[1] != "request-key" {
		t.Fatalf("idempotency key changed across retries: %#v", billing.keys)
	}
}
