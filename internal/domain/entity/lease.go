package entity

import (
	"go-property-rental/internal/domain/valueobject"
	"time"
)

type Lease struct {
	ID, PropertyID, TenantID, Status string
	StartDate, EndDate               time.Time
	MonthlyRent, Deposit             valueobject.Money
}
type LeaseVersion struct {
	ID, LeaseID          string
	VersionNo            int
	MonthlyRent, Deposit valueobject.Money
	StartDate, EndDate   time.Time
	CreatedAt            time.Time
}
type LeaseOccupant struct {
	ID, LeaseID, Name, Phone, IdentityNo string
	Primary                              bool
}
type LeaseStateHistory struct {
	ID, LeaseID, FromStatus, ToStatus, Reason, ActorID string
	CreatedAt                                          time.Time
}

func (l Lease) OutstandingDays(at time.Time) int {
	if !at.After(l.EndDate) {
		return 0
	}
	return int(at.Sub(l.EndDate).Hours() / 24)
}
