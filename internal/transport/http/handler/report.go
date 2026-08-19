package handler

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-property-rental/internal/domain/entity"
)

func reportFilter(r *http.Request) (entity.ReportFilter, error) {
	filter := entity.ReportFilter{PropertyID: r.URL.Query().Get("property_id")}
	if value := r.URL.Query().Get("start_date"); value != "" {
		parsed, err := parseDate(value)
		if err != nil {
			return filter, err
		}
		filter.StartDate = &parsed
	}
	if value := r.URL.Query().Get("end_date"); value != "" {
		parsed, err := parseDate(value)
		if err != nil {
			return filter, err
		}
		filter.EndDate = &parsed
	}
	if filter.StartDate != nil && filter.EndDate != nil && filter.StartDate.After(*filter.EndDate) {
		return filter, errors.New("start_date must not be after end_date")
	}
	return filter, nil
}

func (h Handler) OccupancyReport(w http.ResponseWriter, r *http.Request) {
	filter, err := reportFilter(r)
	if err != nil {
		fail(w, 400, err)
		return
	}
	result, err := h.Reports.Occupancy(r.Context(), filter)
	if err != nil {
		fail(w, 500, err)
		return
	}
	write(w, 200, result)
}
func (h Handler) RentRollReport(w http.ResponseWriter, r *http.Request) {
	filter, err := reportFilter(r)
	if err != nil {
		fail(w, 400, err)
		return
	}
	result, err := h.Reports.RentRoll(r.Context(), filter)
	if err != nil {
		fail(w, 500, err)
		return
	}
	write(w, 200, map[string]any{"items": result})
}
func (h Handler) ReceivableAgingReport(w http.ResponseWriter, r *http.Request) {
	filter, err := reportFilter(r)
	if err != nil {
		fail(w, 400, err)
		return
	}
	result, err := h.Reports.ReceivableAging(r.Context(), filter, time.Now().UTC())
	if err != nil {
		fail(w, 500, err)
		return
	}
	write(w, 200, map[string]any{"items": result})
}
func (h Handler) DepositReconciliationReport(w http.ResponseWriter, r *http.Request) {
	filter, err := reportFilter(r)
	if err != nil {
		fail(w, 400, err)
		return
	}
	result, err := h.Reports.DepositReconciliation(r.Context(), filter)
	if err != nil {
		fail(w, 500, err)
		return
	}
	write(w, 200, map[string]any{"items": result})
}
func (h Handler) MaintenanceSLAReport(w http.ResponseWriter, r *http.Request) {
	filter, err := reportFilter(r)
	if err != nil {
		fail(w, 400, err)
		return
	}
	result, err := h.Reports.MaintenanceSLA(r.Context(), filter, time.Now().UTC())
	if err != nil {
		fail(w, 500, err)
		return
	}
	write(w, 200, result)
}

func (h Handler) ExportReport(w http.ResponseWriter, r *http.Request) {
	filter, err := reportFilter(r)
	if err != nil {
		fail(w, 400, err)
		return
	}
	reportType := r.PathValue("type")
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%s.csv"`, reportType, time.Now().Format("20060102")))
	_, _ = w.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(w)
	defer writer.Flush()
	switch reportType {
	case "rent-roll":
		err = h.exportRentRoll(r, writer, filter)
	case "receivable-aging":
		err = h.exportReceivableAging(r, writer, filter)
	case "deposit-reconciliation":
		err = h.exportDepositReconciliation(r, writer, filter)
	default:
		err = errors.New("unsupported report type")
	}
	if err != nil {
		return
	}
}

func (h Handler) exportRentRoll(r *http.Request, w *csv.Writer, filter entity.ReportFilter) error {
	items, err := h.Reports.RentRoll(r.Context(), filter)
	if err != nil {
		return err
	}
	_ = w.Write([]string{"楼栋", "房间", "租客", "租约状态", "开始日期", "结束日期", "月租金(分)", "押金(分)"})
	for _, x := range items {
		_ = w.Write([]string{x.Building, x.Room, x.TenantName, x.LeaseStatus, x.StartDate.Format("2006-01-02"), x.EndDate.Format("2006-01-02"), strconv.FormatInt(x.MonthlyRent, 10), strconv.FormatInt(x.Deposit, 10)})
	}
	return w.Error()
}

func (h Handler) exportReceivableAging(r *http.Request, w *csv.Writer, filter entity.ReportFilter) error {
	items, err := h.Reports.ReceivableAging(r.Context(), filter, time.Now().UTC())
	if err != nil {
		return err
	}
	_ = w.Write([]string{"租客", "房源", "未到期", "1-30天", "31-60天", "61-90天", "90天以上", "合计(分)"})
	for _, x := range items {
		_ = w.Write([]string{x.TenantName, x.PropertyLabel, strconv.FormatInt(x.Current, 10), strconv.FormatInt(x.Days1To30, 10), strconv.FormatInt(x.Days31To60, 10), strconv.FormatInt(x.Days61To90, 10), strconv.FormatInt(x.DaysOver90, 10), strconv.FormatInt(x.Total, 10)})
	}
	return w.Error()
}

func (h Handler) exportDepositReconciliation(r *http.Request, w *csv.Writer, filter entity.ReportFilter) error {
	items, err := h.Reports.DepositReconciliation(r.Context(), filter)
	if err != nil {
		return err
	}
	_ = w.Write([]string{"租客", "房源", "合同押金(分)", "流水余额(分)", "差异(分)"})
	for _, x := range items {
		_ = w.Write([]string{x.TenantName, x.PropertyLabel, strconv.FormatInt(x.ContractDeposit, 10), strconv.FormatInt(x.LedgerBalance, 10), strconv.FormatInt(x.Difference, 10)})
	}
	return w.Error()
}
