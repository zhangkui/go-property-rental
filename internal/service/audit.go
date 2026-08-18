package service

import (
	"context"
	"encoding/json"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
	"go-property-rental/internal/repository"
)

type AuditService struct{ Repo repository.AuditStore }

func (s AuditService) Write(c context.Context, actor, action, resource, resourceID string, detail any) error {
	raw, _ := json.Marshal(detail)
	return s.Repo.Append(c, entity.AuditLog{ID: id.New(), ActorID: actor, Action: action, Resource: resource, ResourceID: resourceID, Detail: string(raw)})
}
func (s AuditService) List(c context.Context, page, size int, resource, actor string) ([]entity.AuditLog, error) {
	limit, offset := pageValues(page, size)
	return s.Repo.List(c, limit, offset, resource, actor)
}
