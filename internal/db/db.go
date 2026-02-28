package db

import (
	"api-proxy/internal/config"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Connect builds a connection to a database, specified by the config, verifies it can connect and returns the connection pool
func Connect(dbConfig *config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=true", dbConfig.Username, dbConfig.Password, dbConfig.URL, dbConfig.Port, dbConfig.DBName))

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(15 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
