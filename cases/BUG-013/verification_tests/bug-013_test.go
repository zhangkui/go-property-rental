package verify

import (
	"context"
	"errors"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
)

type bug013ApprovalStore struct {
	owner   string
	cancels int
}

func (*bug013ApprovalStore) Create(context.Context, entity.ApprovalRequest, []entity.ApprovalStep) error {
	return nil
}
func (*bug013ApprovalStore) List(context.Context, int, int, string, string, string) ([]entity.ApprovalRequest, error) {
	return nil, nil
}
func (*bug013ApprovalStore) Get(context.Context, string) (entity.ApprovalRequest, []entity.ApprovalStep, []entity.ApprovalHistory, error) {
	return entity.ApprovalRequest{}, nil, nil, nil
}
func (*bug013ApprovalStore) Decide(context.Context, string, string, string, string) error { return nil }
func (s *bug013ApprovalStore) Cancel(_ context.Context, _ string, applicant, _ string) error {
	if applicant != s.owner {
		return errors.New("not applicant")
	}
	s.cancels++
	return nil
}

type bug013AuditStore struct{ entries []entity.AuditLog }

func (s *bug013AuditStore) Append(_ context.Context, x entity.AuditLog) error {
	s.entries = append(s.entries, x)
	return nil
}
func (*bug013AuditStore) List(context.Context, int, int, string, string) ([]entity.AuditLog, error) {
	return nil, nil
}
func TestBug013_BusinessRegression(t *testing.T) {
	repo := &bug013ApprovalStore{owner: "applicant-1"}
	audits := &bug013AuditStore{}
	svc := service.ApprovalService{Repo: repo, Audits: audits}
	if err := svc.Cancel(context.Background(), "approval-1", "applicant-1", "cancel"); err != nil {
		t.Fatal(err)
	}
	if len(audits.entries) != 1 || audits.entries[0].ActorID != "applicant-1" {
		t.Fatalf("wrong cancellation audit actor: %#v", audits.entries)
	}
	if err := svc.Cancel(context.Background(), "approval-1", "other", "cancel"); err == nil {
		t.Fatal("non-applicant cancellation must fail")
	}
	if len(audits.entries) != 1 || repo.cancels != 1 {
		t.Fatal("rejected cancellation created side effects")
	}
}
