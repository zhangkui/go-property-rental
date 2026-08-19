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
	"sync/atomic"
	"testing"
	"time"
)

type bug008BillingStore struct {
	bill        entity.Bill
	adjustments []entity.BillAdjustment
}

func (*bug008BillingStore) Generate(context.Context, entity.Bill, string) (entity.Bill, bool, error) {
	return entity.Bill{}, false, nil
}
func (*bug008BillingStore) List(context.Context, int, int, string, string) ([]entity.Bill, error) {
	return nil, nil
}
func (s *bug008BillingStore) Get(context.Context, string) (entity.Bill, error) { return s.bill, nil }
func (*bug008BillingStore) Items(context.Context, string) ([]entity.BillItem, error) {
	return nil, nil
}
func (s *bug008BillingStore) Adjust(_ context.Context, adjustment entity.BillAdjustment) error {
	s.adjustments = append(s.adjustments, adjustment)
	return nil
}
func (*bug008BillingStore) RecordPayment(context.Context, entity.Payment, []entity.PaymentAllocation) error {
	return nil
}

type bug008SQLState struct{ executions int }
type bug008Driver struct{ state *bug008SQLState }
type bug008Conn struct{ state *bug008SQLState }
type bug008Tx struct{}
type bug008Rows struct{ sent bool }

var bug008DriverSeq atomic.Uint64

func (driverInstance bug008Driver) Open(string) (driver.Conn, error) {
	return &bug008Conn{state: driverInstance.state}, nil
}
func (*bug008Conn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unsupported") }
func (*bug008Conn) Close() error                        { return nil }
func (*bug008Conn) Begin() (driver.Tx, error)           { return bug008Tx{}, nil }
func (*bug008Conn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return bug008Tx{}, nil
}
func (bug008Tx) Commit() error        { return nil }
func (bug008Tx) Rollback() error      { return nil }
func (*bug008Rows) Columns() []string { return []string{"status"} }
func (*bug008Rows) Close() error      { return nil }
func (rows *bug008Rows) Next(destination []driver.Value) error {
	if rows.sent {
		return io.EOF
	}
	destination[0] = "paid"
	rows.sent = true
	return nil
}
func (*bug008Conn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return &bug008Rows{}, nil
}
func (connection *bug008Conn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	connection.state.executions++
	return driver.RowsAffected(1), nil
}
func bug008DB(t *testing.T, state *bug008SQLState) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("bug008-%d", bug008DriverSeq.Add(1))
	sql.Register(name, bug008Driver{state: state})
	database, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database
}

func TestBug008_BusinessRegression(t *testing.T) {
	due := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	paid := &bug008BillingStore{bill: entity.Bill{ID: "bill-1", Status: "paid", DueDate: due, Amount: 10000, Paid: 10000}}
	if err := (service.BillingService{Repo: paid}).ApplyLateFee(context.Background(), "bill-1", due.AddDate(0, 0, 10), 100, 30, "admin"); err != nil {
		t.Fatal(err)
	}
	if len(paid.adjustments) != 0 {
		t.Fatalf("paid bill received late-fee adjustments: %#v", paid.adjustments)
	}
	unpaid := &bug008BillingStore{bill: entity.Bill{ID: "bill-2", Status: "unpaid", DueDate: due, Amount: 10000}}
	if err := (service.BillingService{Repo: unpaid}).ApplyLateFee(context.Background(), "bill-2", due.AddDate(0, 0, 2), 100, 30, "admin"); err != nil {
		t.Fatal(err)
	}
	if len(unpaid.adjustments) != 1 || int64(unpaid.adjustments[0].Amount) != 200 {
		t.Fatalf("overdue unpaid bill did not receive expected fee: %#v", unpaid.adjustments)
	}
	boundary := &bug008BillingStore{bill: entity.Bill{ID: "bill-3", Status: "unpaid", DueDate: due, Amount: 10000}}
	_ = (service.BillingService{Repo: boundary}).ApplyLateFee(context.Background(), "bill-3", due, 100, 30, "admin")
	if len(boundary.adjustments) != 0 {
		t.Fatal("bill received fee exactly at due date")
	}

	state := &bug008SQLState{}
	repository := mysqlrepo.Billing{DB: bug008DB(t, state)}
	err := repository.Adjust(context.Background(), entity.BillAdjustment{ID: "adjustment-1", BillID: "bill-1", Kind: "penalty", Amount: 100, Reason: "late", ActorID: "admin"})
	if err == nil {
		t.Fatal("paid bill accepted repository penalty adjustment")
	}
	if state.executions != 0 {
		t.Fatalf("paid bill adjustment performed writes: %d", state.executions)
	}
}

var _ driver.QueryerContext = (*bug008Conn)(nil)
var _ driver.ExecerContext = (*bug008Conn)(nil)
var _ driver.ConnBeginTx = (*bug008Conn)(nil)
