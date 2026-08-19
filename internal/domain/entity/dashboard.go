package entity

import "time"

type DashboardPropertyMetrics struct {
	Total, Available, Occupied, Maintenance int
	OccupancyBasisPoints                    int64
}

type DashboardLeaseMetrics struct {
	Total, Draft, Active, Expiring30Days, Renewed int
}

type DashboardFinanceMetrics struct {
	Outstanding, OverdueOutstanding int64
	OverdueBills, DueWithin7Days    int
	ReceivedThisMonth               int64
	ReceivedPreviousMonth           int64
	DepositBalance                  int64
}

type DashboardOperationMetrics struct {
	ActiveTenants, OpenWorkOrders, OverdueWorkOrders int
	PendingSettlements, PendingApprovals             int
	UnreadNotifications                              int
}

type DashboardCashflowPoint struct {
	Month    string
	Received int64
	Billed   int64
}

type DashboardStatusPoint struct {
	Name  string
	Value int64
}

type DashboardUpcomingLease struct {
	LeaseID, PropertyLabel, TenantName string
	EndDate                            time.Time
	DaysRemaining                      int
}

type DashboardOverdueBill struct {
	BillID, PropertyLabel, TenantName string
	DueDate                           time.Time
	Outstanding                       int64
	OverdueDays                       int
}

type DashboardSummary struct {
	GeneratedAt        time.Time
	Properties         DashboardPropertyMetrics
	Leases             DashboardLeaseMetrics
	Finance            DashboardFinanceMetrics
	Operations         DashboardOperationMetrics
	Cashflow           []DashboardCashflowPoint
	LeaseStatuses      []DashboardStatusPoint
	UpcomingExpiries   []DashboardUpcomingLease
	TopOverdueBills    []DashboardOverdueBill
	RecentAuditActions []AuditLog
}
