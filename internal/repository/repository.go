package repository

import (
	"api-proxy/internal/model"
	"database/sql"
	"errors"
)

var ErrNoRowsAffectedOnInsert = errors.New("no rows affected during insertion - expected 1 row to be affected")
var ErrNoRowsAffectedOnUpdate = errors.New("no rows affected during update - expected at least 1 row to be affected")

type Repository interface {
	DB() *sql.DB
}

func insert[T Repository, F model.Identifiable](repository T, data F, query string, args ...any) (*F, error) {
	exec, err := repository.DB().Exec(query, args...)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := exec.RowsAffected()

	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrNoRowsAffectedOnInsert
	}

	id, err := exec.LastInsertId()

	if err != nil {
		return nil, err
	}

	data.SetID(int(id))
	return &data, nil
}

func update[T Repository, F model.Identifiable](repository T, data F, query string, args ...any) (*F, error) {
	exec, err := repository.DB().Exec(query, args...)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := exec.RowsAffected()

	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrNoRowsAffectedOnUpdate
	}

	return &data, nil
}

func deleteById[T Repository](query string, id int, repository T) error {
	_, err := repository.DB().Exec(query, id)

	if err != nil {
		return err
	}

	return nil
}
