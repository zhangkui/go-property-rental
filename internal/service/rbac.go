package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
	"go-property-rental/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type RBACService struct {
	Security repository.SecurityStore
	Audits   repository.AuditStore
}

func (s RBACService) ListUsers(ctx context.Context, page, size int, status, keyword string) ([]entity.UserDetail, int, error) {
	limit, offset := pageValues(page, size)
	return s.Security.ListUsers(ctx, limit, offset, status, strings.TrimSpace(keyword))
}

func (s RBACService) CreateUser(ctx context.Context, actor entity.AuthIdentity, username, password, displayName, email string, roleIDs []string) (entity.User, error) {
	if err := validateUsername(strings.TrimSpace(username)); err != nil {
		return entity.User{}, err
	}
	if err := validatePassword(password); err != nil {
		return entity.User{}, err
	}
	if len(roleIDs) == 0 {
		return entity.User{}, errors.New("at least one role is required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return entity.User{}, err
	}
	now := time.Now().UTC()
	user := entity.User{ID: id.New(), Username: strings.TrimSpace(username), PasswordHash: string(hash), DisplayName: strings.TrimSpace(displayName), Email: strings.TrimSpace(email), Status: "active", CreatedAt: now, UpdatedAt: now}
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}
	if err := s.Security.CreateUser(ctx, user, uniqueStrings(roleIDs)); err != nil {
		return entity.User{}, err
	}
	_ = s.audit(ctx, actor, "rbac.user.created", "user", user.ID, map[string]any{"roles": roleIDs})
	return user, nil
}

func (s RBACService) UpdateProfile(ctx context.Context, actor entity.AuthIdentity, userID, displayName, email string) error {
	user, err := s.Security.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}
	user.DisplayName = strings.TrimSpace(displayName)
	user.Email = strings.TrimSpace(email)
	if user.DisplayName == "" {
		return errors.New("display name is required")
	}
	if err := s.Security.UpdateUser(ctx, user); err != nil {
		return err
	}
	return s.audit(ctx, actor, "rbac.user.updated", "user", userID, map[string]any{"display_name": user.DisplayName, "email": user.Email})
}

func (s RBACService) SetUserStatus(ctx context.Context, actor entity.AuthIdentity, userID, status string) error {
	if actor.UserID == userID && status == "disabled" {
		return errors.New("cannot disable current user")
	}
	if err := s.Security.SetUserStatus(ctx, userID, status); err != nil {
		return err
	}
	return s.audit(ctx, actor, "rbac.user.status_changed", "user", userID, map[string]any{"status": status})
}

func (s RBACService) ResetPassword(ctx context.Context, actor entity.AuthIdentity, userID, password string) error {
	if err := validatePassword(password); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.Security.UpdateUserPassword(ctx, userID, string(hash)); err != nil {
		return err
	}
	if err := s.Security.RevokeUserSessions(ctx, userID, time.Now().UTC()); err != nil {
		return err
	}
	return s.audit(ctx, actor, "rbac.user.password_reset", "user", userID, nil)
}

func (s RBACService) ReplaceUserRoles(ctx context.Context, actor entity.AuthIdentity, userID string, roleIDs []string) error {
	if len(roleIDs) == 0 {
		return errors.New("at least one role is required")
	}
	if err := s.Security.ReplaceUserRoles(ctx, userID, uniqueStrings(roleIDs)); err != nil {
		return err
	}
	if err := s.Security.RevokeUserSessions(ctx, userID, time.Now().UTC()); err != nil {
		return err
	}
	return s.audit(ctx, actor, "rbac.user.roles_replaced", "user", userID, map[string]any{"roles": roleIDs})
}

func (s RBACService) ListRoles(ctx context.Context) ([]entity.RoleDetail, error) {
	return s.Security.ListRoles(ctx)
}

func (s RBACService) CreateRole(ctx context.Context, actor entity.AuthIdentity, name, description string, permissionIDs []string) (entity.Role, error) {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 80 {
		return entity.Role{}, errors.New("role name length must be between 2 and 80")
	}
	role := entity.Role{ID: id.New(), Name: name, Description: strings.TrimSpace(description), BuiltIn: false}
	if err := s.Security.CreateRole(ctx, role); err != nil {
		return entity.Role{}, err
	}
	if len(permissionIDs) > 0 {
		if err := s.Security.ReplaceRolePermissions(ctx, role.ID, uniqueStrings(permissionIDs)); err != nil {
			return entity.Role{}, err
		}
	}
	_ = s.audit(ctx, actor, "rbac.role.created", "role", role.ID, map[string]any{"permissions": permissionIDs})
	return role, nil
}

func (s RBACService) UpdateRole(ctx context.Context, actor entity.AuthIdentity, roleID, name, description string) error {
	current, err := s.Security.FindRole(ctx, roleID)
	if err != nil {
		return err
	}
	if current.BuiltIn {
		return errors.New("built-in role cannot be renamed")
	}
	current.Name = strings.TrimSpace(name)
	current.Description = strings.TrimSpace(description)
	if current.Name == "" {
		return errors.New("role name is required")
	}
	if err := s.Security.UpdateRole(ctx, current.Role); err != nil {
		return err
	}
	return s.audit(ctx, actor, "rbac.role.updated", "role", roleID, nil)
}

func (s RBACService) DeleteRole(ctx context.Context, actor entity.AuthIdentity, roleID string) error {
	if err := s.Security.DeleteRole(ctx, roleID); err != nil {
		return err
	}
	return s.audit(ctx, actor, "rbac.role.deleted", "role", roleID, nil)
}

func (s RBACService) ReplaceRolePermissions(ctx context.Context, actor entity.AuthIdentity, roleID string, permissionIDs []string) error {
	role, err := s.Security.FindRole(ctx, roleID)
	if err != nil {
		return err
	}
	if role.BuiltIn && role.Name == "business_admin" {
		allPermissions, err := s.Security.ListPermissions(ctx)
		if err != nil {
			return err
		}
		permissionIDs = make([]string, 0, len(allPermissions))
		for _, permission := range allPermissions {
			permissionIDs = append(permissionIDs, permission.ID)
		}
	}
	if err := s.Security.ReplaceRolePermissions(ctx, roleID, uniqueStrings(permissionIDs)); err != nil {
		return err
	}
	return s.audit(ctx, actor, "rbac.role.permissions_replaced", "role", roleID, map[string]any{"permissions": permissionIDs})
}

func (s RBACService) ListPermissionGroups(ctx context.Context) ([]entity.PermissionGroup, error) {
	permissions, err := s.Security.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	groups := make(map[string][]entity.Permission)
	modules := make([]string, 0)
	for _, permission := range permissions {
		if _, exists := groups[permission.Module]; !exists {
			modules = append(modules, permission.Module)
		}
		groups[permission.Module] = append(groups[permission.Module], permission)
	}
	sort.Strings(modules)
	result := make([]entity.PermissionGroup, 0, len(modules))
	for _, module := range modules {
		result = append(result, entity.PermissionGroup{Module: module, Permissions: groups[module]})
	}
	return result, nil
}

func (s RBACService) CreatePermission(ctx context.Context, actor entity.AuthIdentity, code, module, description string) (entity.Permission, error) {
	code = strings.TrimSpace(code)
	module = strings.TrimSpace(module)
	if !strings.Contains(code, ":") {
		return entity.Permission{}, errors.New("permission code must use module:action format")
	}
	if module == "" {
		return entity.Permission{}, errors.New("permission module is required")
	}
	permission := entity.Permission{ID: id.New(), Code: code, Module: module, Description: strings.TrimSpace(description)}
	if err := s.Security.CreatePermission(ctx, permission); err != nil {
		return entity.Permission{}, err
	}
	_ = s.audit(ctx, actor, "rbac.permission.created", "permission", permission.ID, permission)
	return permission, nil
}

func (s RBACService) UpdatePermission(ctx context.Context, actor entity.AuthIdentity, permission entity.Permission) error {
	if permission.ID == "" || permission.Code == "" || permission.Module == "" {
		return errors.New("permission id, code and module are required")
	}
	if err := s.Security.UpdatePermission(ctx, permission); err != nil {
		return err
	}
	return s.audit(ctx, actor, "rbac.permission.updated", "permission", permission.ID, permission)
}

func (s RBACService) audit(ctx context.Context, actor entity.AuthIdentity, action, resource, resourceID string, detail any) error {
	if s.Audits == nil {
		return nil
	}
	return AuditService{Repo: s.Audits}.Write(ctx, actor.UserID, action, resource, resourceID, detail)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
