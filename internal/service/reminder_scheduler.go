package service

import (
	"context"
	"log"
	"time"
)

type ReminderScheduler struct {
	Notifications NotificationService
	Interval      time.Duration
	Now           func() time.Time
}

func (s ReminderScheduler) Start(ctx context.Context) {
	interval := s.Interval
	if interval <= 0 {
		interval = 6 * time.Hour
	}
	now := s.Now
	if now == nil {
		now = time.Now
	}
	run := func() {
		runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		if err := s.Notifications.RunReminders(runCtx, now().UTC()); err != nil && ctx.Err() == nil {
			log.Printf("reminder scheduler failed: %v", err)
		}
	}
	go func() {
		timer := time.NewTimer(time.Minute)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			run()
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}
