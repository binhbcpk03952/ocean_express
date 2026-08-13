package domain

import (
	"context"
	"time"
)

// AuditLog records business logic changes for traceability
type AuditLog struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ActorID     string    `json:"actor_id" gorm:"column:actor_id"`
	Action      string    `json:"action" gorm:"column:action"`
	Entity      string    `json:"entity" gorm:"column:entity"`
	EntityID    string    `json:"entity_id" gorm:"column:entity_id"`
	BeforeState string    `json:"before_state" gorm:"column:before_state;type:jsonb"`
	AfterState  string    `json:"after_state" gorm:"column:after_state;type:jsonb"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

type AuditRepository interface {
	LogAction(ctx context.Context, log *AuditLog) error
}

type AuditUseCase interface {
	LogAction(ctx context.Context, actorID, action, entity, entityID, beforeState, afterState string) error
}

