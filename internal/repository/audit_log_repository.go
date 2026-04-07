package repository

import (
	"api-proxy/internal/model"
	"database/sql"
	"fmt"
)

const (
	selectAll                = "SELECT id, entity_id, entity_name, performed_by_id, action, created_at FROM audit_log WHERE 1 = 1"
	entityTypeClause         = " AND entity_id = ?"
	actionClause             = " AND action = ?"
	createdAfterWhereClause  = " AND created_at > ?"
	createdBeforeWhereClause = " AND created_at < ?"
	insertAuditLog           = "INSERT INTO audit_log (entity_id, entity_name, performed_by_id, action, created_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP(6))"
)

type AuditLogRepository struct {
	db *sql.DB
}

func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{
		db: db,
	}
}

func (alr *AuditLogRepository) FindByFilter(filter *model.AuditLogFilter) ([]*model.AuditLog, error) {
	args := make([]any, 0)
	query := selectAll

	if filter != nil && filter.EntityType != "" {
		query += entityTypeClause
		args = append(args, filter.EntityType)
	}

	if filter != nil && filter.Action != "" {
		query += actionClause
		args = append(args, filter.Action)
	}

	if filter != nil && filter.CreatedAfter != nil {
		query += createdAfterWhereClause
		args = append(args, filter.CreatedAfter)
	}

	if filter != nil && filter.CreatedBefore != nil {
		query += createdBeforeWhereClause
		args = append(args, filter.CreatedBefore)
	}

	return alr.findAuditLogs(query, args...)
}

func (alr *AuditLogRepository) Insert(auditLog *model.AuditLog) (*model.AuditLog, error) {
	id, err := execInsert(
		alr.db,
		insertAuditLog,
		auditLog.EntityID,
		auditLog.EntityType,
		auditLog.PerformedByID,
		auditLog.Action,
		auditLog.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	auditLog.ID = id
	return auditLog, nil
}

func (alr *AuditLogRepository) findAuditLogs(query string, args ...any) ([]*model.AuditLog, error) {
	auditLogs := make([]*model.AuditLog, 0)

	result, err := alr.db.Query(query, args...)

	if err != nil {
		return nil, err
	}
	defer result.Close()

	for result.Next() {
		var auditLog model.AuditLog

		rowErr := result.Scan(
			&auditLog.ID,
			&auditLog.EntityID,
			&auditLog.EntityType,
			&auditLog.PerformedByID,
			&auditLog.Action,
			&auditLog.CreatedAt,
		)

		if rowErr != nil {
			return nil, fmt.Errorf("error scanning result set: %w", rowErr)
		}

		auditLogs = append(auditLogs, &auditLog)
	}

	return auditLogs, nil
}
