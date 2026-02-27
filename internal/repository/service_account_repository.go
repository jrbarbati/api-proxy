package repository

import (
	"api-proxy/internal/model"
	"database/sql"
	"errors"
)

const (
	findActiveServiceAccounts    = "SELECT id, org_id, identifier, client_id, client_secret, created_at, updated_at, inactivated_at FROM service_account where inactivated_at is null"
	identifierWhereClause        = " AND identifier = ?"
	clientIdWhereClause          = " AND client_id = ?"
	findServiceAccountByID       = "SELECT id, org_id, identifier, client_id, client_secret, created_at, updated_at, inactivated_at FROM service_account where id = ?"
	findServiceAccountByClientID = "SELECT id, org_id, identifier, client_id, client_secret, created_at, updated_at, inactivated_at FROM service_account where client_id = ?"
	insertServiceAccount         = "INSERT INTO service_account (org_id, identifier, client_id, client_secret, updated_at, inactivated_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP(6), null)"
	updateServiceAccount         = "UPDATE service_account SET identifier = ?, client_id = ?, client_secret = ?, updated_at = CURRENT_TIMESTAMP(6), inactivated_at = ? WHERE id = ?"
	deleteServiceAccount         = "DELETE FROM service_account WHERE id = ?"
)

// ServiceAccountRepository represents an object through which ServiceAccount queries can be run
type ServiceAccountRepository struct {
	db *sql.DB
}

type ServiceAccountFilter struct {
	Identifier string
	ClientID   string
}

func NewServiceAccountRepository(db *sql.DB) *ServiceAccountRepository {
	return &ServiceAccountRepository{db}
}

// FindActiveByFilter queries service accounts from the DB using the specified filters
func (sar *ServiceAccountRepository) FindActiveByFilter(filter *ServiceAccountFilter) ([]*model.ServiceAccount, error) {
	var args []any
	query := findActiveServiceAccounts

	if filter.Identifier != "" {
		query += identifierWhereClause
		args = append(args, filter.Identifier)
	}

	if filter.ClientID != "" {
		query += clientIdWhereClause
		args = append(args, filter.ClientID)
	}

	return sar.findServiceAccounts(query, args...)
}

// FindByID queries the DB and returns a single service account with matching ID
func (sar *ServiceAccountRepository) FindByID(id int) (*model.ServiceAccount, error) {
	return sar.findServiceAccount(findServiceAccountByID, id)
}

// FindByClientID queries the DB and returns a single service account with matching client id
func (sar *ServiceAccountRepository) FindByClientID(clientID string) (*model.ServiceAccount, error) {
	return sar.findServiceAccount(findServiceAccountByClientID, clientID)
}

// Insert creates a new active service account in the database and returns it
func (sar *ServiceAccountRepository) Insert(serviceAccount *model.ServiceAccount) (*model.ServiceAccount, error) {
	createdId, err := execInsert(sar.db, insertServiceAccount, serviceAccount.OrgID, serviceAccount.Identifier, serviceAccount.ClientID, serviceAccount.ClientSecret)

	if err != nil {
		return nil, err
	}

	serviceAccount.ID = createdId
	return serviceAccount, nil
}

// Update updates an existing service account in the database and returns the updated data
func (sar *ServiceAccountRepository) Update(serviceAccount *model.ServiceAccount) (*model.ServiceAccount, error) {
	err := execUpdate(
		sar.db,
		updateServiceAccount,
		serviceAccount.Identifier,
		serviceAccount.ClientID,
		serviceAccount.ClientSecret,
		serviceAccount.InactivatedAt,
		serviceAccount.ID,
	)

	if err != nil {
		return nil, err
	}

	return serviceAccount, nil
}

// Delete removes any existing service account if it's ID matches the given id
func (sar *ServiceAccountRepository) Delete(id int) error {
	return execDelete(sar.db, deleteServiceAccount, id)
}

func (sar *ServiceAccountRepository) findServiceAccounts(query string, args ...any) ([]*model.ServiceAccount, error) {
	serviceAccounts := make([]*model.ServiceAccount, 0)

	result, err := sar.db.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer result.Close()

	for result.Next() {
		var serviceAccount model.ServiceAccount

		rowErr := result.Scan(
			&serviceAccount.ID,
			&serviceAccount.OrgID,
			&serviceAccount.Identifier,
			&serviceAccount.ClientID,
			&serviceAccount.ClientSecret,
			&serviceAccount.CreatedAt,
			&serviceAccount.UpdatedAt,
			&serviceAccount.InactivatedAt,
		)

		if rowErr != nil {
			return nil, rowErr
		}

		serviceAccounts = append(serviceAccounts, &serviceAccount)
	}

	if err := result.Err(); err != nil {
		return nil, err
	}

	return serviceAccounts, nil
}

func (sar *ServiceAccountRepository) findServiceAccount(query string, args ...any) (*model.ServiceAccount, error) {
	var serviceAccount model.ServiceAccount
	row := sar.db.QueryRow(query, args...)

	err := row.Scan(
		&serviceAccount.ID,
		&serviceAccount.OrgID,
		&serviceAccount.Identifier,
		&serviceAccount.ClientID,
		&serviceAccount.ClientSecret,
		&serviceAccount.CreatedAt,
		&serviceAccount.UpdatedAt,
		&serviceAccount.InactivatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &serviceAccount, nil
}
