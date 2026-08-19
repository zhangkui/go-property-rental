package handler

import (
	"net/http"
	"testing"
)

func TestReportFilterValidation(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "/api/reports/rent-roll?start_date=2026-02-01&end_date=2026-02-28", nil)
	filter, err := reportFilter(request)
	if err != nil || filter.StartDate == nil || filter.EndDate == nil {
		t.Fatalf("expected valid date filter, got %#v %v", filter, err)
	}
	request, _ = http.NewRequest(http.MethodGet, "/api/reports/rent-roll?start_date=2026-03-01&end_date=2026-02-01", nil)
	if _, err = reportFilter(request); err == nil {
		t.Fatal("expected invalid range to fail")
	}
}
