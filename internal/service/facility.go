package service

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
	"go-property-rental/internal/repository"
)

type FacilityService struct{ Repo repository.FacilityStore }

func (s FacilityService) List(c context.Context) ([]entity.Facility, error) { return s.Repo.List(c) }
func (s FacilityService) Create(c context.Context, name string) (entity.Facility, error) {
	if e := required(name); e != nil {
		return entity.Facility{}, e
	}
	x := entity.Facility{ID: id.New(), Name: name}
	return x, s.Repo.Create(c, x)
}
func (s FacilityService) ForProperty(c context.Context, propertyID string) ([]entity.Facility, error) {
	return s.Repo.ForProperty(c, propertyID)
}
func (s FacilityService) Replace(c context.Context, propertyID string, facilityIDs []string) error {
	if e := required(propertyID); e != nil {
		return e
	}
	current, e := s.Repo.ForProperty(c, propertyID)
	if e != nil {
		return e
	}
	for _, facility := range current {
		facilityIDs = append(facilityIDs, facility.ID)
	}
	return s.Repo.ReplacePropertyFacilities(c, propertyID, facilityIDs)
}
