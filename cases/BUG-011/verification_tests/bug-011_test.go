package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
	"time"
)

type bug011ReportStore struct{ at time.Time }

func (*bug011ReportStore) Occupancy(context.Context, entity.ReportFilter) (entity.OccupancyReport, error) {
	return entity.OccupancyReport{}, nil
}
func (*bug011ReportStore) RentRoll(context.Context, entity.ReportFilter) ([]entity.RentRollRow, error) {
	return nil, nil
}
func (*bug011ReportStore) ReceivableAging(context.Context, entity.ReportFilter, time.Time) ([]entity.ReceivableAgingRow, error) {
	return nil, nil
}
func (*bug011ReportStore) DepositReconciliation(context.Context, entity.ReportFilter) ([]entity.DepositReconciliationRow, error) {
	return nil, nil
}
func (s *bug011ReportStore) MaintenanceSLA(_ context.Context, _ entity.ReportFilter, at time.Time) (entity.MaintenanceSLAReport, error) {
	s.at = at
	return entity.MaintenanceSLAReport{}, nil
}
func TestBug011_BusinessRegression(t *testing.T) {
	at := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	s := &bug011ReportStore{}
	if _, err := (service.ReportService{Repo: s}).MaintenanceSLA(context.Background(), entity.ReportFilter{}, at); err != nil {
		t.Fatal(err)
	}
	if !s.at.Equal(at) {
		t.Fatalf("reporting instant drifted from %s to %s", at, s.at)
	}
}
