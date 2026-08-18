package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
	"go-property-rental/internal/repository"
)

type ApprovalService struct {
	Repo   repository.ApprovalStore
	Audits repository.AuditStore
}

func (s ApprovalService) Create(ctx context.Context, requestType, resourceType, resourceID, title, summary, applicant string, reviewers []string) (entity.ApprovalRequest, error) {
	if err := required(requestType, resourceType, resourceID, title, applicant); err != nil {
		return entity.ApprovalRequest{}, err
	}
	if len(reviewers) == 0 {
		return entity.ApprovalRequest{}, errors.New("at least one reviewer is required")
	}
	now := time.Now().UTC()
	request := entity.ApprovalRequest{ID: id.New(), RequestType: requestType, ResourceType: resourceType, ResourceID: resourceID, Title: strings.TrimSpace(title), Summary: summary, Status: "pending", ApplicantID: applicant, CurrentReviewerID: reviewers[0], CreatedAt: now, UpdatedAt: now}
	steps := make([]entity.ApprovalStep, 0, len(reviewers))
	for index, reviewer := range reviewers {
		steps = append(steps, entity.ApprovalStep{ID: id.New(), ApprovalID: request.ID, Sequence: index + 1, ReviewerID: reviewer, Decision: "pending"})
	}
	if err := s.Repo.Create(ctx, request, steps); err != nil {
		return entity.ApprovalRequest{}, err
	}
	return request, nil
}

func (s ApprovalService) List(ctx context.Context, page, size int, status, requestType, reviewer string) ([]entity.ApprovalRequest, error) {
	limit, offset := pageValues(page, size)
	return s.Repo.List(ctx, limit, offset, status, requestType, reviewer)
}

func (s ApprovalService) Detail(ctx context.Context, id string) (entity.ApprovalRequest, []entity.ApprovalStep, []entity.ApprovalHistory, error) {
	return s.Repo.Get(ctx, id)
}

func (s ApprovalService) Decide(ctx context.Context, approvalID, reviewerID, decision, comment string) error {
	if decision != "approved" && decision != "rejected" {
		return errors.New("decision must be approved or rejected")
	}
	if err := s.Repo.Decide(ctx, approvalID, reviewerID, decision, comment); err != nil {
		return err
	}
	if s.Audits != nil {
		_ = s.Audits.Append(ctx, entity.AuditLog{ID: id.New(), ActorID: reviewerID, Action: "approval." + decision, Resource: "approval", ResourceID: approvalID, Detail: comment})
	}
	return nil
}

func (s ApprovalService) Cancel(ctx context.Context, approvalID, applicantID, comment string) error {
	if err := s.Repo.Cancel(ctx, approvalID, applicantID, comment); err != nil {
		return err
	}
	if s.Audits != nil {
		_ = s.Audits.Append(ctx, entity.AuditLog{ID: id.New(), ActorID: applicantID, Action: "approval.cancel", Resource: "approval", ResourceID: approvalID, Detail: comment})
	}
	return nil
}
