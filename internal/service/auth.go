package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
	"go-property-rental/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type LoginMetadata struct {
	IPAddress string
	UserAgent string
}

type AuthService struct {
	Security      repository.SecurityStore
	Limiter       repository.LoginLimiter
	Audits        repository.AuditStore
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	LoginLimit    int
	LoginWindow   time.Duration
	TouchInterval time.Duration
}

func NewAuthService(security repository.SecurityStore, limiter repository.LoginLimiter, audits repository.AuditStore) AuthService {
	return AuthService{
		Security:      security,
		Limiter:       limiter,
		Audits:        audits,
		AccessTTL:     30 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
		LoginLimit:    5,
		LoginWindow:   15 * time.Minute,
		TouchInterval: 5 * time.Minute,
	}
}

func (s AuthService) Login(ctx context.Context, username, password string, metadata LoginMetadata) (entity.User, entity.SessionTokenPair, error) {
	username = strings.TrimSpace(username)
	limiterKey := strings.ToLower(username) + ":" + metadata.IPAddress
	if s.Limiter != nil {
		allowed, count, err := s.Limiter.Allow(ctx, limiterKey, s.LoginLimit, s.LoginWindow)
		if err == nil && !allowed {
			_ = s.writeAudit(ctx, "", "auth.login.rate_limited", "user", "", map[string]any{"username": username, "ip": metadata.IPAddress, "attempts": count})
			return entity.User{}, entity.SessionTokenPair{}, errors.New("too many login attempts")
		}
	}
	user, err := s.Security.FindUserByUsername(ctx, username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		_ = s.writeAudit(ctx, "", "auth.login.failed", "user", "", map[string]any{"username": username, "ip": metadata.IPAddress})
		return entity.User{}, entity.SessionTokenPair{}, errors.New("invalid credentials")
	}
	if user.Status != "active" {
		_ = s.writeAudit(ctx, user.ID, "auth.login.disabled", "user", user.ID, map[string]any{"ip": metadata.IPAddress})
		return entity.User{}, entity.SessionTokenPair{}, errors.New("user is disabled")
	}
	now := time.Now().UTC()
	pair, session, err := s.newSession(user.ID, metadata, now)
	if err != nil {
		return entity.User{}, entity.SessionTokenPair{}, err
	}
	if err := s.Security.CreateSession(ctx, session); err != nil {
		return entity.User{}, entity.SessionTokenPair{}, err
	}
	_ = s.Security.MarkUserLogin(ctx, user.ID, now)
	if s.Limiter != nil {
		_ = s.Limiter.Reset(ctx, limiterKey)
	}
	_ = s.writeAudit(ctx, user.ID, "auth.login.succeeded", "session", session.ID, map[string]any{"ip": metadata.IPAddress, "user_agent": metadata.UserAgent})
	return user, pair, nil
}

func (s AuthService) Register(ctx context.Context, username, password, displayName, email string) (entity.User, error) {
	username = strings.TrimSpace(username)
	if err := validateUsername(username); err != nil {
		return entity.User{}, err
	}
	if err := validatePassword(password); err != nil {
		return entity.User{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return entity.User{}, err
	}
	roleIDs := make([]string, 0, 1)
	if role, roleErr := s.Security.FindRoleByName(ctx, "readonly_auditor"); roleErr == nil {
		roleIDs = append(roleIDs, role.ID)
	}
	user := entity.User{
		ID:           id.New(),
		Username:     username,
		PasswordHash: string(hash),
		DisplayName:  strings.TrimSpace(displayName),
		Email:        strings.TrimSpace(email),
		Status:       "active",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}
	if err := s.Security.CreateUser(ctx, user, roleIDs); err != nil {
		return entity.User{}, err
	}
	_ = s.writeAudit(ctx, user.ID, "auth.register", "user", user.ID, map[string]any{"username": user.Username})
	return user, nil
}

func (s AuthService) Authenticate(ctx context.Context, accessToken string) (entity.AuthIdentity, error) {
	if strings.TrimSpace(accessToken) == "" {
		return entity.AuthIdentity{}, errors.New("missing access token")
	}
	session, err := s.Security.FindSessionByAccessHash(ctx, HashToken(accessToken))
	if err != nil {
		return entity.AuthIdentity{}, errors.New("invalid access token")
	}
	now := time.Now().UTC()
	if session.RevokedAt != nil || !now.Before(session.AccessExpiresAt) {
		return entity.AuthIdentity{}, errors.New("access token expired or revoked")
	}
	user, err := s.Security.FindUserByID(ctx, session.UserID)
	if err != nil || user.Status != "active" {
		return entity.AuthIdentity{}, errors.New("user is unavailable")
	}
	permissions, err := s.Security.UserPermissions(ctx, user.ID)
	if err != nil {
		return entity.AuthIdentity{}, err
	}
	codes := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		codes = append(codes, permission.Code)
	}
	if now.Sub(session.LastSeenAt) >= s.TouchInterval {
		_ = s.Security.TouchSession(ctx, session.ID, now)
	}
	return entity.AuthIdentity{UserID: user.ID, Username: user.Username, DisplayName: user.DisplayName, Status: user.Status, SessionID: session.ID, Permissions: codes}, nil
}

func (s AuthService) Refresh(ctx context.Context, refreshToken string) (entity.SessionTokenPair, error) {
	session, err := s.Security.FindSessionByRefreshHash(ctx, HashToken(refreshToken))
	if err != nil {
		return entity.SessionTokenPair{}, errors.New("invalid refresh token")
	}
	now := time.Now().UTC()
	if session.RevokedAt != nil || !now.Before(session.RefreshExpiresAt) {
		return entity.SessionTokenPair{}, errors.New("refresh token expired or revoked")
	}
	user, err := s.Security.FindUserByID(ctx, session.UserID)
	if err != nil || user.Status != "active" {
		return entity.SessionTokenPair{}, errors.New("user is unavailable")
	}
	accessToken, err := randomToken(48)
	if err != nil {
		return entity.SessionTokenPair{}, err
	}
	nextRefreshToken, err := randomToken(64)
	if err != nil {
		return entity.SessionTokenPair{}, err
	}
	session.AccessTokenHash = HashToken(accessToken)
	session.RefreshTokenHash = HashToken(nextRefreshToken)
	session.AccessExpiresAt = now.Add(s.accessTTL())
	session.RefreshExpiresAt = now.Add(s.refreshTTL())
	session.LastSeenAt = now
	if err := s.Security.RotateSession(ctx, session); err != nil {
		return entity.SessionTokenPair{}, err
	}
	_ = s.writeAudit(ctx, user.ID, "auth.session.refreshed", "session", session.ID, nil)
	return entity.SessionTokenPair{AccessToken: accessToken, RefreshToken: nextRefreshToken, AccessExpiresAt: session.AccessExpiresAt, RefreshExpiresAt: session.RefreshExpiresAt}, nil
}

func (s AuthService) Logout(ctx context.Context, identity entity.AuthIdentity) error {
	now := time.Now().UTC()
	if err := s.Security.RevokeSession(ctx, identity.SessionID, now); err != nil {
		return err
	}
	return s.writeAudit(ctx, identity.UserID, "auth.logout", "session", identity.SessionID, nil)
}

func (s AuthService) CurrentUser(ctx context.Context, identity entity.AuthIdentity) (entity.UserDetail, error) {
	user, err := s.Security.FindUserByID(ctx, identity.UserID)
	if err != nil {
		return entity.UserDetail{}, err
	}
	roles, err := s.Security.UserRoles(ctx, user.ID)
	if err != nil {
		return entity.UserDetail{}, err
	}
	permissions, err := s.Security.UserPermissions(ctx, user.ID)
	if err != nil {
		return entity.UserDetail{}, err
	}
	return entity.UserDetail{ID: user.ID, Username: user.Username, DisplayName: user.DisplayName, Email: user.Email, Status: user.Status, Roles: roles, Permissions: permissions, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, LastLoginAt: user.LastLoginAt}, nil
}

func (s AuthService) ChangePassword(ctx context.Context, identity entity.AuthIdentity, currentPassword, nextPassword string) error {
	user, err := s.Security.FindUserByID(ctx, identity.UserID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)) != nil {
		return errors.New("current password is incorrect")
	}
	if err := validatePassword(nextPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(nextPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.Security.UpdateUserPassword(ctx, user.ID, string(hash)); err != nil {
		return err
	}
	now := time.Now().UTC()
	if err := s.Security.RevokeUserSessions(ctx, user.ID, now); err != nil {
		return err
	}
	return s.writeAudit(ctx, user.ID, "auth.password.changed", "user", user.ID, nil)
}

func (s AuthService) newSession(userID string, metadata LoginMetadata, now time.Time) (entity.SessionTokenPair, entity.AuthSession, error) {
	accessToken, err := randomToken(48)
	if err != nil {
		return entity.SessionTokenPair{}, entity.AuthSession{}, err
	}
	refreshToken, err := randomToken(64)
	if err != nil {
		return entity.SessionTokenPair{}, entity.AuthSession{}, err
	}
	session := entity.AuthSession{
		ID:               id.New(),
		UserID:           userID,
		AccessTokenHash:  HashToken(accessToken),
		RefreshTokenHash: HashToken(refreshToken),
		AccessExpiresAt:  now.Add(s.accessTTL()),
		RefreshExpiresAt: now.Add(s.refreshTTL()),
		CreatedAt:        now,
		LastSeenAt:       now,
		UserAgent:        truncate(metadata.UserAgent, 255),
		IPAddress:        truncate(metadata.IPAddress, 64),
	}
	return entity.SessionTokenPair{AccessToken: accessToken, RefreshToken: refreshToken, AccessExpiresAt: session.AccessExpiresAt, RefreshExpiresAt: session.RefreshExpiresAt}, session, nil
}

func (s AuthService) accessTTL() time.Duration {
	if s.AccessTTL <= 0 {
		return 30 * time.Minute
	}
	return s.AccessTTL
}

func (s AuthService) refreshTTL() time.Duration {
	if s.RefreshTTL <= 0 {
		return 7 * 24 * time.Hour
	}
	return s.RefreshTTL
}

func (s AuthService) writeAudit(ctx context.Context, actorID, action, resource, resourceID string, detail any) error {
	if s.Audits == nil {
		return nil
	}
	return AuditService{Repo: s.Audits}.Write(ctx, actorID, action, resource, resourceID, detail)
}

func HashToken(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func randomToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 64 {
		return errors.New("username length must be between 3 and 64")
	}
	for _, character := range username {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '_' || character == '-' || character == '.' {
			continue
		}
		return errors.New("username contains unsupported characters")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 10 || len(password) > 128 {
		return errors.New("password length must be between 10 and 128")
	}
	var upper, lower, digit, special bool
	for _, character := range password {
		switch {
		case character >= 'A' && character <= 'Z':
			upper = true
		case character >= 'a' && character <= 'z':
			lower = true
		case character >= '0' && character <= '9':
			digit = true
		default:
			special = true
		}
	}
	if !upper || !lower || !digit || !special {
		return fmt.Errorf("password must contain upper, lower, digit and special characters")
	}
	return nil
}

func truncate(value string, size int) string {
	if len(value) <= size {
		return value
	}
	return value[:size]
}
