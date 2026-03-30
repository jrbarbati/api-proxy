package repository

import (
	"api-proxy/internal/model"
	"database/sql"
	"fmt"
	"time"
)

const (
	findRequestsBetween        = "SELECT id, route_id, method, url, status_code, latency, created_at FROM request WHERE ? <= created_at AND created_at <= ?"
	insertRequest              = "INSERT INTO request (route_id, method, url, status_code, latency, created_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP(6))"
	deleteAllRequestsOlderThan = "DELETE FROM request WHERE created_at < ?"
)

// RequestRepository represents an object through which Request queries can be run
type RequestRepository struct {
	db *sql.DB
}

func NewRequestRepository(db *sql.DB) *RequestRepository {
	return &RequestRepository{db: db}
}

func (rr *RequestRepository) FindBetween(start time.Time, end time.Time) ([]*model.Request, error) {
	return rr.findRequests(findRequestsBetween, start, end)
}

// Insert creates a new active request in the database and returns it
func (rr *RequestRepository) Insert(request *model.Request) (*model.Request, error) {
	createdId, err := execInsert(
		rr.db,
		insertRequest,
		request.RouteID,
		request.Method,
		request.URL,
		request.StatusCode,
		request.Latency,
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

func (rr *RequestRepository) findRequests(query string, args ...any) ([]*model.Request, error) {
	requests := make([]*model.Request, 0)

	result, err := rr.db.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer result.Close()

	for result.Next() {
		var request model.Request

		rowErr := result.Scan(
			&request.ID,
			&request.RouteID,
			&request.Method,
			&request.URL,
			&request.StatusCode,
			&request.Latency,
			&request.CreatedAt,
		)

		if rowErr != nil {
			return nil, fmt.Errorf("error scanning result set: %w", rowErr)
		}

		requests = append(requests, &request)
	}

	if err := result.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}
