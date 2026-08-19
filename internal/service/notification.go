package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
	"go-property-rental/internal/repository"
)

type NotificationService struct{ Repo repository.NotificationStore }

func (s NotificationService) List(ctx context.Context, userID string, page, size int, status, category string) ([]entity.Notification, error) {
	limit, offset := pageValues(page, size)
	return s.Repo.List(ctx, userID, limit, offset, status, category)
}
func (s NotificationService) UnreadCount(ctx context.Context, userID string) (int, error) {
	return s.Repo.UnreadCount(ctx, userID)
}
func (s NotificationService) MarkRead(ctx context.Context, userID, notificationID string) error {
	if notificationID == "" {
		return errors.New("notification id is required")
	}
	return s.Repo.MarkRead(ctx, notificationID, userID)
}
func (s NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	return s.Repo.MarkAllRead(ctx, userID)
}
func (s NotificationService) Rules(ctx context.Context) ([]entity.ReminderRule, error) {
	return s.Repo.ListRules(ctx)
}
func (s NotificationService) SaveRule(ctx context.Context, rule entity.ReminderRule) error {
	if rule.ID == "" {
		rule.ID = id.New()
	}
	if rule.Name == "" || rule.Category == "" || rule.RecipientRole == "" {
		return errors.New("rule name, category and recipient role are required")
	}
	return s.Repo.SaveRule(ctx, rule)
}
func (s NotificationService) EnableRule(ctx context.Context, ruleID string, enabled bool) error {
	return s.Repo.SetRuleEnabled(ctx, ruleID, enabled)
}

func (s NotificationService) RunReminders(ctx context.Context, now time.Time) error {
	rules, err := s.Repo.ListRules(ctx)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		if err := s.runRule(ctx, rule, now); err != nil {
			return err
		}
	}
	return nil
}

func (s NotificationService) runRule(ctx context.Context, rule entity.ReminderRule, now time.Time) error {
	run := entity.ReminderRun{ID: id.New(), RuleID: rule.ID, RunKey: now.UTC().Format("2006-01-02"), StartedAt: now.UTC()}
	started, err := s.Repo.StartRun(ctx, run)
	if err != nil || !started {
		return err
	}
	recipients, err := s.Repo.RecipientIDsByRole(ctx, rule.RecipientRole)
	if err != nil {
		_ = s.Repo.FinishRun(ctx, run.ID, "failed", 0, 1, time.Now().UTC())
		return err
	}
	notifications := make([]entity.Notification, 0)
	switch rule.Category {
	case "lease_expiry":
		items, e := s.Repo.LeaseExpiryCandidates(ctx, rule.DaysBefore)
		if e != nil {
			err = e
			break
		}
		for _, item := range items {
			for _, recipient := range recipients {
				notifications = append(notifications, entity.Notification{ID: id.New(), UserID: recipient, Category: rule.Category, Title: "租约即将到期", Content: fmt.Sprintf("租约 %s 将于 %s 到期，请及时跟进续租或退租。", item.ID, item.EndDate.Format("2006-01-02")), ResourceType: "lease", ResourceID: item.ID, Status: "unread", CreatedAt: now})
			}
		}
	case "overdue_bill":
		items, e := s.Repo.OverdueBillCandidates(ctx, rule.DaysAfter)
		if e != nil {
			err = e
			break
		}
		for _, item := range items {
			for _, recipient := range recipients {
				notifications = append(notifications, entity.Notification{ID: id.New(), UserID: recipient, Category: rule.Category, Title: "账单逾期提醒", Content: fmt.Sprintf("账单 %s 已超过到期日，请核查收款和滞纳金。", item.ID), ResourceType: "bill", ResourceID: item.ID, Status: "unread", CreatedAt: now})
			}
		}
	case "work_order_overdue":
		items, e := s.Repo.OverdueWorkOrderCandidates(ctx, rule.DaysAfter)
		if e != nil {
			err = e
			break
		}
		for _, item := range items {
			for _, recipient := range recipients {
				notifications = append(notifications, entity.Notification{ID: id.New(), UserID: recipient, Category: rule.Category, Title: "维修工单超时提醒", Content: fmt.Sprintf("工单 %s 已处理超过 %d 天，请检查进度。", item.ID, rule.DaysAfter), ResourceType: "work_order", ResourceID: item.ID, Status: "unread", CreatedAt: now})
			}
		}
	default:
		err = errors.New("unsupported reminder category")
	}
	if err == nil {
		err = s.Repo.CreateMany(ctx, notifications)
	}
	status := "completed"
	failed := 0
	if err != nil {
		status = "failed"
		failed = 1
	}
	_ = s.Repo.FinishRun(ctx, run.ID, status, len(notifications), failed, time.Now().UTC())
	return err
}
