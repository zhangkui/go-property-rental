package verify

import (
	"context"
	"testing"
	"time"

	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"golang.org/x/crypto/bcrypt"
)

type bug024Security struct {
	user          entity.User
	updatedUserID string
	updatedHash   string
	updateCalls   int
	revokedUserID string
	revokeCalls   int
}

func (*bug024Security) FindUserByUsername(context.Context, string) (entity.User, error) {
	return entity.User{}, nil
}
func (s *bug024Security) FindUserByID(context.Context, string) (entity.User, error) {
	return s.user, nil
}
func (*bug024Security) ListUsers(context.Context, int, int, string, string) ([]entity.UserDetail, int, error) {
	return nil, 0, nil
}
func (*bug024Security) CreateUser(context.Context, entity.User, []string) error { return nil }
func (*bug024Security) UpdateUser(context.Context, entity.User) error           { return nil }
func (*bug024Security) SetUserStatus(context.Context, string, string) error     { return nil }
func (s *bug024Security) UpdateUserPassword(_ context.Context, userID, hash string) error {
	s.updatedUserID, s.updatedHash = userID, hash
	s.updateCalls++
	return nil
}
func (*bug024Security) MarkUserLogin(context.Context, string, time.Time) error   { return nil }
func (*bug024Security) UserRoles(context.Context, string) ([]entity.Role, error) { return nil, nil }
func (*bug024Security) UserPermissions(context.Context, string) ([]entity.Permission, error) {
	return nil, nil
}
func (*bug024Security) ReplaceUserRoles(context.Context, string, []string) error { return nil }
func (*bug024Security) ListRoles(context.Context) ([]entity.RoleDetail, error)   { return nil, nil }
func (*bug024Security) FindRole(context.Context, string) (entity.RoleDetail, error) {
	return entity.RoleDetail{}, nil
}
func (*bug024Security) FindRoleByName(context.Context, string) (entity.Role, error) {
	return entity.Role{}, nil
}
func (*bug024Security) CreateRole(context.Context, entity.Role) error                  { return nil }
func (*bug024Security) UpdateRole(context.Context, entity.Role) error                  { return nil }
func (*bug024Security) DeleteRole(context.Context, string) error                       { return nil }
func (*bug024Security) ReplaceRolePermissions(context.Context, string, []string) error { return nil }
func (*bug024Security) ListPermissions(context.Context) ([]entity.Permission, error)   { return nil, nil }
func (*bug024Security) CreatePermission(context.Context, entity.Permission) error      { return nil }
func (*bug024Security) UpdatePermission(context.Context, entity.Permission) error      { return nil }
func (*bug024Security) CreateSession(context.Context, entity.AuthSession) error        { return nil }
func (*bug024Security) FindSessionByAccessHash(context.Context, string) (entity.AuthSession, error) {
	return entity.AuthSession{}, nil
}
func (*bug024Security) FindSessionByRefreshHash(context.Context, string) (entity.AuthSession, error) {
	return entity.AuthSession{}, nil
}
func (*bug024Security) RotateSession(context.Context, entity.AuthSession) error { return nil }
func (*bug024Security) TouchSession(context.Context, string, time.Time) error   { return nil }
func (*bug024Security) RevokeSession(context.Context, string, time.Time) error  { return nil }
func (s *bug024Security) RevokeUserSessions(_ context.Context, userID string, _ time.Time) error {
	s.revokedUserID = userID
	s.revokeCalls++
	return nil
}

func TestBug024_BusinessRegression(t *testing.T) {
	current := "CurrentPass1!"
	next := "NextSecure2@"
	hash, err := bcrypt.GenerateFromPassword([]byte(current), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("valid change stores next password for current user", func(t *testing.T) {
		security := &bug024Security{user: entity.User{ID: "user-1", PasswordHash: string(hash), Status: "active"}}
		svc := service.AuthService{Security: security}
		if err := svc.ChangePassword(context.Background(), entity.AuthIdentity{UserID: "user-1"}, current, next); err != nil {
			t.Fatalf("change password failed: %v", err)
		}
		if security.updatedUserID != "user-1" || security.updateCalls != 1 {
			t.Fatalf("wrong update target/calls: %q/%d", security.updatedUserID, security.updateCalls)
		}
		if bcrypt.CompareHashAndPassword([]byte(security.updatedHash), []byte(next)) != nil {
			t.Fatal("stored hash does not accept new password")
		}
		if bcrypt.CompareHashAndPassword([]byte(security.updatedHash), []byte(current)) == nil {
			t.Fatal("stored hash still accepts old password")
		}
		if security.revokedUserID != "user-1" || security.revokeCalls != 1 {
			t.Fatalf("sessions not revoked for current user: %q/%d", security.revokedUserID, security.revokeCalls)
		}
	})

	t.Run("wrong current password has no side effects", func(t *testing.T) {
		security := &bug024Security{user: entity.User{ID: "user-1", PasswordHash: string(hash), Status: "active"}}
		svc := service.AuthService{Security: security}
		if err := svc.ChangePassword(context.Background(), entity.AuthIdentity{UserID: "user-1"}, "WrongPass9!", next); err == nil {
			t.Fatal("wrong current password was accepted")
		}
		if security.updateCalls != 0 || security.revokeCalls != 0 {
			t.Fatalf("rejected change caused side effects: updates=%d revokes=%d", security.updateCalls, security.revokeCalls)
		}
	})
}
