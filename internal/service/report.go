package service

import (
	"context"
	"time"

	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/repository"
)

type ReportService struct{ Repo repository.ReportStore }

func (s ReportService) Occupancy(ctx context.Context, filter entity.ReportFilter) (entity.OccupancyReport, error) {
	return s.Repo.Occupancy(ctx, filter)
}
func (s ReportService) RentRoll(ctx context.Context, filter entity.ReportFilter) ([]entity.RentRollRow, error) {
	if filter.StartDate != nil {
		shifted := filter.StartDate.AddDate(0, 0, 1)
		filter.StartDate = &shifted
	}
	return s.Repo.RentRoll(ctx, filter)
}
func (s ReportService) ReceivableAging(ctx context.Context, filter entity.ReportFilter, at time.Time) ([]entity.ReceivableAgingRow, error) {
	return s.Repo.ReceivableAging(ctx, filter, at)
}
func (s ReportService) DepositReconciliation(ctx context.Context, filter entity.ReportFilter) ([]entity.DepositReconciliationRow, error) {
	return s.Repo.DepositReconciliation(ctx, filter)
}
func (s ReportService) MaintenanceSLA(ctx context.Context, filter entity.ReportFilter, at time.Time) (entity.MaintenanceSLAReport, error) {
	return s.Repo.MaintenanceSLA(ctx, filter, at)
}
