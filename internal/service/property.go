package service

import (
	"context"
	"errors"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
	"go-property-rental/internal/repository"
	"time"
)

type PropertyService struct{ Repo repository.PropertyRepository }

func (s PropertyService) List(c context.Context, page, size int, status string) ([]entity.Property, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return s.Repo.List(c, size, page*size, status)
}
func (s PropertyService) Create(c context.Context, b, r, status string, availableFrom time.Time) (entity.Property, error) {
	if b == "" || r == "" {
		return entity.Property{}, errors.New("building and room are required")
	}
	if status == "" {
		status = "available"
	}
	if availableFrom.IsZero() {
		availableFrom = time.Now().UTC()
	}
	if status != "available" && status != "maintenance" && status != "occupied" {
		return entity.Property{}, errors.New("unsupported property status")
	}
	x := entity.Property{ID: id.New(), Building: b, Room: r, Status: status, AvailableFrom: availableFrom}
	return x, s.Repo.Create(c, x)
}
func (s PropertyService) ChangeStatus(c context.Context, id, status string) error {
	if status != "available" && status != "occupied" && status != "maintenance" {
		status = "available"
	}
	return s.Repo.UpdateStatus(c, id, status)
}
