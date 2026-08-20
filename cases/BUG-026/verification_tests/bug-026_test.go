package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
	"time"
)

type bug026Store struct {
	keys    []string
	started map[string]bool
	created int
}

func (*bug026Store) List(context.Context, string, int, int, string, string) ([]entity.Notification, error) {
	return nil, nil
}
func (*bug026Store) UnreadCount(context.Context, string) (int, error) { return 0, nil }
func (s *bug026Store) CreateMany(_ context.Context, x []entity.Notification) error {
	s.created += len(x)
	return nil
}
func (*bug026Store) MarkRead(context.Context, string, string) error { return nil }
func (*bug026Store) MarkAllRead(context.Context, string) error      { return nil }
func (*bug026Store) ListRules(context.Context) ([]entity.ReminderRule, error) {
	return []entity.ReminderRule{{ID: "rule-1", Enabled: true, Category: "lease_expiry", RecipientRole: "ops", DaysBefore: 30}}, nil
}
func (*bug026Store) SaveRule(context.Context, entity.ReminderRule) error { return nil }
func (*bug026Store) SetRuleEnabled(context.Context, string, bool) error  { return nil }
func (s *bug026Store) StartRun(_ context.Context, r entity.ReminderRun) (bool, error) {
	s.keys = append(s.keys, r.RunKey)
	if s.started == nil {
		s.started = map[string]bool{}
	}
	if s.started[r.RunKey] {
		return false, nil
	}
	s.started[r.RunKey] = true
	return true, nil
}
func (*bug026Store) FinishRun(context.Context, string, string, int, int, time.Time) error { return nil }
func (*bug026Store) RecipientIDsByRole(context.Context, string) ([]string, error) {
	return []string{"user-1"}, nil
}
func (*bug026Store) LeaseExpiryCandidates(context.Context, int) ([]entity.Lease, error) {
	return []entity.Lease{{ID: "lease-1", EndDate: time.Now()}}, nil
}
func (*bug026Store) OverdueBillCandidates(context.Context, int) ([]entity.Bill, error) {
	return nil, nil
}
func (*bug026Store) OverdueWorkOrderCandidates(context.Context, int) ([]entity.WorkOrder, error) {
	return nil, nil
}
func TestBug026_BusinessRegression(t *testing.T) {
	s := &bug026Store{}
	svc := service.NotificationService{Repo: s}
	now := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	if err := svc.RunReminders(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	if err := svc.RunReminders(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	if len(s.keys) != 2 || s.keys[0] != "2026-08-19" || s.keys[1] != "2026-08-19" || s.created != 1 {
		t.Fatalf("daily idempotency failed: keys=%#v created=%d", s.keys, s.created)
	}
}
