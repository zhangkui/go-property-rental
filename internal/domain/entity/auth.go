package entity

import "time"

type AuthSession struct {
	ID               string
	UserID           string
	AccessTokenHash  string
	RefreshTokenHash string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
	LastSeenAt       time.Time
	UserAgent        string
	IPAddress        string
}

type AuthIdentity struct {
	UserID      string
	Username    string
	DisplayName string
	Status      string
	SessionID   string
	Permissions []string
}

func (i AuthIdentity) HasPermission(code string) bool {
	for _, permission := range i.Permissions {
		if permission == code || permission == "*" {
			return true
		}
	}
	return false
}

type UserDetail struct {
	ID          string
	Username    string
	DisplayName string
	Email       string
	Status      string
	Roles       []Role
	Permissions []Permission
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastLoginAt *time.Time
}

type RoleDetail struct {
	Role
	Permissions []Permission
	UserCount   int
}

type PermissionGroup struct {
	Module      string
	Permissions []Permission
}

type SessionTokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}
