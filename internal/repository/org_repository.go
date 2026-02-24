package repository

import (
	"api-proxy/internal/model"
	"database/sql"
	"errors"
)

const (
	findActiveOrgs = "SELECT id, name, created_at, updated_at, inactivated_at FROM org where inactivated_at is null"
	findOrgByID    = "SELECT id, name, created_at, updated_at, inactivated_at FROM org where id = ?"
	insertOrg      = "INSERT INTO org (name, updated_at, inactivated_at) VALUES (?, CURRENT_TIMESTAMP(6), null)"
	updateOrg      = "UPDATE org SET name = ?, updated_at = CURRENT_TIMESTAMP(6), inactivated_at = ? WHERE id = ?"
	deleteOrg      = "DELETE FROM org WHERE id = ?"
)

var ErrNoRowsAffectedOnOrgInsert = errors.New("no rows affected during insertion of org - expected 1 row to be affected")
var ErrNoRowsAffectedOnOrgUpdate = errors.New("no rows affected during update of org - expected at least 1 row to be affected")

// OrgRepository represents an object through which Org queries can be run
type OrgRepository struct {
	db *sql.DB
}

func NewOrgRepository(db *sql.DB) *OrgRepository {
	return &OrgRepository{db}
}

func (or *OrgRepository) DB() *sql.DB {
	return or.db
}

// FindActive queries orgs from the DB using the specified filters
func (or *OrgRepository) FindActive() ([]*model.Org, error) {
	return or.findOrgs(findActiveOrgs)
}

// FindByID queries the DB and returns a single org with matching ID
func (or *OrgRepository) FindByID(id int) (*model.Org, error) {
	return or.findOrg(findOrgByID, id)
}

// Insert creates a new active org in the database and returns it
func (or *OrgRepository) Insert(org *model.Org) (*model.Org, error) {
	exec, err := or.db.Exec(insertOrg, org.Name)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := exec.RowsAffected()

	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrNoRowsAffectedOnOrgInsert
	}

	id, err := exec.LastInsertId()

	if err != nil {
		return nil, err
	}

	org.ID = int(id)
	return org, nil
}

// Update updates an existing org in the database and returns the updated data
func (or *OrgRepository) Update(org *model.Org) (*model.Org, error) {
	exec, err := or.db.Exec(updateOrg, org.Name, org.InactivatedAt, org.ID)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := exec.RowsAffected()

	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrNoRowsAffectedOnOrgUpdate
	}

	return org, nil
}

// Delete removes any existing org if it's ID matches the given id
func (or *OrgRepository) Delete(id int) error {
	return deleteById[*OrgRepository](deleteOrg, id, or)
}

func (or *OrgRepository) findOrgs(query string, args ...any) ([]*model.Org, error) {
	orgs := make([]*model.Org, 0)

	result, err := or.db.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer result.Close()

	for result.Next() {
		var org model.Org

		rowErr := result.Scan(
			&org.ID,
			&org.Name,
			&org.CreatedAt,
			&org.UpdatedAt,
			&org.InactivatedAt,
		)

		if rowErr != nil {
			return nil, rowErr
		}

		orgs = append(orgs, &org)
	}

	if err := result.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}

func (or *OrgRepository) findOrg(query string, args ...any) (*model.Org, error) {
	var org model.Org
	row := or.db.QueryRow(query, args...)

	err := row.Scan(
		&org.ID,
		&org.Name,
		&org.CreatedAt,
		&org.UpdatedAt,
		&org.InactivatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &org, nil
}
