package entity

import (
	"go-property-rental/internal/domain/valueobject"
	"time"
)

type MeterReading struct {
	ID, LeaseID, Kind string
	Reading           valueobject.Money
	ReadAt            time.Time
}
type Settlement struct {
	ID, LeaseID, Status             string
	Total, DepositDeduction, Refund valueobject.Money
	SettledAt                       *time.Time
	CreatedAt                       time.Time
}
type SettlementItem struct {
	ID, SettlementID, Kind, Description string
	Amount                              valueobject.Money
}

func (s Settlement) NetPayable() valueobject.Money { return s.Total.Sub(s.DepositDeduction) }
