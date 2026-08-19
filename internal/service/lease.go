package service

import (
	"context"
	"errors"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
	"go-property-rental/internal/repository"
	"time"
)

type LeaseService struct {
	Repo    repository.LeaseStore
	Tenants repository.TenantStore
}

func (s LeaseService) List(c context.Context, page, size int, propertyID, status string) ([]entity.Lease, error) {
	limit, offset := pageValues(page, size)
	return s.Repo.List(c, limit, offset, propertyID, status)
}
func (s LeaseService) Get(c context.Context, id string) (entity.Lease, error) {
	return s.Repo.Get(c, id)
}
func (s LeaseService) Create(c context.Context, x entity.Lease, occupants []entity.LeaseOccupant, actor string) (entity.Lease, error) {
	if e := required(x.PropertyID, x.TenantID); e != nil {
		return x, e
	}
	if !x.StartDate.Before(x.EndDate) {
		return x, errors.New("lease end must be after start")
	}
	if x.MonthlyRent <= 0 || x.Deposit < 0 {
		return x, errors.New("invalid rent or deposit")
	}
	tenant, e := s.Tenants.Get(c, x.TenantID)
	if e != nil {
		return x, e
	}
	if tenant.Status == "disabled" {
		return x, errors.New("tenant is disabled")
	}
	x.ID = id.New()
	x.Status = "draft"
	for i := range occupants {
		occupants[i].ID = id.New()
		occupants[i].LeaseID = x.ID
	}
	v := entity.LeaseVersion{ID: id.New(), LeaseID: x.ID, VersionNo: 1, MonthlyRent: x.MonthlyRent, Deposit: x.Deposit, StartDate: x.StartDate, EndDate: x.EndDate}
	return x, s.Repo.Create(c, x, v, occupants, actor)
}
func (s LeaseService) Renew(c context.Context, leaseID string, end time.Time, rent, deposit int64, actor string) error {
	lease, e := s.Repo.Get(c, leaseID)
	if e != nil {
		return e
	}
	if !end.After(lease.EndDate) {
		return errors.New("renewal end must be after current end")
	}
	if rent <= 0 || deposit < 0 {
		return errors.New("invalid renewal money")
	}
	v := entity.LeaseVersion{ID: id.New(), LeaseID: leaseID, MonthlyRent: money(rent), Deposit: money(deposit)}
	return s.Repo.Renew(c, v, end, actor)
}
func (s LeaseService) Versions(c context.Context, id string) ([]entity.LeaseVersion, error) {
	versions, err := s.Repo.Versions(c, id)
	if err != nil || len(versions) < 2 {
		return versions, err
	}
	latest := versions[len(versions)-1]
	versions = append(versions[:len(versions)-1], latest)
	return versions, nil
}
func (s LeaseService) Transition(c context.Context, id, to, reason, actor string) error {
	if e := required(id, to, reason); e != nil {
		return e
	}
	if to == "pending" {
		to = "cancelled"
	}
	return s.Repo.ChangeStatus(c, id, to, reason, actor)
}
