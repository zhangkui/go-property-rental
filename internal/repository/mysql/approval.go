package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
)

type Approvals struct{ DB *sql.DB }

func (r Approvals) Create(ctx context.Context, request entity.ApprovalRequest, steps []entity.ApprovalStep) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO approval_requests(id,request_type,resource_type,resource_id,title,summary,status,applicant_id,current_reviewer_id,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, request.ID, request.RequestType, request.ResourceType, request.ResourceID, request.Title, request.Summary, request.Status, request.ApplicantID, nullableString(request.CurrentReviewerID), request.CreatedAt, request.UpdatedAt)
	if err != nil {
		return err
	}
	for _, step := range steps {
		_, err = tx.ExecContext(ctx, `INSERT INTO approval_steps(id,approval_id,sequence_no,reviewer_id,decision,comment) VALUES(?,?,?,?,?,?)`, step.ID, request.ID, step.Sequence, nullableString(step.ReviewerID), step.Decision, step.Comment)
		if err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO approval_state_history(id,approval_id,from_status,to_status,actor_id,comment,created_at) VALUES(?,?,?,?,?,?,?)`, id.New(), request.ID, "", request.Status, request.ApplicantID, "提交审批", request.CreatedAt)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r Approvals) List(ctx context.Context, limit, offset int, status, requestType, reviewerID string) ([]entity.ApprovalRequest, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,request_type,resource_type,resource_id,title,summary,status,applicant_id,COALESCE(current_reviewer_id,''),created_at,updated_at,completed_at FROM approval_requests WHERE (?='' OR status=?) AND (?='' OR request_type=?) AND (?='' OR current_reviewer_id=? OR applicant_id=?) ORDER BY created_at DESC LIMIT ? OFFSET ?`, status, status, requestType, requestType, reviewerID, reviewerID, reviewerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entity.ApprovalRequest, 0)
	for rows.Next() {
		request, err := scanApproval(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, request)
	}
	return items, rows.Err()
}

func (r Approvals) Get(ctx context.Context, approvalID string) (entity.ApprovalRequest, []entity.ApprovalStep, []entity.ApprovalHistory, error) {
	request, err := scanApproval(r.DB.QueryRowContext(ctx, `SELECT id,request_type,resource_type,resource_id,title,summary,status,applicant_id,COALESCE(current_reviewer_id,''),created_at,updated_at,completed_at FROM approval_requests WHERE id=?`, approvalID))
	if err != nil {
		return request, nil, nil, err
	}
	steps, err := r.listSteps(ctx, approvalID)
	if err != nil {
		return request, nil, nil, err
	}
	history, err := r.listHistory(ctx, approvalID)
	return request, steps, history, err
}

func (r Approvals) listSteps(ctx context.Context, approvalID string) ([]entity.ApprovalStep, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,approval_id,sequence_no,COALESCE(reviewer_id,''),decision,comment,decided_at FROM approval_steps WHERE approval_id=? ORDER BY sequence_no`, approvalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	steps := make([]entity.ApprovalStep, 0)
	for rows.Next() {
		var step entity.ApprovalStep
		if err := rows.Scan(&step.ID, &step.ApprovalID, &step.Sequence, &step.ReviewerID, &step.Decision, &step.Comment, &step.DecidedAt); err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, rows.Err()
}

func (r Approvals) listHistory(ctx context.Context, approvalID string) ([]entity.ApprovalHistory, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,approval_id,from_status,to_status,actor_id,comment,created_at FROM approval_state_history WHERE approval_id=? ORDER BY created_at,id`, approvalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	history := make([]entity.ApprovalHistory, 0)
	for rows.Next() {
		var item entity.ApprovalHistory
		if err := rows.Scan(&item.ID, &item.ApprovalID, &item.FromStatus, &item.ToStatus, &item.ActorID, &item.Comment, &item.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, item)
	}
	return history, rows.Err()
}

func scanApproval(scanner interface{ Scan(...any) error }) (entity.ApprovalRequest, error) {
	var request entity.ApprovalRequest
	err := scanner.Scan(&request.ID, &request.RequestType, &request.ResourceType, &request.ResourceID, &request.Title, &request.Summary, &request.Status, &request.ApplicantID, &request.CurrentReviewerID, &request.CreatedAt, &request.UpdatedAt, &request.CompletedAt)
	return request, err
}

func (r Approvals) Decide(ctx context.Context, approvalID, reviewerID, decision, comment string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	status, stepID, err := lockPendingApproval(ctx, tx, approvalID, reviewerID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if _, err = tx.ExecContext(ctx, `UPDATE approval_steps SET reviewer_id=COALESCE(reviewer_id,?),decision=?,comment=?,decided_at=? WHERE id=?`, reviewerID, decision, comment, now, stepID); err != nil {
		return err
	}
	nextStatus, nextReviewer, err := nextApprovalState(ctx, tx, approvalID, decision)
	if err != nil {
		return err
	}
	var completedAt any
	if nextStatus != "pending" {
		completedAt = now
	}
	if _, err = tx.ExecContext(ctx, `UPDATE approval_requests SET status=?,current_reviewer_id=?,completed_at=? WHERE id=?`, nextStatus, nextReviewer, completedAt, approvalID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO approval_state_history(id,approval_id,from_status,to_status,actor_id,comment,created_at) VALUES(?,?,?,?,?,?,?)`, id.New(), approvalID, status, nextStatus, reviewerID, comment, now); err != nil {
		return err
	}
	return tx.Commit()
}

func lockPendingApproval(ctx context.Context, tx *sql.Tx, approvalID, reviewerID string) (string, string, error) {
	var status, assignedReviewer string
	err := tx.QueryRowContext(ctx, `SELECT status,COALESCE(current_reviewer_id,'') FROM approval_requests WHERE id=? FOR UPDATE`, approvalID).Scan(&status, &assignedReviewer)
	if err != nil {
		return "", "", err
	}
	if status != "pending" {
		return "", "", errors.New("approval is not pending")
	}
	var stepID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM approval_steps WHERE approval_id=? AND decision='pending' ORDER BY sequence_no LIMIT 1 FOR UPDATE`, approvalID).Scan(&stepID)
	return status, stepID, err
}

func nextApprovalState(ctx context.Context, tx *sql.Tx, approvalID, decision string) (string, sql.NullString, error) {
	var reviewer sql.NullString
	if decision != "approved" {
		return "rejected", reviewer, nil
	}
	err := tx.QueryRowContext(ctx, `SELECT reviewer_id FROM approval_steps WHERE approval_id=? AND decision='pending' ORDER BY sequence_no LIMIT 1`, approvalID).Scan(&reviewer)
	if errors.Is(err, sql.ErrNoRows) {
		return "approved", reviewer, nil
	}
	return "pending", reviewer, err
}

func (r Approvals) Cancel(ctx context.Context, approvalID, applicantID, comment string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status, owner string
	if err = tx.QueryRowContext(ctx, `SELECT status,applicant_id FROM approval_requests WHERE id=? FOR UPDATE`, approvalID).Scan(&status, &owner); err != nil {
		return err
	}
	if status != "pending" {
		return errors.New("approval cannot be cancelled")
	}
	if owner != applicantID {
		return errors.New("only the applicant can cancel an approval")
	}
	now := time.Now().UTC()
	if _, err = tx.ExecContext(ctx, `UPDATE approval_requests SET status='cancelled',current_reviewer_id=NULL,completed_at=? WHERE id=?`, now, approvalID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO approval_state_history(id,approval_id,from_status,to_status,actor_id,comment,created_at) VALUES(?,?,?,?,?,?,?)`, id.New(), approvalID, status, "cancelled", applicantID, comment, now); err != nil {
		return err
	}
	return tx.Commit()
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
