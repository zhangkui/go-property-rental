package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
	"time"
)

type bug028ReportStore struct{ filter entity.ReportFilter }

func (*bug028ReportStore) Occupancy(context.Context, entity.ReportFilter) (entity.OccupancyReport, error) {
	return entity.OccupancyReport{}, nil
}
func (s *bug028ReportStore) RentRoll(_ context.Context, f entity.ReportFilter) ([]entity.RentRollRow, error) {
	s.filter = f
	return nil, nil
}
func (*bug028ReportStore) ReceivableAging(context.Context, entity.ReportFilter, time.Time) ([]entity.ReceivableAgingRow, error) {
	return nil, nil
}
func (*bug028ReportStore) DepositReconciliation(context.Context, entity.ReportFilter) ([]entity.DepositReconciliationRow, error) {
	return nil, nil
}
func (*bug028ReportStore) MaintenanceSLA(context.Context, entity.ReportFilter, time.Time) (entity.MaintenanceSLAReport, error) {
	return entity.MaintenanceSLAReport{}, nil
}
func TestBug028_BusinessRegression(t *testing.T) {
	date := time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)
	s := &bug028ReportStore{}
	_, err := (service.ReportService{Repo: s}).RentRoll(context.Background(), entity.ReportFilter{StartDate: &date, EndDate: &date})
	if err != nil {
		t.Fatal(err)
	}
	if s.filter.StartDate == nil || !s.filter.StartDate.Equal(date) {
		t.Fatalf("inclusive start boundary shifted to %v", s.filter.StartDate)
	}
}
