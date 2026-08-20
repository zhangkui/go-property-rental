package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
)

type bug017WorkStore struct{ assignee, actor string }

func (*bug017WorkStore) List(context.Context, int, int, string, string) ([]entity.WorkOrder, error) {
	return nil, nil
}
func (*bug017WorkStore) Get(context.Context, string) (entity.WorkOrder, []entity.WorkOrderMaterial, error) {
	return entity.WorkOrder{}, nil, nil
}
func (*bug017WorkStore) Create(context.Context, entity.WorkOrder) error { return nil }
func (s *bug017WorkStore) Assign(_ context.Context, _ string, assignee, actor string) error {
	s.assignee = assignee
	s.actor = actor
	return nil
}
func (*bug017WorkStore) AddMaterial(context.Context, entity.WorkOrderMaterial) error      { return nil }
func (*bug017WorkStore) Transition(context.Context, string, string, string, string) error { return nil }
func (*bug017WorkStore) Confirm(context.Context, string) error                            { return nil }
func TestBug017_BusinessRegression(t *testing.T) {
	s := &bug017WorkStore{}
	if err := (service.WorkOrderService{Repo: s}).Assign(context.Background(), "work-1", "technician-1", "admin-1"); err != nil {
		t.Fatal(err)
	}
	if s.assignee != "technician-1" || s.actor != "admin-1" {
		t.Fatalf("assignee and actor crossed: assignee=%q actor=%q", s.assignee, s.actor)
	}
}
