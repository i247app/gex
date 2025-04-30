// internal/database/db.go
package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// DBConfig holds the database configuration.
type DBConfig struct {
	DriverName      string
	ConnString      string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewDBConnection creates a new database connection pool.
func NewDBConnection(cfg DBConfig) (*sql.DB, error) {
	// Open a database connection.
	db, err := sql.Open(cfg.DriverName, cfg.ConnString)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	// Set connection pool parameters.
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test the connection.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Connected to database")
	return db, nil
}
