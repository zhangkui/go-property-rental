package entity

import "testing"

func TestApprovalRequestCanReview(t *testing.T) {
	request := ApprovalRequest{Status: "pending", CurrentReviewerID: "reviewer-1"}
	if !request.CanReview("reviewer-1") {
		t.Fatal("assigned reviewer should be allowed")
	}
	if request.CanReview("reviewer-2") {
		t.Fatal("different reviewer should be rejected")
	}
	request.CurrentReviewerID = ""
	if !request.CanReview("reviewer-2") {
		t.Fatal("unassigned pending approval should be reviewable")
	}
	request.Status = "approved"
	if request.CanReview("reviewer-2") {
		t.Fatal("completed approval should not be reviewable")
	}
}
