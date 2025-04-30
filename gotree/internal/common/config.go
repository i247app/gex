// app_svr/internal/common/config.go
package common

import (
	"log"
	"os"
	"strconv"
)

type AppConfig struct {
	Port int
	Env  string
}

var Config AppConfig

func LoadConfig() {
	portStr := os.Getenv("APP_PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid APP_PORT: %v", err)
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	Config = AppConfig{
		Port: port,
		Env:  env,
	}
}

