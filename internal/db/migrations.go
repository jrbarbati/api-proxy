package db

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
)

const (
	migrationsPath             = "migrations"
	findMigrationByFileNameSQL = "SELECT id FROM migration WHERE filename = ?"
	insertMigration            = "INSERT INTO migration (filename, applied_at) VALUES (?, CURRENT_TIMESTAMP)"
)

var ErrNoRowsAffectedOnInsert = errors.New("no rows affected during insertion of migration - expected 1 row to be affected")

//go:embed migrations/*.sql
var migrationFiles embed.FS

// RunMigration ensures all migration files have run
func RunMigration(db *sql.DB) error {
	if err := bootstrapMigrationsTable(db); err != nil {
		return err
	}

	dir, err := fs.ReadDir(migrationFiles, migrationsPath)

	if err != nil {
		return err
	}

	return verifyMigrations(db, dir)
}

func bootstrapMigrationsTable(db *sql.DB) error {
	script, err := fs.ReadFile(migrationFiles, "migrations/bootstrap.sql")

	if err != nil {
		return err
	}

	_, err = db.Exec(string(script))
	return err
}

func verifyMigrations(db *sql.DB, migrationFiles []fs.DirEntry) error {
	for _, file := range migrationFiles {
		if file.Name() == "bootstrap.sql" {
			continue
		}

		isApplied, err := isMigrationApplied(db, file)

		if err == nil && isApplied {
			continue
		}

		if err != nil {
			return err
		}

		err = runMigration(db, file)

		if err != nil {
			return err
		}
	}

	return nil
}

func isMigrationApplied(db *sql.DB, file fs.DirEntry) (bool, error) {
	var id int
	row := db.QueryRow(findMigrationByFileNameSQL, file.Name())

	err := row.Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func runMigration(db *sql.DB, file fs.DirEntry) error {
	migrationScript, readFileErr := fs.ReadFile(migrationFiles, fmt.Sprintf("%v/%v", migrationsPath, file.Name()))

	if readFileErr != nil {
		return readFileErr
	}

	_, err := db.Exec(string(migrationScript))

	if err != nil {
		return err
	}

	return insertNewMigration(db, file)
}

func insertNewMigration(db *sql.DB, file fs.DirEntry) error {
	exec, err := db.Exec(insertMigration, file.Name())

	if err != nil {
		return err
	}

	rowsAffected, err := exec.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNoRowsAffectedOnInsert
	}

	return nil
}
