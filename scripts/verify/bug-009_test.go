package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
	"time"
)

type bug009NotificationStore struct{ userID, notificationID string }

func (*bug009NotificationStore) List(context.Context, string, int, int, string, string) ([]entity.Notification, error) {
	return nil, nil
}
func (*bug009NotificationStore) UnreadCount(context.Context, string) (int, error)        { return 0, nil }
func (*bug009NotificationStore) CreateMany(context.Context, []entity.Notification) error { return nil }
func (s *bug009NotificationStore) MarkRead(_ context.Context, userID, notificationID string) error {
	s.userID = userID
	s.notificationID = notificationID
	return nil
}
func (*bug009NotificationStore) MarkAllRead(context.Context, string) error { return nil }
func (*bug009NotificationStore) ListRules(context.Context) ([]entity.ReminderRule, error) {
	return nil, nil
}
func (*bug009NotificationStore) SaveRule(context.Context, entity.ReminderRule) error { return nil }
func (*bug009NotificationStore) SetRuleEnabled(context.Context, string, bool) error  { return nil }
func (*bug009NotificationStore) StartRun(context.Context, entity.ReminderRun) (bool, error) {
	return false, nil
}
func (*bug009NotificationStore) FinishRun(context.Context, string, string, int, int, time.Time) error {
	return nil
}
func (*bug009NotificationStore) RecipientIDsByRole(context.Context, string) ([]string, error) {
	return nil, nil
}
func (*bug009NotificationStore) LeaseExpiryCandidates(context.Context, int) ([]entity.Lease, error) {
	return nil, nil
}
func (*bug009NotificationStore) OverdueBillCandidates(context.Context, int) ([]entity.Bill, error) {
	return nil, nil
}
func (*bug009NotificationStore) OverdueWorkOrderCandidates(context.Context, int) ([]entity.WorkOrder, error) {
	return nil, nil
}
func TestBug009_BusinessRegression(t *testing.T) {
	s := &bug009NotificationStore{}
	svc := service.NotificationService{Repo: s}
	if err := svc.MarkRead(context.Background(), "owner-1", "notification-1"); err != nil {
		t.Fatal(err)
	}
	if s.userID != "owner-1" || s.notificationID != "notification-1" {
		t.Fatalf("ownership arguments reversed: user=%q notification=%q", s.userID, s.notificationID)
	}
	if err := svc.MarkRead(context.Background(), "owner-1", ""); err == nil {
		t.Fatal("empty notification id must be rejected")
	}
}
