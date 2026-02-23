package repository

import (
	"api-proxy/internal/model"
	"database/sql"
	"errors"
)

const (
	findActiveByPatternAndMethod = "SELECT id, pattern, backend_url, method, created_at, updated_at, inactivated_at FROM route where pattern = ? and method = ? and inactivated_at IS NULL"
	insertRoute                  = "INSERT INTO route (pattern, backend_url, method, updated_at, inactivated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP(6), null)"
	updateRoute                  = "UPDATE route SET backend_url = ?, method = ?, updated_at = CURRENT_TIMESTAMP(6), inactivated_at = ? WHERE id = ?"
)

var ErrNoRowsAffectedOnRouteInsert = errors.New("no rows affected during insertion of route - expected 1 row to be affected")
var ErrNoRowsAffectedOnRouteUpdate = errors.New("no rows affected during update of route - expected at least 1 row to be affected")

// RouteRepository represents an object through which Route queries can be run
type RouteRepository struct {
	db *sql.DB
}

func NewRouteRepository(db *sql.DB) *RouteRepository {
	return &RouteRepository{db}
}

// FindActiveByPatternAndMethod queries and returns an active route matching the pattern and method provided
func (rr *RouteRepository) FindActiveByPatternAndMethod(pattern, method string) (*model.Route, error) {
	var route model.Route
	row := rr.db.QueryRow(findActiveByPatternAndMethod, pattern, method)

	err := row.Scan(
		&route.ID,
		&route.Pattern,
		&route.BackendURL,
		&route.Method,
		&route.CreatedAt,
		&route.UpdatedAt,
		&route.InactivatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &route, nil
}

// Insert creates a new active route in the database and returns it
func (rr *RouteRepository) Insert(route *model.Route) (*model.Route, error) {
	exec, err := rr.db.Exec(insertRoute, route.Pattern, route.BackendURL, route.Method)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := exec.RowsAffected()

	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrNoRowsAffectedOnRouteInsert
	}

	id, err := exec.LastInsertId()

	if err != nil {
		return nil, err
	}

	route.ID = int(id)
	return route, nil
}

// Update updates an existing route in the database and returns the updated data
func (rr *RouteRepository) Update(route *model.Route) (*model.Route, error) {
	exec, err := rr.db.Exec(updateRoute, route.BackendURL, route.Method, route.InactivatedAt, route.ID)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := exec.RowsAffected()

	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrNoRowsAffectedOnRouteUpdate
	}

	return route, nil
}
