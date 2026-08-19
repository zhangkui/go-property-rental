package entity

import "time"

type AuditLog struct {
	ID, ActorID, Action, Resource, ResourceID, Detail string
	CreatedAt                                         time.Time
}
