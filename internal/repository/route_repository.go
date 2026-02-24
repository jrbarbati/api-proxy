package repository

import (
	"api-proxy/internal/model"
	"database/sql"
	"errors"
)

const (
	findActiveRoutes   = "SELECT id, pattern, backend_url, method, created_at, updated_at, inactivated_at FROM route where inactivated_at is null"
	patternWhereClause = " AND pattern = ?"
	methodWhereClause  = " AND method = ?"
	findRouteByID      = "SELECT id, pattern, backend_url, method, created_at, updated_at, inactivated_at FROM route where id = ?"
	insertRoute        = "INSERT INTO route (pattern, backend_url, method, updated_at, inactivated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP(6), null)"
	updateRoute        = "UPDATE route SET backend_url = ?, method = ?, updated_at = CURRENT_TIMESTAMP(6), inactivated_at = ? WHERE id = ?"
)

var ErrNoRowsAffectedOnRouteInsert = errors.New("no rows affected during insertion of route - expected 1 row to be affected")
var ErrNoRowsAffectedOnRouteUpdate = errors.New("no rows affected during update of route - expected at least 1 row to be affected")

// RouteRepository represents an object through which Route queries can be run
type RouteRepository struct {
	db *sql.DB
}

type RouteFilter struct {
	Pattern string
	Method  string
}

func NewRouteRepository(db *sql.DB) *RouteRepository {
	return &RouteRepository{db}
}

// FindActiveByFilter queries routes from the DB using the specified filters
func (rr *RouteRepository) FindActiveByFilter(filter *RouteFilter) ([]*model.Route, error) {
	var args []any
	query := findActiveRoutes

	if filter.Pattern != "" {
		query += patternWhereClause
		args = append(args, filter.Pattern)
	}

	if filter.Method != "" {
		query += methodWhereClause
		args = append(args, filter.Method)
	}

	return rr.findRoutes(query, args...)
}

// FindByID queries the DB and returns a single route with matching ID
func (rr *RouteRepository) FindByID(id int) (*model.Route, error) {
	return rr.findRoute(findRouteByID, id)
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

// Delete removes any existing route if it's ID matches the given id
func (rr *RouteRepository) Delete(id int) error {
	return errors.New("not implemented")
}

func (rr *RouteRepository) findRoutes(query string, args ...any) ([]*model.Route, error) {
	routes := make([]*model.Route, 0)

	result, err := rr.db.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer result.Close()

	for result.Next() {
		var route model.Route

		rowErr := result.Scan(
			&route.ID,
			&route.Pattern,
			&route.BackendURL,
			&route.Method,
			&route.CreatedAt,
			&route.UpdatedAt,
			&route.InactivatedAt,
		)

		if rowErr != nil {
			return nil, rowErr
		}

		routes = append(routes, &route)
	}

	if err := result.Err(); err != nil {
		return nil, err
	}

	return routes, nil
}

func (rr *RouteRepository) findRoute(query string, args ...any) (*model.Route, error) {
	var route model.Route
	row := rr.db.QueryRow(query, args...)

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
