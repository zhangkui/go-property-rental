package repository

import (
	"context"
	"time"

	"go-property-rental/internal/domain/entity"
)

type SecurityStore interface {
	FindUserByUsername(context.Context, string) (entity.User, error)
	FindUserByID(context.Context, string) (entity.User, error)
	ListUsers(context.Context, int, int, string, string) ([]entity.UserDetail, int, error)
	CreateUser(context.Context, entity.User, []string) error
	UpdateUser(context.Context, entity.User) error
	SetUserStatus(context.Context, string, string) error
	UpdateUserPassword(context.Context, string, string) error
	MarkUserLogin(context.Context, string, time.Time) error
	UserRoles(context.Context, string) ([]entity.Role, error)
	UserPermissions(context.Context, string) ([]entity.Permission, error)
	ReplaceUserRoles(context.Context, string, []string) error

	ListRoles(context.Context) ([]entity.RoleDetail, error)
	FindRole(context.Context, string) (entity.RoleDetail, error)
	FindRoleByName(context.Context, string) (entity.Role, error)
	CreateRole(context.Context, entity.Role) error
	UpdateRole(context.Context, entity.Role) error
	DeleteRole(context.Context, string) error
	ReplaceRolePermissions(context.Context, string, []string) error

	ListPermissions(context.Context) ([]entity.Permission, error)
	CreatePermission(context.Context, entity.Permission) error
	UpdatePermission(context.Context, entity.Permission) error

	CreateSession(context.Context, entity.AuthSession) error
	FindSessionByAccessHash(context.Context, string) (entity.AuthSession, error)
	FindSessionByRefreshHash(context.Context, string) (entity.AuthSession, error)
	RotateSession(context.Context, entity.AuthSession) error
	TouchSession(context.Context, string, time.Time) error
	RevokeSession(context.Context, string, time.Time) error
	RevokeUserSessions(context.Context, string, time.Time) error
}

type LoginLimiter interface {
	Allow(context.Context, string, int, time.Duration) (bool, int64, error)
	Reset(context.Context, string) error
}
