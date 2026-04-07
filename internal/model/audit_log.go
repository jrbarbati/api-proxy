package model

import "time"

type EntityType string
type Action string

func (entityType EntityType) String() string {
	return string(entityType)
}

func (action Action) String() string {
	return string(action)
}

const (
	ROUTE      EntityType = "route"
	RATE_LIMIT EntityType = "rate_limit"

	CREATE Action = "create"
	UPDATE Action = "update"
)

type AuditLog struct {
	ID            int        `json:"id"`
	EntityID      int        `json:"entity_id"`
	EntityType    EntityType `json:"entity_type"`
	PerformedByID int        `json:"performed_by_id"`
	Action        Action     `json:"action"`
	CreatedAt     time.Time  `json:"created_at"`
}

type AuditLogFilter struct {
	EntityType    EntityType
	Action        Action
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}
