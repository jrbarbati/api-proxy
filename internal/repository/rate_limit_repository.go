package repository

import (
	"api-proxy/internal/model"
	"database/sql"
	"errors"
)

const (
	findActiveRateLimits        = "SELECT id, org_id, service_account_id, limit_per_minute, limit_per_day, limit_per_month, created_at, updated_at, inactivated_at FROM rate_limit where inactivated_at is null"
	orgIdWhereClause            = " AND org_id = ?"
	serviceAccountIdWhereClause = " AND service_account_id = ?"
	findRateLimitByID           = "SELECT id, org_id, service_account_id, limit_per_minute, limit_per_day, limit_per_month, created_at, updated_at, inactivated_at FROM service_account where id = ?"
	insertRateLimit             = "INSERT INTO rate_limit (org_id, service_account_id, limit_per_minute, limit_per_day, limit_per_month, updated_at, inactivated_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP(6), null)"
	updateRateLimit             = "UPDATE rate_limit SET service_account_id = ?, limit_per_minute = ?, limit_per_day = ?, limit_per_month = ?, updated_at = CURRENT_TIMESTAMP(6), inactivated_at = ? WHERE id = ?"
)

var ErrNoRowsAffectedOnRateLimitInsert = errors.New("no rows affected during insertion of rate limit - expected 1 row to be affected")
var ErrNoRowsAffectedOnRateLimitUpdate = errors.New("no rows affected during update of rate limit - expected at least 1 row to be affected")

// RateLimitRepository represents an object through which RateLimit queries can be run
type RateLimitRepository struct {
	db *sql.DB
}

type RateLimitFilter struct {
	OrgId            string
	ServiceAccountId string
}

func NewRateLimitRepository(db *sql.DB) *RateLimitRepository {
	return &RateLimitRepository{db}
}

// FindActiveByFilter queries rate limits from the DB using the specified filters
func (sar *RateLimitRepository) FindActiveByFilter(filter *RateLimitFilter) ([]*model.RateLimit, error) {
	var args []any
	query := findActiveRateLimits

	if filter.OrgId != "" {
		query += orgIdWhereClause
		args = append(args, filter.OrgId)
	}

	if filter.ServiceAccountId != "" {
		query += serviceAccountIdWhereClause
		args = append(args, filter.ServiceAccountId)
	}

	return sar.findRateLimits(query, args...)
}

// FindByID queries the DB and returns a single rate limit with matching ID
func (sar *RateLimitRepository) FindByID(id int) (*model.RateLimit, error) {
	return sar.findRateLimit(findRateLimitByID, id)
}

// Insert creates a new active rate limit in the database and returns it
func (sar *RateLimitRepository) Insert(rateLimit *model.RateLimit) (*model.RateLimit, error) {
	exec, err := sar.db.Exec(
		insertRateLimit,
		rateLimit.OrgID,
		rateLimit.ServiceAccountID,
		rateLimit.LimitPerMinute,
		rateLimit.LimitPerDay,
		rateLimit.LimitPerMonth,
	)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := exec.RowsAffected()

	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrNoRowsAffectedOnRateLimitInsert
	}

	id, err := exec.LastInsertId()

	if err != nil {
		return nil, err
	}

	rateLimit.ID = int(id)
	return rateLimit, nil
}

// Update updates an existing rate limit in the database and returns the updated data
func (sar *RateLimitRepository) Update(rateLimit *model.RateLimit) (*model.RateLimit, error) {
	exec, err := sar.db.Exec(
		updateRateLimit,
		rateLimit.ServiceAccountID,
		rateLimit.LimitPerMinute,
		rateLimit.LimitPerDay,
		rateLimit.LimitPerMonth,
		rateLimit.InactivatedAt,
		rateLimit.ID,
	)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := exec.RowsAffected()

	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrNoRowsAffectedOnRateLimitUpdate
	}

	return rateLimit, nil
}

// Delete removes any existing rate limit if it's ID matches the given id
func (sar *RateLimitRepository) Delete(id int) error {
	return errors.New("not implemented")
}

func (sar *RateLimitRepository) findRateLimits(query string, args ...any) ([]*model.RateLimit, error) {
	rateLimits := make([]*model.RateLimit, 0)

	result, err := sar.db.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer result.Close()

	for result.Next() {
		var rateLimit model.RateLimit

		rowErr := result.Scan(
			&rateLimit.ID,
			&rateLimit.OrgID,
			&rateLimit.ServiceAccountID,
			&rateLimit.LimitPerMinute,
			&rateLimit.LimitPerDay,
			&rateLimit.LimitPerMonth,
			&rateLimit.CreatedAt,
			&rateLimit.UpdatedAt,
			&rateLimit.InactivatedAt,
		)

		if rowErr != nil {
			return nil, rowErr
		}

		rateLimits = append(rateLimits, &rateLimit)
	}

	if err := result.Err(); err != nil {
		return nil, err
	}

	return rateLimits, nil
}

func (sar *RateLimitRepository) findRateLimit(query string, args ...any) (*model.RateLimit, error) {
	var rateLimit model.RateLimit
	row := sar.db.QueryRow(query, args...)

	err := row.Scan(
		&rateLimit.ID,
		&rateLimit.OrgID,
		&rateLimit.ServiceAccountID,
		&rateLimit.LimitPerMinute,
		&rateLimit.LimitPerDay,
		&rateLimit.CreatedAt,
		&rateLimit.UpdatedAt,
		&rateLimit.InactivatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &rateLimit, nil
}
