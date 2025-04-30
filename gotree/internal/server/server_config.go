// gosvr_svr/internal/server/server_config.go
package server

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the server configuration.
type Config struct {
	Host         string        `json:"host"`
	Port         string        `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	AllowOrigins []string      `json:"allow_origins"`
	AllowHeaders []string      `json:"allow_headers"`
	AllowMethods []string      `json:"allow_methods"`
	CertFile     string        `json:"cert_file"` // Added for HTTPS
	KeyFile      string        `json:"key_file"`  // Added for HTTPS
}

// NewConfig creates a new server configuration.
func NewConfig() Config {
	cfg := Config{
		Host:         "localhost",
		Port:         "8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		AllowOrigins: []string{"*"}, // Consider tightening this in production.
		AllowHeaders: []string{"*"}, // Consider tightening this in production.
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Consider limiting this.
		CertFile:     "",
		KeyFile:      "",
	}

	// Load configuration from .env file.
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file: using default values") // Don't fatal, use defaults
	}

	// Override configuration values from environment variables if set.
	cfg.Host = getEnvString("SERVER_HOST", cfg.Host)
	cfg.Port = getEnvString("SERVER_PORT", cfg.Port)
	cfg.ReadTimeout = getEnvDuration("SERVER_READ_TIMEOUT", cfg.ReadTimeout)
	cfg.WriteTimeout = getEnvDuration("SERVER_WRITE_TIMEOUT", cfg.WriteTimeout)
	cfg.IdleTimeout = getEnvDuration("SERVER_IDLE_TIMEOUT", cfg.IdleTimeout)
	cfg.AllowOrigins = getEnvSlice("SERVER_ALLOW_ORIGINS", cfg.AllowOrigins)
	cfg.AllowHeaders = getEnvSlice("SERVER_ALLOW_HEADERS", cfg.AllowHeaders)
	cfg.AllowMethods = getEnvSlice("SERVER_ALLOW_METHODS", cfg.AllowMethods)
	cfg.CertFile = getEnvString("SERVER_CERT_FILE", cfg.CertFile) // Load cert file path
	cfg.KeyFile = getEnvString("SERVER_KEY_FILE", cfg.KeyFile)    // Load key file path

	return cfg
}

// Helper function to get string value from environment variable.
func getEnvString(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper function to get time.Duration value from environment variable.
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr != "" {
		duration, err := time.ParseDuration(valueStr)
		if err != nil {
			log.Printf("Error parsing %s as duration: %v, using default value: %v", key, err, defaultValue)
			return defaultValue
		}
		return duration
	}
	return defaultValue
}

// Helper function to get int value from an environment variable.
func getEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr != "" {
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			log.Printf("Error parsing %s as int: %v, using default value: %d", key, err, defaultValue)
			return defaultValue
		}
		return value
	}
	return defaultValue
}

// Helper function to get a string slice from an environment variable.
// The environment variable is expected to be a comma-separated string.
func getEnvSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr != "" {
		return splitString(valueStr, ",")
	}
	return defaultValue
}

// Helper function to split a string by a delimiter (comma in this case).
func splitString(s, delimiter string) []string {
	var result []string
	if s == "" {
		return result
	}
	parts := os.SplitList(s, delimiter) // Use os.SplitList
	for _, p := range parts {
		if p != "" { // added check to prevent empty strings in slice
			result = append(result, p)
		}
	}
	return result
}
