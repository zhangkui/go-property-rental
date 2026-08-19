package verify

import (
	"context"
	"testing"
	"time"

	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
)

type bug021Security struct {
	user            entity.User
	session         entity.AuthSession
	statusUserID    string
	status          string
	revokedUserID   string
	permissionCalls int
	rotateCalls     int
}

func (s *bug021Security) FindUserByUsername(context.Context, string) (entity.User, error) {
	return s.user, nil
}
func (s *bug021Security) FindUserByID(context.Context, string) (entity.User, error) {
	return s.user, nil
}
func (*bug021Security) ListUsers(context.Context, int, int, string, string) ([]entity.UserDetail, int, error) {
	return nil, 0, nil
}
func (*bug021Security) CreateUser(context.Context, entity.User, []string) error { return nil }
func (*bug021Security) UpdateUser(context.Context, entity.User) error           { return nil }
func (s *bug021Security) SetUserStatus(_ context.Context, userID, status string) error {
	s.statusUserID, s.status = userID, status
	return nil
}
func (*bug021Security) UpdateUserPassword(context.Context, string, string) error { return nil }
func (*bug021Security) MarkUserLogin(context.Context, string, time.Time) error   { return nil }
func (*bug021Security) UserRoles(context.Context, string) ([]entity.Role, error) { return nil, nil }
func (s *bug021Security) UserPermissions(context.Context, string) ([]entity.Permission, error) {
	s.permissionCalls++
	return nil, nil
}
func (*bug021Security) ReplaceUserRoles(context.Context, string, []string) error { return nil }
func (*bug021Security) ListRoles(context.Context) ([]entity.RoleDetail, error)   { return nil, nil }
func (*bug021Security) FindRole(context.Context, string) (entity.RoleDetail, error) {
	return entity.RoleDetail{}, nil
}
func (*bug021Security) FindRoleByName(context.Context, string) (entity.Role, error) {
	return entity.Role{}, nil
}
func (*bug021Security) CreateRole(context.Context, entity.Role) error                  { return nil }
func (*bug021Security) UpdateRole(context.Context, entity.Role) error                  { return nil }
func (*bug021Security) DeleteRole(context.Context, string) error                       { return nil }
func (*bug021Security) ReplaceRolePermissions(context.Context, string, []string) error { return nil }
func (*bug021Security) ListPermissions(context.Context) ([]entity.Permission, error)   { return nil, nil }
func (*bug021Security) CreatePermission(context.Context, entity.Permission) error      { return nil }
func (*bug021Security) UpdatePermission(context.Context, entity.Permission) error      { return nil }
func (*bug021Security) CreateSession(context.Context, entity.AuthSession) error        { return nil }
func (s *bug021Security) FindSessionByAccessHash(context.Context, string) (entity.AuthSession, error) {
	return s.session, nil
}
func (s *bug021Security) FindSessionByRefreshHash(context.Context, string) (entity.AuthSession, error) {
	return s.session, nil
}
func (s *bug021Security) RotateSession(context.Context, entity.AuthSession) error {
	s.rotateCalls++
	return nil
}
func (*bug021Security) TouchSession(context.Context, string, time.Time) error  { return nil }
func (*bug021Security) RevokeSession(context.Context, string, time.Time) error { return nil }
func (s *bug021Security) RevokeUserSessions(_ context.Context, userID string, _ time.Time) error {
	s.revokedUserID = userID
	return nil
}

func TestBug021_BusinessRegression(t *testing.T) {
	t.Run("disabling target revokes target sessions", func(t *testing.T) {
		security := &bug021Security{}
		svc := service.RBACService{Security: security}
		err := svc.SetUserStatus(context.Background(), entity.AuthIdentity{UserID: "actor-1"}, "target-1", "disabled")
		if err != nil {
			t.Fatalf("disable failed: %v", err)
		}
		if security.statusUserID != "target-1" || security.status != "disabled" {
			t.Fatalf("wrong status update: %q %q", security.statusUserID, security.status)
		}
		if security.revokedUserID != "target-1" {
			t.Fatalf("disabled user's sessions not revoked: %q", security.revokedUserID)
		}
	})

	t.Run("disabled access and refresh are rejected", func(t *testing.T) {
		now := time.Now().UTC()
		security := &bug021Security{
			user:    entity.User{ID: "user-1", Username: "disabled", Status: "disabled"},
			session: entity.AuthSession{ID: "session-1", UserID: "user-1", AccessExpiresAt: now.Add(time.Hour), RefreshExpiresAt: now.Add(time.Hour), LastSeenAt: now},
		}
		svc := service.AuthService{Security: security, TouchInterval: time.Hour}
		if _, err := svc.Authenticate(context.Background(), "access-token"); err == nil {
			t.Fatal("disabled user authenticated with access token")
		}
		if _, err := svc.Refresh(context.Background(), "refresh-token"); err == nil {
			t.Fatal("disabled user refreshed a session")
		}
		if security.permissionCalls != 0 || security.rotateCalls != 0 {
			t.Fatalf("disabled user caused downstream work: permissions=%d rotates=%d", security.permissionCalls, security.rotateCalls)
		}
	})

	t.Run("active access and refresh remain valid", func(t *testing.T) {
		now := time.Now().UTC()
		security := &bug021Security{
			user:    entity.User{ID: "user-1", Username: "active", Status: "active"},
			session: entity.AuthSession{ID: "session-1", UserID: "user-1", AccessExpiresAt: now.Add(time.Hour), RefreshExpiresAt: now.Add(time.Hour), LastSeenAt: now},
		}
		svc := service.AuthService{Security: security, TouchInterval: time.Hour, AccessTTL: time.Hour, RefreshTTL: time.Hour}
		if _, err := svc.Authenticate(context.Background(), "access-token"); err != nil {
			t.Fatalf("active access rejected: %v", err)
		}
		if _, err := svc.Refresh(context.Background(), "refresh-token"); err != nil {
			t.Fatalf("active refresh rejected: %v", err)
		}
		if security.permissionCalls != 1 || security.rotateCalls != 1 {
			t.Fatalf("active path calls mismatch: permissions=%d rotates=%d", security.permissionCalls, security.rotateCalls)
		}
	})
}
