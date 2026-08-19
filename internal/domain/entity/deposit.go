package entity

import (
	"go-property-rental/internal/domain/valueobject"
	"time"
)

type DepositLedger struct {
	ID, LeaseID, Kind, Reference, Reason, ActorID string
	Amount                                        valueobject.Money
	CreatedAt                                     time.Time
}
type DepositBalance struct {
	LeaseID                      string
	Received, Deducted, Refunded valueobject.Money
}

func (d DepositBalance) Available() valueobject.Money {
	return d.Received.Sub(d.Deducted).Sub(d.Refunded)
}
