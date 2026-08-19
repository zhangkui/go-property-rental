package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-property-rental/internal/config"
	"go-property-rental/internal/platform/bootstrap"
	"go-property-rental/internal/platform/migration"
	platformmysql "go-property-rental/internal/platform/mysql"
	platformredis "go-property-rental/internal/platform/redis"
	mysqlrepo "go-property-rental/internal/repository/mysql"
	redisrepo "go-property-rental/internal/repository/redis"
	"go-property-rental/internal/service"
	httptransport "go-property-rental/internal/transport/http"
	"go-property-rental/internal/transport/http/handler"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	db, err := platformmysql.Open(cfg.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err = migration.Run(ctx, db); err != nil {
		log.Fatal(err)
	}
	if err = bootstrap.Admin(ctx, db, cfg.AdminUser, cfg.AdminPassword); err != nil {
		log.Fatal(err)
	}
	redisClient := platformredis.Open(cfg.RedisAddr, cfg.RedisPassword)
	defer redisClient.Close()
	if err = redisClient.Ping(ctx); err != nil {
		log.Printf("redis unavailable at startup: %v", err)
	}

	properties := mysqlrepo.Properties{DB: db}
	tenants := mysqlrepo.Tenants{DB: db}
	facilities := mysqlrepo.Facilities{DB: db}
	leases := mysqlrepo.LeaseStore{DB: db}
	billing := mysqlrepo.Billing{DB: db}
	billingPlans := mysqlrepo.BillingPlans{DB: db}
	deposits := mysqlrepo.Deposits{DB: db}
	workOrders := mysqlrepo.WorkOrders{DB: db}
	settlements := mysqlrepo.Settlements{DB: db}
	audits := mysqlrepo.Audits{DB: db}
	security := mysqlrepo.Security{DB: db}
	approvals := mysqlrepo.Approvals{DB: db}
	notifications := mysqlrepo.Notifications{DB: db}
	reports := mysqlrepo.Reports{DB: db}
	dashboard := mysqlrepo.Dashboard{DB: db}
	loginLimiter := redisrepo.Security{Client: redisClient}
	dashboardCache := redisrepo.Dashboard{Client: redisClient}
	billingPlanLock := redisrepo.BillingPlans{Client: redisClient}
	authService := service.NewAuthService(security, loginLimiter, audits)
	notificationService := service.NotificationService{Repo: notifications}
	billingPlanService := service.BillingPlanService{Repo: billingPlans, Lock: billingPlanLock, Leases: leases}

	h := handler.Handler{
		Auth:          authService,
		RBAC:          service.RBACService{Security: security, Audits: audits},
		Properties:    service.PropertyService{Repo: properties},
		Tenants:       service.TenantService{Repo: tenants},
		Facilities:    service.FacilityService{Repo: facilities},
		Leases:        service.LeaseService{Repo: leases, Tenants: tenants},
		Billing:       service.BillingService{Repo: billing, Leases: leases},
		BillingPlans:  billingPlanService,
		Deposits:      service.DepositService{Repo: deposits},
		WorkOrders:    service.WorkOrderService{Repo: workOrders},
		Settlements:   service.SettlementService{Repo: settlements},
		Audits:        service.AuditService{Repo: audits},
		Approvals:     service.ApprovalService{Repo: approvals, Audits: audits},
		Notifications: notificationService,
		Reports:       service.ReportService{Repo: reports},
		Dashboard:     service.DashboardService{Repo: dashboard, Cache: dashboardCache},
	}
	runtimeCtx, stopRuntime := context.WithCancel(context.Background())
	defer stopRuntime()
	service.ReminderScheduler{Notifications: notificationService}.Start(runtimeCtx)
	service.BillingPlanScheduler{Plans: billingPlanService}.Start(runtimeCtx)
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: httptransport.Router(h), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		log.Printf("api listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	stopRuntime()
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdown)
}
