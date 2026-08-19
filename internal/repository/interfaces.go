package repository

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"time"
)

type UserRepository interface {
	FindByUsername(context.Context, string) (entity.User, error)
	FindByID(context.Context, string) (entity.User, error)
	Create(context.Context, entity.User) error
	UpdatePassword(context.Context, string, string) error
}
type PropertyRepository interface {
	List(context.Context, int, int, string) ([]entity.Property, error)
	Get(context.Context, string) (entity.Property, error)
	Create(context.Context, entity.Property) error
	UpdateStatus(context.Context, string, string) error
}
type TenantRepository interface {
	List(context.Context, int, int) ([]entity.Tenant, error)
	Create(context.Context, entity.Tenant) error
}
type LeaseRepository interface {
	Create(context.Context, entity.Lease) error
	Overlaps(context.Context, string, time.Time, time.Time) (bool, error)
}
