package repository

import (
	"database/sql"
	"errors"
)

var ErrNoRowsAffectedOnInsert = errors.New("no rows affected during insertion - expected 1 row to be affected")
var ErrNoRowsAffectedOnUpdate = errors.New("no rows affected during update - expected at least 1 row to be affected")

func execInsert(db *sql.DB, query string, args ...any) (int, error) {
	exec, err := db.Exec(query, args...)

	if err != nil {
		return 0, err
	}

	rowsAffected, err := exec.RowsAffected()

	if err != nil {
		return 0, err
	}

	if rowsAffected == 0 {
		return 0, ErrNoRowsAffectedOnInsert
	}

	id, err := exec.LastInsertId()

	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func execUpdate(db *sql.DB, query string, args ...any) error {
	exec, err := db.Exec(query, args...)

	if err != nil {
		return err
	}

	rowsAffected, err := exec.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNoRowsAffectedOnUpdate
	}

	return nil
}

func execDelete(db *sql.DB, query string, args ...any) error {
	_, err := db.Exec(query, args)

	if err != nil {
		return err
	}

	return nil
}
