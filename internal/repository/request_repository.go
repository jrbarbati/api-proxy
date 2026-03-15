package repository

import (
	"api-proxy/internal/model"
	"database/sql"
	"time"
)

const (
	insertRequest              = "INSERT INTO request (method, url, status_code, latency, created_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP(6))"
	deleteAllRequestsOlderThan = "DELETE FROM request WHERE created_at < ?"
)

// RequestRepository represents an object through which Request queries can be run
type RequestRepository struct {
	db *sql.DB
}

func NewRequestRepository(db *sql.DB) *RequestRepository {
	return &RequestRepository{db: db}
}

// Insert creates a new active request in the database and returns it
func (rr *RequestRepository) Insert(request *model.Request) (*model.Request, error) {
	createdId, err := execInsert(
		rr.db,
		insertRequest,
		request.Method,
		request.URL,
		request.StatusCode,
		request.Latency.Milliseconds(),
	)

	if err != nil {
		return nil, err
	}

	request.ID = createdId
	return request, nil
}

func (rr *RequestRepository) DeleteOlderThan(days int) error {
	threshold := time.Now().Add(-time.Hour * 24 * time.Duration(days))

	return execDelete(rr.db, deleteAllRequestsOlderThan, threshold)
}
