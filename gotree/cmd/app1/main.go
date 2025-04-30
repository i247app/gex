// cmd/app1/main.go
package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/gosvr_svr/cmd/app1/service" // Import the service package for app1.
	"github.com/your-org/gosvr_svr/internal/config"    // Import the internal config package.
	"github.com/your-org/gosvr_svr/internal/database"  // Import the internal database package.
	"github.com/your-org/gosvr_svr/internal/logger"      // Import the internal logger package.
	"github.com/your-org/gosvr_svr/internal/middleware" // Import the internal middleware package.
	"github.com/your-org/gosvr_svr/internal/service"   // Import the internal service package.
	"github.com/your-org/gosvr_svr/internal/session"     // Import the internal session package.
)

func main() {
	// Load configuration.  We'll assume you have a way to specify the config
	// file (e.g., via environment variable).  For simplicity, I'm using
	// a default path here.
	cfg, err := config.LoadConfig("./config/config.yaml") // Use a relative path
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Initialize logging.
	l, err := logger.NewLogger(cfg.Log)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer l.Sync() // Ensure buffered logs are flushed before exiting.

    // Set the global logger.
    logger.SetLogger(l)

	// Initialize the database connection.
	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		l.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close() // Ensure the database connection is closed.

	// Initialize the session manager.
	sessionStore, err := session.NewStore(cfg.Session, db)
	if err != nil {
		l.Fatalf("failed to initialize session store: %v", err)
	}
    defer sessionStore.Close()

	// Create the base service.  This holds common dependencies that
	// are shared by all of the application-specific services.
	baseService := service.NewBaseService(cfg, l, db, sessionStore)

	// Initialize the services for app1.
	app1Service := service.NewApp1Service(baseService) // Corrected this line

	// Create a new Gin router.
	r := gin.New()

	// Register global middleware.
	r.Use(middleware.RequestLogger(l))     // Log every request.
	r.Use(middleware.CORSMiddleware())       // Handle CORS headers.
	r.Use(middleware.SessionMiddleware(sessionStore)) // Handle Sessions
	r.Use(gin.Recovery())                 // Recover from panics.

	// Define a simple health check endpoint.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Register the routes for app1.  This is where we connect the
	// service methods to the HTTP endpoints.
	registerApp1Routes(r, app1Service)

	// Determine the port to listen on.
	port := os.Getenv("PORT") // Use environment variable.
	if port == "" {
		port = strconv.Itoa(cfg.Server.Port) // Fallback to config.
	}
	if port == "" {
		port = "8080" // Default port.
	}

	// Start the server.
	l.Infof("starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		l.Fatalf("failed to start server: %v", err)
	}
}

// registerApp1Routes registers the HTTP routes for app1.  This function
// takes the Gin router and the App1Service struct as input.
func registerApp1Routes(r *gin.Engine, s *service.App1Service) {
	// Group the routes for app1 under the /api/app1 path.
	api := r.Group("/api/app1")
	{
		// Example route:  GET /api/app1/echo/:msg
		api.GET("/echo/:msg", s.EchoService.EchoHandler)

		// Login route: POST /api/app1/login
		api.POST("/login", s.LoginService.LoginHandler)

		// Logout route: POST /api/app1/logout
		api.POST("/logout", s.LogoutService.LogoutHandler)

		// Time route: GET /api/app1/time
		api.GET("/time", s.TimeService.TimeHandler)

		// WhoAmI route: GET /api/app1/whoami
		api.GET("/whoami", s.WhoAmIService.WhoAmIHandler)
	}
}
