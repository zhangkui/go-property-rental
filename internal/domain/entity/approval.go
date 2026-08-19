package entity

import "time"

type ApprovalRequest struct {
	ID, RequestType, ResourceType, ResourceID string
	Title, Summary, Status                    string
	ApplicantID, CurrentReviewerID            string
	CreatedAt, UpdatedAt                      time.Time
	CompletedAt                               *time.Time
}

type ApprovalStep struct {
	ID, ApprovalID, ReviewerID, Decision string
	Sequence                             int
	Comment                              string
	DecidedAt                            *time.Time
}

type ApprovalHistory struct {
	ID, ApprovalID, FromStatus, ToStatus string
	ActorID, Comment                     string
	CreatedAt                            time.Time
}

func (a ApprovalRequest) CanReview(userID string) bool {
	return a.Status == "pending" && (a.CurrentReviewerID == "" || a.CurrentReviewerID == userID)
}
