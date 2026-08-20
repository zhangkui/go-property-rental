package verify

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"go-property-rental/internal/domain/entity"
	mysqlrepo "go-property-rental/internal/repository/mysql"
	"go-property-rental/internal/service"
)

type bug022Security struct {
	replaceUserID string
	roleIDs       []string
	revokedUserID string
}

func (*bug022Security) FindUserByUsername(context.Context, string) (entity.User, error) {
	return entity.User{}, nil
}
func (*bug022Security) FindUserByID(context.Context, string) (entity.User, error) {
	return entity.User{}, nil
}
func (*bug022Security) ListUsers(context.Context, int, int, string, string) ([]entity.UserDetail, int, error) {
	return nil, 0, nil
}
func (*bug022Security) CreateUser(context.Context, entity.User, []string) error  { return nil }
func (*bug022Security) UpdateUser(context.Context, entity.User) error            { return nil }
func (*bug022Security) SetUserStatus(context.Context, string, string) error      { return nil }
func (*bug022Security) UpdateUserPassword(context.Context, string, string) error { return nil }
func (*bug022Security) MarkUserLogin(context.Context, string, time.Time) error   { return nil }
func (*bug022Security) UserRoles(context.Context, string) ([]entity.Role, error) { return nil, nil }
func (*bug022Security) UserPermissions(context.Context, string) ([]entity.Permission, error) {
	return nil, nil
}
func (s *bug022Security) ReplaceUserRoles(_ context.Context, userID string, roleIDs []string) error {
	s.replaceUserID = userID
	s.roleIDs = append([]string(nil), roleIDs...)
	return nil
}
func (*bug022Security) ListRoles(context.Context) ([]entity.RoleDetail, error) { return nil, nil }
func (*bug022Security) FindRole(context.Context, string) (entity.RoleDetail, error) {
	return entity.RoleDetail{}, nil
}
func (*bug022Security) FindRoleByName(context.Context, string) (entity.Role, error) {
	return entity.Role{}, nil
}
func (*bug022Security) CreateRole(context.Context, entity.Role) error                  { return nil }
func (*bug022Security) UpdateRole(context.Context, entity.Role) error                  { return nil }
func (*bug022Security) DeleteRole(context.Context, string) error                       { return nil }
func (*bug022Security) ReplaceRolePermissions(context.Context, string, []string) error { return nil }
func (*bug022Security) ListPermissions(context.Context) ([]entity.Permission, error)   { return nil, nil }
func (*bug022Security) CreatePermission(context.Context, entity.Permission) error      { return nil }
func (*bug022Security) UpdatePermission(context.Context, entity.Permission) error      { return nil }
func (*bug022Security) CreateSession(context.Context, entity.AuthSession) error        { return nil }
func (*bug022Security) FindSessionByAccessHash(context.Context, string) (entity.AuthSession, error) {
	return entity.AuthSession{}, nil
}
func (*bug022Security) FindSessionByRefreshHash(context.Context, string) (entity.AuthSession, error) {
	return entity.AuthSession{}, nil
}
func (*bug022Security) RotateSession(context.Context, entity.AuthSession) error { return nil }
func (*bug022Security) TouchSession(context.Context, string, time.Time) error   { return nil }
func (*bug022Security) RevokeSession(context.Context, string, time.Time) error  { return nil }
func (s *bug022Security) RevokeUserSessions(_ context.Context, userID string, _ time.Time) error {
	s.revokedUserID = userID
	return nil
}

type bug022SQLState struct{ execArgs [][]driver.NamedValue }
type bug022Driver struct{ state *bug022SQLState }
type bug022Conn struct{ state *bug022SQLState }
type bug022Tx struct{}

func (d bug022Driver) Open(string) (driver.Conn, error) { return &bug022Conn{state: d.state}, nil }
func (*bug022Conn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unsupported") }
func (*bug022Conn) Close() error                        { return nil }
func (*bug022Conn) Begin() (driver.Tx, error)           { return bug022Tx{}, nil }
func (c *bug022Conn) ExecContext(_ context.Context, _ string, args []driver.NamedValue) (driver.Result, error) {
	copied := append([]driver.NamedValue(nil), args...)
	c.state.execArgs = append(c.state.execArgs, copied)
	return driver.RowsAffected(1), nil
}
func (bug022Tx) Commit() error   { return nil }
func (bug022Tx) Rollback() error { return nil }

var bug022Sequence atomic.Uint64

func bug022DB(t *testing.T, state *bug022SQLState) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("bug022-%d", bug022Sequence.Add(1))
	sql.Register(name, bug022Driver{state: state})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestBug022_BusinessRegression(t *testing.T) {
	t.Run("service updates and revokes target user", func(t *testing.T) {
		security := &bug022Security{}
		svc := service.RBACService{Security: security}
		err := svc.ReplaceUserRoles(context.Background(), entity.AuthIdentity{UserID: "actor-1"}, "target-1", []string{" role-a ", "role-b", "role-a", ""})
		if err != nil {
			t.Fatalf("replace roles failed: %v", err)
		}
		if security.replaceUserID != "target-1" || !reflect.DeepEqual(security.roleIDs, []string{"role-a", "role-b"}) {
			t.Fatalf("wrong role replacement: user=%q roles=%v", security.replaceUserID, security.roleIDs)
		}
		if security.revokedUserID != "target-1" {
			t.Fatalf("revoked sessions for %q, want target-1", security.revokedUserID)
		}
	})

	t.Run("repository deletes roles by user id", func(t *testing.T) {
		state := &bug022SQLState{}
		repo := mysqlrepo.Security{DB: bug022DB(t, state)}
		if err := repo.ReplaceUserRoles(context.Background(), "target-1", []string{"role-a", "role-b"}); err != nil {
			t.Fatalf("repository replace failed: %v", err)
		}
		if len(state.execArgs) != 3 {
			t.Fatalf("expected delete plus two inserts, got %d", len(state.execArgs))
		}
		if len(state.execArgs[0]) != 1 || state.execArgs[0][0].Value != "target-1" {
			t.Fatalf("delete used wrong argument: %#v", state.execArgs[0])
		}
		if len(state.execArgs[1]) != 2 || state.execArgs[1][0].Value != "target-1" || state.execArgs[1][1].Value != "role-a" {
			t.Fatalf("first insert args wrong: %#v", state.execArgs[1])
		}
	})
}

var _ driver.ExecerContext = (*bug022Conn)(nil)
