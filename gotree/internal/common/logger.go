// app_svr/internal/common/logger.go
package common

import (
	"log"
	"os"
)

var Logger *log.Logger

func InitLogger() {
	Logger = log.New(os.Stdout, "[app_svr] ", log.LstdFlags|log.Lshortfile)
}

