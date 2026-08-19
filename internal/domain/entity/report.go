package entity

import "time"

type OccupancyReport struct {
	TotalUnits, OccupiedUnits, VacantUnits, MaintenanceUnits int
	OccupancyBasisPoints                                     int64
}

type RentRollRow struct {
	PropertyID, Building, Room, LeaseID, TenantName, LeaseStatus string
	StartDate, EndDate                                           time.Time
	MonthlyRent, Deposit                                         int64
}

type ReceivableAgingRow struct {
	LeaseID, TenantName, PropertyLabel string
	Current, Days1To30, Days31To60     int64
	Days61To90, DaysOver90, Total      int64
}

type DepositReconciliationRow struct {
	LeaseID, TenantName, PropertyLabel string
	ContractDeposit, LedgerBalance     int64
	Difference                         int64
}

type MaintenanceSLAReport struct {
	Total, Open, Completed, Overdue int
	AverageResolutionHours          int64
}

type ReportFilter struct {
	PropertyID string
	StartDate  *time.Time
	EndDate    *time.Time
}
