package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
	"time"
)

type bug005LeaseStore struct {
	lease   entity.Lease
	version entity.LeaseVersion
	end     time.Time
	calls   int
}

func (*bug005LeaseStore) List(context.Context, int, int, string, string) ([]entity.Lease, error) {
	return nil, nil
}
func (s *bug005LeaseStore) Get(context.Context, string) (entity.Lease, error) { return s.lease, nil }
func (*bug005LeaseStore) Create(context.Context, entity.Lease, entity.LeaseVersion, []entity.LeaseOccupant, string) error {
	return nil
}
func (s *bug005LeaseStore) Renew(_ context.Context, v entity.LeaseVersion, end time.Time, _ string) error {
	s.calls++
	s.version = v
	s.end = end
	return nil
}
func (*bug005LeaseStore) Versions(context.Context, string) ([]entity.LeaseVersion, error) {
	return nil, nil
}
func (*bug005LeaseStore) ChangeStatus(context.Context, string, string, string, string) error {
	return nil
}
func TestBug005_BusinessRegression(t *testing.T) {
	old := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	next := time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC)
	s := &bug005LeaseStore{lease: entity.Lease{ID: "lease-1", EndDate: old}}
	if err := (service.LeaseService{Repo: s}).Renew(context.Background(), "lease-1", next, 120000, 50000, "admin"); err != nil {
		t.Fatal(err)
	}
	if int64(s.version.MonthlyRent) != 120000 || int64(s.version.Deposit) != 50000 || !s.end.Equal(next) {
		t.Fatalf("renewal values crossed: %#v", s.version)
	}
	if err := (service.LeaseService{Repo: s}).Renew(context.Background(), "lease-1", old, 120000, 50000, "admin"); err == nil {
		t.Fatal("expected non-increasing end date to fail")
	}
	if s.calls != 1 {
		t.Fatalf("invalid renewal reached repository, calls=%d", s.calls)
	}
}
