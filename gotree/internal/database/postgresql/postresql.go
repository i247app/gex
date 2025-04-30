// internal/database/postgresql/postgresql.go
package postgresql

import (
	"fmt"
	"os"
)

// NewPostgreSQLConnection creates a new PostgreSQL database connection pool.
func NewPostgreSQLConnection() (string, error) {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432" // Default PostgreSQL port
	}

	connString := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)
	return connString, nil
}
