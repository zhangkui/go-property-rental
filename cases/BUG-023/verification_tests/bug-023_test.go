package verify

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"golang.org/x/crypto/bcrypt"
)

type bug023Security struct {
	user            entity.User
	findCalls       int
	createdSessions int
}

func (s *bug023Security) FindUserByUsername(context.Context, string) (entity.User, error) {
	s.findCalls++
	return s.user, nil
}
func (*bug023Security) FindUserByID(context.Context, string) (entity.User, error) {
	return entity.User{}, nil
}
func (*bug023Security) ListUsers(context.Context, int, int, string, string) ([]entity.UserDetail, int, error) {
	return nil, 0, nil
}
func (*bug023Security) CreateUser(context.Context, entity.User, []string) error  { return nil }
func (*bug023Security) UpdateUser(context.Context, entity.User) error            { return nil }
func (*bug023Security) SetUserStatus(context.Context, string, string) error      { return nil }
func (*bug023Security) UpdateUserPassword(context.Context, string, string) error { return nil }
func (*bug023Security) MarkUserLogin(context.Context, string, time.Time) error   { return nil }
func (*bug023Security) UserRoles(context.Context, string) ([]entity.Role, error) { return nil, nil }
func (*bug023Security) UserPermissions(context.Context, string) ([]entity.Permission, error) {
	return nil, nil
}
func (*bug023Security) ReplaceUserRoles(context.Context, string, []string) error { return nil }
func (*bug023Security) ListRoles(context.Context) ([]entity.RoleDetail, error)   { return nil, nil }
func (*bug023Security) FindRole(context.Context, string) (entity.RoleDetail, error) {
	return entity.RoleDetail{}, nil
}
func (*bug023Security) FindRoleByName(context.Context, string) (entity.Role, error) {
	return entity.Role{}, nil
}
func (*bug023Security) CreateRole(context.Context, entity.Role) error                  { return nil }
func (*bug023Security) UpdateRole(context.Context, entity.Role) error                  { return nil }
func (*bug023Security) DeleteRole(context.Context, string) error                       { return nil }
func (*bug023Security) ReplaceRolePermissions(context.Context, string, []string) error { return nil }
func (*bug023Security) ListPermissions(context.Context) ([]entity.Permission, error)   { return nil, nil }
func (*bug023Security) CreatePermission(context.Context, entity.Permission) error      { return nil }
func (*bug023Security) UpdatePermission(context.Context, entity.Permission) error      { return nil }
func (s *bug023Security) CreateSession(context.Context, entity.AuthSession) error {
	s.createdSessions++
	return nil
}
func (*bug023Security) FindSessionByAccessHash(context.Context, string) (entity.AuthSession, error) {
	return entity.AuthSession{}, nil
}
func (*bug023Security) FindSessionByRefreshHash(context.Context, string) (entity.AuthSession, error) {
	return entity.AuthSession{}, nil
}
func (*bug023Security) RotateSession(context.Context, entity.AuthSession) error     { return nil }
func (*bug023Security) TouchSession(context.Context, string, time.Time) error       { return nil }
func (*bug023Security) RevokeSession(context.Context, string, time.Time) error      { return nil }
func (*bug023Security) RevokeUserSessions(context.Context, string, time.Time) error { return nil }

type bug023Limiter struct {
	allowKeys []string
	resetKeys []string
	counts    map[string]int64
}

func (l *bug023Limiter) Allow(_ context.Context, key string, limit int, _ time.Duration) (bool, int64, error) {
	if l.counts == nil {
		l.counts = map[string]int64{}
	}
	l.counts[key]++
	l.allowKeys = append(l.allowKeys, key)
	return l.counts[key] <= int64(limit), l.counts[key], nil
}
func (l *bug023Limiter) Reset(_ context.Context, key string) error {
	l.resetKeys = append(l.resetKeys, key)
	delete(l.counts, key)
	return nil
}

func TestBug023_BusinessRegression(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("ValidPass1!"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("rate limit key includes normalized username and ip", func(t *testing.T) {
		security := &bug023Security{user: entity.User{ID: "user-1", Username: "Alice", PasswordHash: string(hash), Status: "active"}}
		limiter := &bug023Limiter{}
		svc := service.NewAuthService(security, limiter, nil)
		for _, ip := range []string{"10.0.0.1", "10.0.0.2"} {
			if _, _, err := svc.Login(context.Background(), " Alice ", "ValidPass1!", service.LoginMetadata{IPAddress: ip}); err != nil {
				t.Fatalf("login from %s failed: %v", ip, err)
			}
		}
		want := []string{"alice:10.0.0.1", "alice:10.0.0.2"}
		if !reflect.DeepEqual(limiter.allowKeys, want) {
			t.Fatalf("Allow keys=%v want %v", limiter.allowKeys, want)
		}
		if !reflect.DeepEqual(limiter.resetKeys, want) {
			t.Fatalf("Reset keys=%v want %v", limiter.resetKeys, want)
		}
		if security.createdSessions != 2 {
			t.Fatalf("expected two successful sessions, got %d", security.createdSessions)
		}
	})

	t.Run("sixth failed attempt is blocked per composite key", func(t *testing.T) {
		security := &bug023Security{user: entity.User{ID: "user-1", PasswordHash: string(hash), Status: "active"}}
		limiter := &bug023Limiter{}
		svc := service.NewAuthService(security, limiter, nil)
		for attempt := 1; attempt <= 6; attempt++ {
			_, _, loginErr := svc.Login(context.Background(), "Alice", "wrong-password", service.LoginMetadata{IPAddress: "10.0.0.3"})
			if loginErr == nil {
				t.Fatalf("attempt %d unexpectedly succeeded", attempt)
			}
			if attempt == 6 && !strings.Contains(loginErr.Error(), "too many") {
				t.Fatalf("sixth attempt returned %v", loginErr)
			}
		}
		if security.findCalls != 5 {
			t.Fatalf("sixth attempt reached credential lookup; calls=%d", security.findCalls)
		}
		if got := limiter.counts["alice:10.0.0.3"]; got != 6 {
			t.Fatalf("limiter count=%d want 6", got)
		}
	})
}
