package entity

import "time"

type Tenant struct {
	ID, Name, Phone, Email, IdentityNo, Status string
	CreatedAt, UpdatedAt                       time.Time
}

func (t Tenant) CanLease() bool { return t.Status == "active" }
