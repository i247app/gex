// cmd/app1/service/service.go
package service

import (
	"github.com/your-org/gosvr_svr/cmd/app1/service/echo"     // Import the echo service.
	"github.com/your-org/gosvr_svr/cmd/app1/service/login"    // Import the login service.
	"github.com/your-org/gosvr_svr/cmd/app1/service/logout"   // Import the logout service.
	"github.com/your-org/gosvr_svr/cmd/app1/service/time"     // Import the time service.
	"github.com/your-org/gosvr_svr/cmd/app1/service/whoami"   // Import the whoami service.
	"github.com/your-org/gosvr_svr/internal/service"         // Import the internal service package.
	"github.com/your-org/gosvr_svr/internal/session"       // Import the internal session package
)

// App1Service holds all the service implementations for app1.  This
// struct acts as a central registry for the services.
type App1Service struct {
	EchoService   *echo.Service
	LoginService  *login.Service
	LogoutService *logout.Service
	TimeService   *time.Service
	WhoAmIService *whoami.Service
}

// NewApp1Service creates a new App1Service struct, initializing
// all the individual services.  It takes a pointer to the global
// BaseService struct, which contains common dependencies.
func NewApp1Service(baseService *service.BaseService) *App1Service {
	return &App1Service{
		EchoService:   echo.NewService(baseService),
		LoginService:  login.NewService(baseService),
		LogoutService: logout.NewService(baseService),
		TimeService:   time.NewService(baseService),
		WhoAmIService: whoami.NewService(baseService),
	}
}
