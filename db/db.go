package db

import (
	"database/sql"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
	DriverName string
}

func NewDB(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &DB{DB: db, DriverName: "postgres"}, nil
}

// NewDBWithDriver allows specifying the SQL driver (e.g., "sqlite3" or "postgres")
func NewDBWithDriver(driver, dsn string) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &DB{DB: db, DriverName: driver}, nil
}

func (db *DB) Driver() string {
	return db.DriverName
}
