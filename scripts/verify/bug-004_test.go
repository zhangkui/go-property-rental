package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
	"time"
)

type bug004LeaseStore struct {
	to, reason, actor string
	calls             int
}

func (*bug004LeaseStore) List(context.Context, int, int, string, string) ([]entity.Lease, error) {
	return nil, nil
}
func (*bug004LeaseStore) Get(context.Context, string) (entity.Lease, error) {
	return entity.Lease{}, nil
}
func (*bug004LeaseStore) Create(context.Context, entity.Lease, entity.LeaseVersion, []entity.LeaseOccupant, string) error {
	return nil
}
func (*bug004LeaseStore) Renew(context.Context, entity.LeaseVersion, time.Time, string) error {
	return nil
}
func (*bug004LeaseStore) Versions(context.Context, string) ([]entity.LeaseVersion, error) {
	return nil, nil
}
func (s *bug004LeaseStore) ChangeStatus(_ context.Context, _ string, to, reason, actor string) error {
	s.calls++
	s.to = to
	s.reason = reason
	s.actor = actor
	return nil
}
func TestBug004_BusinessRegression(t *testing.T) {
	s := &bug004LeaseStore{}
	svc := service.LeaseService{Repo: s}
	if err := svc.Transition(context.Background(), "lease-1", "pending", "submit", "admin"); err != nil {
		t.Fatal(err)
	}
	if s.to != "pending" || s.reason != "submit" || s.actor != "admin" || s.calls != 1 {
		t.Fatalf("transition corrupted: %#v", s)
	}
	if err := svc.Transition(context.Background(), "lease-1", "active", "activate", "admin"); err != nil {
		t.Fatal(err)
	}
	if s.to != "active" {
		t.Fatalf("legal target changed to %q", s.to)
	}
}
