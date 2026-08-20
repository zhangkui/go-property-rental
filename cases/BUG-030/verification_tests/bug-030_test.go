package verify

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go-property-rental/internal/domain/entity"
	mysqlrepo "go-property-rental/internal/repository/mysql"
	"go-property-rental/internal/service"
)

type bug030Store struct {
	item          entity.BillingRunItem
	bundle        entity.BillingPlanBundle
	getPlanCalls  int
	generateCalls int
	generatedItem entity.BillingRunItem
}

func (*bug030Store) ListPlans(context.Context, int, int, string) ([]entity.BillingPlanBundle, error) {
	return nil, nil
}
func (s *bug030Store) GetPlan(context.Context, string) (entity.BillingPlanBundle, error) {
	s.getPlanCalls++
	return s.bundle, nil
}
func (*bug030Store) SavePlan(context.Context, entity.BillingPlan, []entity.BillingPlanItem) error {
	return nil
}
func (*bug030Store) SetPlanStatus(context.Context, string, string) error { return nil }
func (*bug030Store) DuePlans(context.Context, time.Time) ([]entity.BillingPlanBundle, error) {
	return nil, nil
}
func (*bug030Store) CreateRun(context.Context, entity.BillingRun) error              { return nil }
func (*bug030Store) ListRuns(context.Context, int, int) ([]entity.BillingRun, error) { return nil, nil }
func (*bug030Store) GetRun(context.Context, string) (entity.BillingRun, []entity.BillingRunItem, error) {
	return entity.BillingRun{}, nil, nil
}
func (s *bug030Store) GetRunItem(context.Context, string) (entity.BillingRunItem, error) {
	return s.item, nil
}
func (s *bug030Store) Generate(_ context.Context, item entity.BillingRunItem, _ entity.Bill, _ []entity.BillItem, _ string) (string, bool, error) {
	s.generateCalls++
	s.generatedItem = item
	return "bill-1", true, nil
}
func (*bug030Store) RecordFailure(context.Context, entity.BillingRunItem) error        { return nil }
func (*bug030Store) FinishRun(context.Context, string, int, int, int, time.Time) error { return nil }

type bug030SQLState struct{ execQueries []string }
type bug030Driver struct{ state *bug030SQLState }
type bug030Conn struct{ state *bug030SQLState }
type bug030Tx struct{}
type bug030Rows struct{ done bool }

func (d bug030Driver) Open(string) (driver.Conn, error) { return &bug030Conn{state: d.state}, nil }
func (*bug030Conn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unsupported") }
func (*bug030Conn) Close() error                        { return nil }
func (*bug030Conn) Begin() (driver.Tx, error)           { return bug030Tx{}, nil }
func (c *bug030Conn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return &bug030Rows{}, nil
}
func (c *bug030Conn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.state.execQueries = append(c.state.execQueries, query)
	return driver.RowsAffected(1), nil
}
func (bug030Tx) Commit() error        { return nil }
func (bug030Tx) Rollback() error      { return nil }
func (*bug030Rows) Columns() []string { return []string{"resource_id"} }
func (*bug030Rows) Close() error      { return nil }
func (r *bug030Rows) Next(values []driver.Value) error {
	if r.done {
		return io.EOF
	}
	values[0] = "existing-bill"
	r.done = true
	return nil
}

var bug030Sequence atomic.Uint64

func bug030DB(t *testing.T, state *bug030SQLState) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("bug030-%d", bug030Sequence.Add(1))
	sql.Register(name, bug030Driver{state: state})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestBug030_BusinessRegression(t *testing.T) {
	period := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	t.Run("retry regenerates the failed period", func(t *testing.T) {
		store := &bug030Store{
			item:   entity.BillingRunItem{ID: "item-1", RunID: "run-1", PlanID: "plan-1", LeaseID: "lease-1", Status: "failed", PeriodStart: period},
			bundle: entity.BillingPlanBundle{Plan: entity.BillingPlan{ID: "plan-1", NextPeriodStart: period.AddDate(0, 5, 0)}, Lease: entity.Lease{ID: "lease-1", EndDate: period.AddDate(1, 0, 0), MonthlyRent: 100000}},
		}
		if err := (service.BillingPlanService{Repo: store}).Retry(context.Background(), "item-1"); err != nil {
			t.Fatalf("retry failed: %v", err)
		}
		if store.generateCalls != 1 {
			t.Fatalf("Generate calls=%d want 1", store.generateCalls)
		}
		if !store.generatedItem.PeriodStart.Equal(period) {
			t.Fatalf("retried period=%s want %s", store.generatedItem.PeriodStart, period)
		}
		if store.generatedItem.ID != "item-1" || store.generatedItem.RunID != "run-1" {
			t.Fatalf("retry identity changed: %#v", store.generatedItem)
		}
	})

	t.Run("nonfailed item is rejected without generation", func(t *testing.T) {
		store := &bug030Store{item: entity.BillingRunItem{Status: "succeeded", PeriodStart: period}}
		if err := (service.BillingPlanService{Repo: store}).Retry(context.Background(), "item-1"); err == nil {
			t.Fatal("succeeded item was retried")
		}
		if store.getPlanCalls != 0 || store.generateCalls != 0 {
			t.Fatalf("rejected retry caused work: plans=%d generate=%d", store.getPlanCalls, store.generateCalls)
		}
	})

	t.Run("repository advances plan by one month", func(t *testing.T) {
		state := &bug030SQLState{}
		repo := mysqlrepo.BillingPlans{DB: bug030DB(t, state)}
		item := entity.BillingRunItem{ID: "item-1", RunID: "run-1", PlanID: "plan-1", LeaseID: "lease-1", PeriodStart: period, PeriodEnd: period.AddDate(0, 1, 0), DueDate: period.AddDate(0, 0, 5), Amount: 100000, Attempts: 1}
		bill := entity.Bill{ID: "bill-1", LeaseID: "lease-1", PeriodStart: period, PeriodEnd: period.AddDate(0, 1, 0), DueDate: period.AddDate(0, 0, 5), Amount: 100000}
		if _, created, err := repo.Generate(context.Background(), item, bill, nil, "plan-1:2026-07-01"); err != nil || created {
			t.Fatalf("existing generation path failed: created=%v err=%v", created, err)
		}
		if len(state.execQueries) != 2 {
			t.Fatalf("expected run-item and plan updates, got %d", len(state.execQueries))
		}
		if !strings.Contains(state.execQueries[1], "INTERVAL 1 MONTH") {
			t.Fatalf("plan advancement is not one month: %s", state.execQueries[1])
		}
	})
}

var _ driver.QueryerContext = (*bug030Conn)(nil)
var _ driver.ExecerContext = (*bug030Conn)(nil)
