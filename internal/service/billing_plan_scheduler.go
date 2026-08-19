package service

import (
	"context"
	"log"
	"time"
)

type BillingPlanScheduler struct {
	Plans    BillingPlanService
	Interval time.Duration
}

func (s BillingPlanScheduler) Start(ctx context.Context) {
	interval := s.Interval
	if interval <= 0 {
		interval = 12 * time.Hour
	}
	go func() {
		timer := time.NewTimer(2 * time.Minute)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			s.run(ctx)
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.run(ctx)
			}
		}
	}()
}
func (s BillingPlanScheduler) run(ctx context.Context) {
	runCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	if _, err := s.Plans.RunDue(runCtx, "scheduled", "", time.Now().UTC()); err != nil && err.Error() != "billing run is already in progress" && ctx.Err() == nil {
		log.Printf("billing plan scheduler failed: %v", err)
	}
}
