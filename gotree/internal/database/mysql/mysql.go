// internal/database/mysql/mysql.go
package mysql

import (
	"fmt"
	"os"
)

// NewMySQLConnection creates a new MySQL database connection pool.
func NewMySQLConnection() (string, error) {
	// Database connection details (from environment variables, config, etc.)
	//  Use environment variables for security.
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306" //default port
	}

	// Construct the connection string.
	connString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	return connString, nil
}
