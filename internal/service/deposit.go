package service

import (
	"context"
	"errors"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
	"go-property-rental/internal/repository"
)

type DepositService struct{ Repo repository.DepositStore }

func (s DepositService) Transaction(c context.Context, leaseID, kind, reference, reason, actor string, amount int64) (entity.DepositLedger, error) {
	if e := required(leaseID, kind, reference, reason); e != nil {
		return entity.DepositLedger{}, e
	}
	if amount <= 0 {
		return entity.DepositLedger{}, errors.New("deposit amount must be positive")
	}
	ledgerID := id.New()
	x := entity.DepositLedger{ID: ledgerID, LeaseID: leaseID, Kind: kind, Reference: reference + ":" + ledgerID, Reason: reason, ActorID: actor, Amount: money(amount)}
	return x, s.Repo.Append(c, x)
}
func (s DepositService) List(c context.Context, leaseID string) ([]entity.DepositLedger, error) {
	return s.Repo.List(c, leaseID)
}
func (s DepositService) Balance(c context.Context, leaseID string) (entity.DepositBalance, error) {
	return s.Repo.Balance(c, leaseID)
}
