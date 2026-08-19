package httptransport

import (
	"net/http"

	"go-property-rental/internal/transport/http/handler"
	"go-property-rental/internal/transport/http/middleware"
)

func Router(h handler.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("POST /api/auth/refresh", h.RefreshSession)

	authenticate := middleware.Auth(h.Auth)
	protected := func(pattern, permission string, fn http.HandlerFunc) {
		mux.Handle(pattern, authenticate(middleware.Require(permission)(fn)))
	}

	protected("GET /api/me", "", h.Me)
	protected("POST /api/auth/logout", "", h.Logout)
	protected("POST /api/me/password", "", h.ChangePassword)
	protected("GET /api/dashboard", "dashboard:read", h.DashboardSummary)

	protected("GET /api/users", "user:read", h.ListUsers)
	protected("POST /api/users", "user:create", h.CreateUser)
	protected("PUT /api/users/{id}", "user:update", h.UpdateUser)
	protected("PATCH /api/users/{id}/status", "user:status", h.SetUserStatus)
	protected("POST /api/users/{id}/reset-password", "user:reset_password", h.ResetUserPassword)
	protected("PUT /api/users/{id}/roles", "user:assign_role", h.ReplaceUserRoles)
	protected("GET /api/roles", "role:read", h.ListRoles)
	protected("POST /api/roles", "role:write", h.CreateRole)
	protected("PUT /api/roles/{id}", "role:write", h.UpdateRole)
	protected("DELETE /api/roles/{id}", "role:write", h.DeleteRole)
	protected("PUT /api/roles/{id}/permissions", "role:write", h.ReplaceRolePermissions)
	protected("GET /api/permissions", "permission:read", h.ListPermissions)
	protected("POST /api/permissions", "permission:write", h.CreatePermission)
	protected("PUT /api/permissions/{id}", "permission:write", h.UpdatePermission)

	protected("GET /api/properties", "property:read", h.ListProperties)
	protected("POST /api/properties", "property:create", h.CreateProperty)
	protected("PATCH /api/properties/{id}/status", "property:status", h.ChangePropertyStatus)
	protected("GET /api/facilities", "facility:read", h.ListFacilities)
	protected("POST /api/facilities", "facility:write", h.CreateFacility)
	protected("GET /api/properties/{id}/facilities", "facility:read", h.PropertyFacilities)
	protected("PUT /api/properties/{id}/facilities", "facility:write", h.ReplacePropertyFacilities)

	protected("GET /api/tenants", "tenant:read", h.ListTenants)
	protected("POST /api/tenants", "tenant:create", h.CreateTenant)
	protected("GET /api/tenants/{id}", "tenant:read", h.GetTenant)
	protected("PUT /api/tenants/{id}", "tenant:update", h.UpdateTenant)
	protected("PATCH /api/tenants/{id}/status", "tenant:status", h.SetTenantStatus)

	protected("GET /api/leases", "lease:read", h.ListLeases)
	protected("POST /api/leases", "lease:create", h.CreateLease)
	protected("GET /api/leases/{id}", "lease:read", h.GetLease)
	protected("POST /api/leases/{id}/renew", "lease:renew", h.RenewLease)
	protected("POST /api/leases/{id}/transition", "", h.TransitionLease)

	protected("GET /api/bills", "billing:read", h.ListBills)
	protected("POST /api/bills/generate", "billing:generate", h.GenerateBill)
	protected("GET /api/bills/{id}", "billing:read", h.GetBill)
	protected("POST /api/bills/{id}/adjustments", "billing:adjust", h.AdjustBill)
	protected("POST /api/bills/{id}/late-fee", "billing:adjust", h.ApplyLateFee)
	protected("POST /api/payments", "payment:post", h.RecordPayment)
	protected("GET /api/billing-plans", "billing:read", h.ListBillingPlans)
	protected("POST /api/billing-plans", "billing:generate", h.SaveBillingPlan)
	protected("GET /api/billing-plans/{id}", "billing:read", h.GetBillingPlan)
	protected("PUT /api/billing-plans/{id}", "billing:generate", h.SaveBillingPlan)
	protected("PATCH /api/billing-plans/{id}/status", "billing:generate", h.SetBillingPlanStatus)
	protected("POST /api/billing-runs/preview", "billing:generate", h.PreviewBillingRun)
	protected("POST /api/billing-runs", "billing:generate", h.ExecuteBillingRun)
	protected("GET /api/billing-runs", "billing:read", h.ListBillingRuns)
	protected("GET /api/billing-runs/{id}", "billing:read", h.GetBillingRun)
	protected("POST /api/billing-runs/{id}/items/{itemID}/retry", "billing:generate", h.RetryBillingRunItem)

	protected("GET /api/leases/{leaseID}/deposits", "deposit:read", h.DepositTransactions)
	protected("POST /api/leases/{leaseID}/deposits", "", h.CreateDepositTransaction)

	protected("GET /api/work-orders", "maintenance:read", h.ListWorkOrders)
	protected("POST /api/work-orders", "maintenance:create", h.CreateWorkOrder)
	protected("GET /api/work-orders/{id}", "maintenance:read", h.GetWorkOrder)
	protected("POST /api/work-orders/{id}/assign", "maintenance:assign", h.AssignWorkOrder)
	protected("POST /api/work-orders/{id}/materials", "maintenance:work", h.AddWorkOrderMaterial)
	protected("POST /api/work-orders/{id}/transition", "maintenance:work", h.TransitionWorkOrder)
	protected("POST /api/work-orders/{id}/confirm", "maintenance:confirm", h.ConfirmWorkOrder)

	protected("POST /api/settlements", "settlement:create", h.CreateSettlement)
	protected("GET /api/settlements/{id}", "settlement:read", h.GetSettlement)
	protected("POST /api/settlements/{id}/complete", "settlement:approve", h.CompleteSettlement)
	protected("GET /api/audit-logs", "audit:read", h.ListAudits)
	protected("GET /api/approvals", "approval:read", h.ListApprovals)
	protected("POST /api/approvals", "approval:submit", h.CreateApproval)
	protected("GET /api/approvals/{id}", "approval:read", h.GetApproval)
	protected("POST /api/approvals/{id}/decision", "approval:review", h.DecideApproval)
	protected("POST /api/approvals/{id}/cancel", "approval:submit", h.CancelApproval)
	protected("GET /api/notifications", "notification:read", h.ListNotifications)
	protected("POST /api/notifications/read-all", "notification:read", h.MarkAllNotificationsRead)
	protected("POST /api/notifications/{id}/read", "notification:read", h.MarkNotificationRead)
	protected("GET /api/reminder-rules", "notification:manage", h.ListReminderRules)
	protected("POST /api/reminder-rules", "notification:manage", h.SaveReminderRule)
	protected("PATCH /api/reminder-rules/{id}/enabled", "notification:manage", h.EnableReminderRule)
	protected("POST /api/reminders/run", "notification:manage", h.RunReminders)
	protected("GET /api/reports/occupancy", "report:read", h.OccupancyReport)
	protected("GET /api/reports/rent-roll", "report:read", h.RentRollReport)
	protected("GET /api/reports/receivable-aging", "report:read", h.ReceivableAgingReport)
	protected("GET /api/reports/deposit-reconciliation", "report:read", h.DepositReconciliationReport)
	protected("GET /api/reports/maintenance-sla", "report:read", h.MaintenanceSLAReport)
	protected("GET /api/reports/{type}/export", "report:export", h.ExportReport)
	return requestLimits(mux)
}

func requestLimits(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		next.ServeHTTP(w, r)
	})
}
