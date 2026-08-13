package usecase

import (
	"context"
	"ocean-express-api/internal/domain"
)

type auditUseCase struct {
	repo domain.AuditRepository
}

func NewAuditUseCase(repo domain.AuditRepository) domain.AuditUseCase {
	return &auditUseCase{repo: repo}
}

func (u *auditUseCase) LogAction(ctx context.Context, actorID, action, entity, entityID, beforeState, afterState string) error {
	log := &domain.AuditLog{
		ActorID:     actorID,
		Action:      action,
		Entity:      entity,
		EntityID:    entityID,
		BeforeState: beforeState,
		AfterState:  afterState,
	}
	return u.repo.LogAction(ctx, log)
}
