package logger

import (
	"api-proxy/internal/model"
	"context"
	"log/slog"
)

type AuditLogDataStorer interface {
	Insert(log *model.AuditLog) (*model.AuditLog, error)
}

type Auditable struct {
	EntityID      int
	EntityType    model.EntityType
	PerformedByID int
	Action        model.Action
}

type AuditLogger struct {
	ch        chan Auditable
	dataStore AuditLogDataStorer
}

func NewAuditLogger(dataStore AuditLogDataStorer, channelSize int) *AuditLogger {
	return &AuditLogger{
		ch:        make(chan Auditable, channelSize),
		dataStore: dataStore,
	}
}

func NewAuditable(entityID int, entityType model.EntityType, performedById int, action model.Action) Auditable {
	return Auditable{
		EntityID:      entityID,
		EntityType:    entityType,
		PerformedByID: performedById,
		Action:        action,
	}
}

func (al *AuditLogger) Log(entityID int, entityType model.EntityType, performedById int, action model.Action) {
	select {
	case al.ch <- NewAuditable(entityID, entityType, performedById, action):
	default:
		slog.Error("audit log channel is full, dropping entry")
	}
}

func (al *AuditLogger) Start(ctx context.Context) {
	slog.Info("starting audit logger...")

	go func() {
		for {
			select {
			case auditable := <-al.ch:
				if _, err := al.dataStore.Insert(&model.AuditLog{
					EntityID:      auditable.EntityID,
					EntityType:    auditable.EntityType,
					PerformedByID: auditable.PerformedByID,
					Action:        auditable.Action,
				}); err != nil {
					slog.Error("failed to insert audit log", "err", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
