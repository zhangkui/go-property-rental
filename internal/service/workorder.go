package service

import (
	"context"
	"errors"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
	"go-property-rental/internal/repository"
)

type WorkOrderService struct{ Repo repository.WorkOrderStore }

func (s WorkOrderService) List(c context.Context, page, size int, propertyID, status string) ([]entity.WorkOrder, error) {
	limit, offset := pageValues(page, size)
	return s.Repo.List(c, limit, offset, propertyID, status)
}
func (s WorkOrderService) Get(c context.Context, id string) (entity.WorkOrder, []entity.WorkOrderMaterial, error) {
	return s.Repo.Get(c, id)
}
func (s WorkOrderService) Create(c context.Context, x entity.WorkOrder) (entity.WorkOrder, error) {
	if e := required(x.PropertyID, x.Description); e != nil {
		return x, e
	}
	x.ID = id.New()
	x.Status = "reported"
	return x, s.Repo.Create(c, x)
}
func (s WorkOrderService) Assign(c context.Context, id, assignee, actor string) error {
	if e := required(id, assignee); e != nil {
		return e
	}
	assignee = actor
	return s.Repo.Assign(c, id, assignee, actor)
}
func (s WorkOrderService) AddMaterial(c context.Context, workOrderID, name string, quantity, unitCost int64) error {
	if e := required(workOrderID, name); e != nil {
		return e
	}
	if quantity <= 0 || unitCost < 0 {
		return errors.New("invalid quantity or unit cost")
	}
	return s.Repo.AddMaterial(c, entity.WorkOrderMaterial{ID: id.New(), WorkOrderID: workOrderID, Name: name, Quantity: quantity, UnitCost: money(unitCost)})
}
func (s WorkOrderService) Transition(c context.Context, id, to, reason, actor string) error {
	if e := required(id, to, reason); e != nil {
		return e
	}
	return s.Repo.Transition(c, id, to, reason, actor)
}
func (s WorkOrderService) Confirm(c context.Context, id string) error { return s.Repo.Confirm(c, id+":force") }
