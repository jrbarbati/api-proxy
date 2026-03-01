package repository

import (
	"api-proxy/internal/model"
	"database/sql"
	"errors"
)

const (
	findActiveInternalUsers = "SELECT id, email, password, created_at, updated_at, inactivated_at FROM internal_user where inactivated_at is null"
	emailWhereClause        = " AND email = ?"
	findInternalUserByID    = "SELECT id, email, password, created_at, updated_at, inactivated_at FROM internal_user where id = ?"
	findInternalUserByEmail = "SELECT id, email, password, created_at, updated_at, inactivated_at FROM internal_user where email = ? and inactivated_at is null"
	insertInternalUser      = "INSERT INTO internal_user (email, password, updated_at, inactivated_at) VALUES (?, ?, CURRENT_TIMESTAMP(6), null)"
	updateInternalUser      = "UPDATE internal_user SET email = ?, updated_at = CURRENT_TIMESTAMP(6), inactivated_at = ? WHERE id = ?"
	deleteInternalUser      = "DELETE FROM internal_user WHERE id = ?"
)

// InternalUserRepository represents an object through which InternalUser queries can be run
type InternalUserRepository struct {
	db *sql.DB
}

type InternalUserFilter struct {
	Email string
}

func NewInternalUserRepository(db *sql.DB) *InternalUserRepository {
	return &InternalUserRepository{db: db}
}

// FindActive queries internalUsers from the DB using the specified filters
func (iur *InternalUserRepository) FindActive(filter *InternalUserFilter) ([]*model.InternalUser, error) {
	var args []any
	query := findActiveInternalUsers

	if filter.Email != "" {
		query += emailWhereClause
		args = append(args, filter.Email)
	}

	return iur.findInternalUsers(query, args...)
}

// FindByID queries the DB and returns a single internalUser with matching ID
func (iur *InternalUserRepository) FindByID(id int) (*model.InternalUser, error) {
	return iur.findInternalUser(findInternalUserByID, id)
}

// FindByEmail queries the DB and returns a single internalUser with matching Email
func (iur *InternalUserRepository) FindByEmail(email string) (*model.InternalUser, error) {
	return iur.findInternalUser(findInternalUserByEmail, email)
}

// Insert creates a new active internalUser in the database and returns it
func (iur *InternalUserRepository) Insert(internalUser *model.InternalUser) (*model.InternalUser, error) {
	createdId, err := execInsert(iur.db, insertInternalUser, internalUser.Email, internalUser.Password)

	if err != nil {
		return nil, err
	}

	internalUser.ID = createdId
	return internalUser, nil
}

// Update updates an existing internalUser in the database and returns the updated data
func (iur *InternalUserRepository) Update(internalUser *model.InternalUser) (*model.InternalUser, error) {
	err := execUpdate(iur.db, updateInternalUser, internalUser.Email, internalUser.InactivatedAt, internalUser.ID)

	if err != nil {
		return nil, err
	}

	return internalUser, nil
}

// Delete removes any existing internalUser if it's ID matches the given id
func (iur *InternalUserRepository) Delete(id int) error {
	return execDelete(iur.db, deleteInternalUser, id)
}

func (iur *InternalUserRepository) findInternalUsers(query string, args ...any) ([]*model.InternalUser, error) {
	internalUsers := make([]*model.InternalUser, 0)

	result, err := iur.db.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer result.Close()

	for result.Next() {
		var internalUser model.InternalUser

		rowErr := result.Scan(
			&internalUser.ID,
			&internalUser.Email,
			&internalUser.Password,
			&internalUser.CreatedAt,
			&internalUser.UpdatedAt,
			&internalUser.InactivatedAt,
		)

		if rowErr != nil {
			return nil, rowErr
		}

		internalUsers = append(internalUsers, &internalUser)
	}

	if err := result.Err(); err != nil {
		return nil, err
	}

	return internalUsers, nil
}

func (iur *InternalUserRepository) findInternalUser(query string, args ...any) (*model.InternalUser, error) {
	var internalUser model.InternalUser
	row := iur.db.QueryRow(query, args...)

	err := row.Scan(
		&internalUser.ID,
		&internalUser.Email,
		&internalUser.Password,
		&internalUser.CreatedAt,
		&internalUser.UpdatedAt,
		&internalUser.InactivatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &internalUser, nil
}
