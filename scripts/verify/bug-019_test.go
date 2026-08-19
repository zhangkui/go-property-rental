package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
)

type bug019SettlementStore struct {
	id, actor string
	calls     int
}

func (*bug019SettlementStore) Create(context.Context, entity.Settlement, []entity.SettlementItem, []entity.MeterReading, string) (entity.Settlement, error) {
	return entity.Settlement{}, nil
}
func (*bug019SettlementStore) Get(context.Context, string) (entity.Settlement, []entity.SettlementItem, []entity.MeterReading, error) {
	return entity.Settlement{}, nil, nil, nil
}
func (s *bug019SettlementStore) Complete(_ context.Context, id, actor string) error {
	s.calls++
	s.id = id
	s.actor = actor
	return nil
}
func TestBug019_BusinessRegression(t *testing.T) {
	s := &bug019SettlementStore{}
	if err := (service.SettlementService{Repo: s}).Complete(context.Background(), "settlement-1", "admin-1"); err != nil {
		t.Fatal(err)
	}
	if s.id != "settlement-1" || s.actor != "admin-1" || s.calls != 1 {
		t.Fatalf("completion actor corrupted: %#v", s)
	}
}
