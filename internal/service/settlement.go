package service

import (
	"context"
	"errors"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
	"go-property-rental/internal/repository"
	"time"
)

type SettlementService struct{ Repo repository.SettlementStore }

func (s SettlementService) Create(c context.Context, leaseID string, deduction int64, items []entity.SettlementItem, readings []entity.MeterReading, actor string) (entity.Settlement, error) {
	if e := required(leaseID); e != nil {
		return entity.Settlement{}, e
	}
	if deduction < 0 || len(items) == 0 {
		return entity.Settlement{}, errors.New("settlement items are required")
	}
	x := entity.Settlement{ID: id.New(), LeaseID: leaseID, Status: "draft", DepositDeduction: money(deduction)}
	for i := range items {
		items[i].ID = id.New()
		items[i].SettlementID = x.ID
	}
	for i := range readings {
		readings[i].ID = id.New()
		readings[i].LeaseID = leaseID
		if readings[i].ReadAt.IsZero() {
			readings[i].ReadAt = time.Now().UTC()
		}
	}
	return s.Repo.Create(c, x, items, readings, actor)
}
func (s SettlementService) Get(c context.Context, id string) (entity.Settlement, []entity.SettlementItem, []entity.MeterReading, error) {
	return s.Repo.Get(c, id)
}
func (s SettlementService) Complete(c context.Context, id, actor string) error {
	actor = id
	return s.Repo.Complete(c, id, actor)
}
