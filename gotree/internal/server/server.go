// gosvr_svr/internal/server/server.go
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors" // Use a well-maintained CORS package
	"yourmodule/internal/session" // Import your session package
)

// Server struct to hold server instance and other dependencies.
type Server struct {
	httpServer *http.Server
	router     *mux.Router
	config     Config
	session    session.SessionManager
}

// NewServer creates a new server instance.
func NewServer(cfg Config, sessionManager session.SessionManager) *Server {
	r := mux.NewRouter()
	server := &Server{
		router:  r,
		config:  cfg,
		session: sessionManager,
		httpServer: &http.Server{
			Addr:         cfg.Host + ":" + cfg.Port,
			Handler:      r, //  Use the router here.
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
	return server
}

// Start starts the server.
func (s *Server) Start() error {
	// Apply CORS configuration.  Use a library.
	c := cors.New(cors.Options{
		AllowedOrigins:   s.config.AllowOrigins,
		AllowedHeaders:   s.config.AllowHeaders,
		AllowedMethods:   s.config.AllowMethods,
		AllowCredentials: true, //  Important for sessions in some cases
	})

	s.httpServer.Handler = c.Handler(s.router) // Wrap the handler with CORS.

	// Listen for shutdown signals.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Register your handlers (this is where the application defines them).
	s.registerHandlers()

	// Start the server in a separate goroutine.
	go func() {
		fmt.Printf("Server running on %s\n", s.httpServer.Addr)
		var err error
		if s.config.CertFile != "" && s.config.KeyFile != "" {
			fmt.Println("Starting HTTPS server")
			err = s.httpServer.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
		} else {
			fmt.Println("Starting HTTP server")
			err = s.httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v\n", err)
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()
	fmt.Println("Shutting down server...")

	// Create a timeout context for graceful shutdown.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server.
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Server shutdown error: %v\n", err)
	}

	fmt.Println("Server shut down gracefully.")
	return nil
}

// RegisterHandler registers an HTTP handler for a specific path and method.
func (s *Server) RegisterHandler(path string, handler http.HandlerFunc, methods ...string) {
	s.router.HandleFunc(path).Methods(methods...)
}

// RegisterMiddleware registers middleware for the server.
func (s *Server) RegisterMiddleware(middleware func(http.Handler) http.Handler) {
	s.router.Use(middleware)
}

// Router returns the router.
func (s *Server) Router() *mux.Router {
	return s.router
}

func (s *Server) registerHandlers() {
	// Handlers are registered by the application in main.go.
}

