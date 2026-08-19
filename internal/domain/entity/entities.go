package entity

import "time"

type User struct {
	ID           string
	Username     string
	PasswordHash string
	DisplayName  string
	Email        string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
}

type Role struct {
	ID          string
	Name        string
	Description string
	BuiltIn     bool
}

type Permission struct {
	ID          string
	Code        string
	Module      string
	Description string
}

type UserRole struct{ UserID, RoleID string }
type RolePermission struct{ RoleID, PermissionID string }
type RefreshToken struct {
	ID, UserID, TokenHash string
	ExpiresAt             time.Time
	Revoked               bool
}
type Property struct {
	ID, Building, Room, Status string
	Facilities                 []string
	AvailableFrom              time.Time
}
type PropertyStateHistory struct {
	ID, PropertyID, FromStatus, ToStatus, Reason, ActorID string
	CreatedAt                                             time.Time
}
type Notification struct {
	ID, UserID, Category, Title, Content string
	ResourceType, ResourceID             string
	Status                               string
	CreatedAt                            time.Time
	ReadAt                               *time.Time
}
