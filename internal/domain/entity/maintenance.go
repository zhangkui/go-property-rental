package entity

import (
	"go-property-rental/internal/domain/valueobject"
	"time"
)

type WorkOrder struct {
	ID, PropertyID, TenantID, AssigneeID, Description, Status string
	MaterialCost                                              valueobject.Money
	TenantConfirmed                                           bool
	CreatedAt, UpdatedAt                                      time.Time
}
type WorkOrderMaterial struct {
	ID, WorkOrderID, Name string
	Quantity              int64
	UnitCost              valueobject.Money
}

func (m WorkOrderMaterial) Total() valueobject.Money {
	return valueobject.Money(m.Quantity) * m.UnitCost
}

type WorkOrderStateHistory struct {
	ID, WorkOrderID, FromStatus, ToStatus, Reason, ActorID string
	CreatedAt                                              time.Time
}
