package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
)

type bug029ApprovalStore struct{ approval, reviewer, decision string }

func (*bug029ApprovalStore) Create(context.Context, entity.ApprovalRequest, []entity.ApprovalStep) error {
	return nil
}
func (*bug029ApprovalStore) List(context.Context, int, int, string, string, string) ([]entity.ApprovalRequest, error) {
	return nil, nil
}
func (*bug029ApprovalStore) Get(context.Context, string) (entity.ApprovalRequest, []entity.ApprovalStep, []entity.ApprovalHistory, error) {
	return entity.ApprovalRequest{}, nil, nil, nil
}
func (s *bug029ApprovalStore) Decide(_ context.Context, a, r, d, _ string) error {
	s.approval = a
	s.reviewer = r
	s.decision = d
	return nil
}
func (*bug029ApprovalStore) Cancel(context.Context, string, string, string) error { return nil }

type bug029AuditStore struct{}

func (*bug029AuditStore) Append(context.Context, entity.AuditLog) error { return nil }
func (*bug029AuditStore) List(context.Context, int, int, string, string) ([]entity.AuditLog, error) {
	return nil, nil
}
func TestBug029_BusinessRegression(t *testing.T) {
	s := &bug029ApprovalStore{}
	svc := service.ApprovalService{Repo: s, Audits: &bug029AuditStore{}}
	if err := svc.Decide(context.Background(), "approval-1", "reviewer-1", "approved", "checked"); err != nil {
		t.Fatal(err)
	}
	if s.approval != "approval-1" || s.reviewer != "reviewer-1" || s.decision != "approved" {
		t.Fatalf("reviewer identity corrupted: %#v", s)
	}
}
