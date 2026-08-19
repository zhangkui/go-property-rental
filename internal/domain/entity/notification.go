package entity

import "time"

type ReminderRule struct {
	ID, Name, Category, RecipientRole string
	DaysBefore, DaysAfter             int
	Enabled                           bool
	CreatedAt, UpdatedAt              time.Time
}

type ReminderRun struct {
	ID, RuleID, RunKey, Status string
	Generated, Failed          int
	StartedAt, FinishedAt      time.Time
}
