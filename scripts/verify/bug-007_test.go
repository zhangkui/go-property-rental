package verify

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"go-property-rental/internal/domain/entity"
	mysqlrepo "go-property-rental/internal/repository/mysql"
	"go-property-rental/internal/service"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type bug007BillingStore struct{ keys []string }

func (s *bug007BillingStore) Generate(_ context.Context, b entity.Bill, key string) (entity.Bill, bool, error) {
	s.keys = append(s.keys, key)
	return b, true, nil
}
func (*bug007BillingStore) List(context.Context, int, int, string, string) ([]entity.Bill, error) {
	return nil, nil
}
func (*bug007BillingStore) Get(context.Context, string) (entity.Bill, error) {
	return entity.Bill{}, nil
}
func (*bug007BillingStore) Items(context.Context, string) ([]entity.BillItem, error) { return nil, nil }
func (*bug007BillingStore) Adjust(context.Context, entity.BillAdjustment) error      { return nil }
func (*bug007BillingStore) RecordPayment(context.Context, entity.Payment, []entity.PaymentAllocation) error {
	return nil
}

type bug007LeaseStore struct{ lease entity.Lease }

func (*bug007LeaseStore) List(context.Context, int, int, string, string) ([]entity.Lease, error) {
	return nil, nil
}
func (s *bug007LeaseStore) Get(context.Context, string) (entity.Lease, error) { return s.lease, nil }
func (*bug007LeaseStore) Create(context.Context, entity.Lease, entity.LeaseVersion, []entity.LeaseOccupant, string) error {
	return nil
}
func (*bug007LeaseStore) Renew(context.Context, entity.LeaseVersion, time.Time, string) error {
	return nil
}
func (*bug007LeaseStore) Versions(context.Context, string) ([]entity.LeaseVersion, error) {
	return nil, nil
}
func (*bug007LeaseStore) ChangeStatus(context.Context, string, string, string, string) error {
	return nil
}

type bug007SQLState struct {
	queryScope  string
	billInserts int
	keyInserts  int
}

type bug007Driver struct{ state *bug007SQLState }
type bug007Conn struct{ state *bug007SQLState }
type bug007Tx struct{}
type bug007Rows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

var bug007DriverSeq atomic.Uint64

func (d bug007Driver) Open(string) (driver.Conn, error) { return &bug007Conn{state: d.state}, nil }
func (*bug007Conn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unsupported") }
func (*bug007Conn) Close() error                        { return nil }
func (*bug007Conn) Begin() (driver.Tx, error)           { return bug007Tx{}, nil }
func (*bug007Conn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return bug007Tx{}, nil
}
func (bug007Tx) Commit() error          { return nil }
func (bug007Tx) Rollback() error        { return nil }
func (r *bug007Rows) Columns() []string { return r.columns }
func (*bug007Rows) Close() error        { return nil }
func (r *bug007Rows) Next(destination []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(destination, r.values[r.index])
	r.index++
	return nil
}
func (c *bug007Conn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(query, "SELECT resource_id FROM idempotency_keys") {
		if strings.Contains(query, "scope='bill.generate'") {
			c.state.queryScope = "bill.generate"
			return &bug007Rows{columns: []string{"resource_id"}, values: [][]driver.Value{{"bill-existing"}}}, nil
		}
		c.state.queryScope = "bill.generate.manual"
		return &bug007Rows{columns: []string{"resource_id"}}, nil
	}
	if strings.Contains(query, "FROM bills WHERE id") {
		start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		return &bug007Rows{columns: []string{"id", "lease_id", "period_start", "period_end", "due_date", "status", "amount", "penalty", "discount", "paid"}, values: [][]driver.Value{{"bill-existing", "lease-1", start, start.AddDate(0, 1, 0), start.AddDate(0, 0, 5), "unpaid", int64(88000), int64(0), int64(0), int64(0)}}}, nil
	}
	return nil, fmt.Errorf("unexpected query: %s", query)
}
func (c *bug007Conn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if strings.Contains(query, "INSERT INTO bills") {
		c.state.billInserts++
	}
	if strings.Contains(query, "INSERT INTO idempotency_keys") {
		c.state.keyInserts++
	}
	return driver.RowsAffected(1), nil
}

func bug007DB(t *testing.T, state *bug007SQLState) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("bug007-%d", bug007DriverSeq.Add(1))
	sql.Register(name, bug007Driver{state: state})
	database, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database
}

func TestBug007_BusinessRegression(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	billing := &bug007BillingStore{}
	leases := &bug007LeaseStore{lease: entity.Lease{ID: "lease-1", Status: "active", StartDate: start.AddDate(0, -1, 0), EndDate: start.AddDate(0, 2, 0), MonthlyRent: 88000}}
	svc := service.BillingService{Repo: billing, Leases: leases}
	for i := 0; i < 2; i++ {
		if _, _, err := svc.GenerateMonthly(context.Background(), "lease-1", start, "request-key"); err != nil {
			t.Fatal(err)
		}
	}
	if len(billing.keys) != 2 || billing.keys[0] != "request-key" || billing.keys[1] != "request-key" {
		t.Fatalf("idempotency key changed across retries: %#v", billing.keys)
	}

	state := &bug007SQLState{}
	repository := mysqlrepo.Billing{DB: bug007DB(t, state)}
	candidate := entity.Bill{ID: "bill-new", LeaseID: "lease-1", PeriodStart: start, PeriodEnd: start.AddDate(0, 1, 0), DueDate: start.AddDate(0, 0, 5), Status: "unpaid", Amount: 88000}
	found, created, err := repository.Generate(context.Background(), candidate, "request-key")
	if err != nil {
		t.Fatal(err)
	}
	if created || found.ID != "bill-existing" {
		t.Fatalf("existing idempotent result was not reused: created=%v bill=%#v", created, found)
	}
	if state.queryScope != "bill.generate" {
		t.Fatalf("idempotency lookup used wrong scope: %q", state.queryScope)
	}
	if state.billInserts != 0 || state.keyInserts != 0 {
		t.Fatalf("duplicate request performed inserts: bills=%d keys=%d", state.billInserts, state.keyInserts)
	}
}

var _ driver.QueryerContext = (*bug007Conn)(nil)
var _ driver.ExecerContext = (*bug007Conn)(nil)
var _ driver.ConnBeginTx = (*bug007Conn)(nil)
