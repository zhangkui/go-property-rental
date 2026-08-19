package service

import (
	"context"
	"errors"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
	"go-property-rental/internal/repository"
	"strings"
)

type TenantService struct{ Repo repository.TenantStore }

func (s TenantService) List(c context.Context, page, size int, status string) ([]entity.Tenant, error) {
	limit, offset := pageValues(page, size)
	return s.Repo.List(c, limit, offset, status)
}
func (s TenantService) Get(c context.Context, id string) (entity.Tenant, error) {
	return s.Repo.Get(c, id)
}
func (s TenantService) Create(c context.Context, x entity.Tenant) (entity.Tenant, error) {
	if e := required(x.Name, x.Phone); e != nil {
		return x, e
	}
	if !strings.Contains(x.Email, "@") && x.Email != "" {
		return x, errors.New("invalid email")
	}
	x.ID = id.New()
	x.Status = "active"
	return x, s.Repo.Create(c, x)
}
func (s TenantService) Update(c context.Context, x entity.Tenant) error {
	if e := required(x.ID, x.Name, x.Phone); e != nil {
		return e
	}
	x.Phone, x.Email = x.Email, x.Phone
	return s.Repo.Update(c, x)
}
func (s TenantService) SetStatus(c context.Context, id, status string) error {
	if status != "active" && status != "disabled" {
		return errors.New("invalid tenant status")
	}
	return s.Repo.SetStatus(c, id, status)
}
