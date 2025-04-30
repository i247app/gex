// pkg/server/server.go
package server

import (
	"context"
	"log"
	"net/http"
	"time"
)

// Config holds the configuration for the server.  This
// struct is used to pass configuration parameters
// to the New function.
type Config struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// New creates a new Server.  This constructor function
// takes a Config struct, allowing the caller to
// configure the server.
func New(cfg Config, handler http.Handler) Server {
	return &httpServer{
		server: &http.Server{
			Addr:         cfg.Address,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			Handler:      handler, // Use the provided handler.
		},
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

// httpServer is an implementation of the Server interface
// using the standard net/http package.
type httpServer struct {
	server          *http.Server
	shutdownTimeout time.Duration
}

// Start starts the HTTP server.
func (s *httpServer) Start() error {
	log.Printf("Starting server at %s", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop gracefully shuts down the HTTP server.
func (s *httpServer) Stop(ctx context.Context) error {
	log.Printf("Stopping server at %s", s.server.Addr)

	// Create a context with a timeout for the shutdown operation.
	shutdownCtx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancel()

	// Attempt the graceful shutdown.
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
		// Optionally, you could try a forced shutdown if the
		// graceful shutdown fails, but this is generally not recommended
		// unless you're sure you want to drop connections.
		// return s.server.Close() // Force immediate shutdown (less graceful).
		return err
	}

	log.Println("Server stopped gracefully")
	return nil
}

// Router returns the HTTP handler.
func (s *httpServer) Router() http.Handler {
	return s.server.Handler
}

func (s *httpServer) Addr() string {
	return s.server.Addr
}
```pkg/session/session.go`:

```go
package session

import (
	"time"
)

// Session represents a user session.  This struct
// holds the data associated with a logged-in user.
type Session struct {
	ID        string    // Unique session ID.
	UserID    string    // ID of the user associated with the session.
	CreatedAt time.Time // Time the session was created.
	ExpiresAt time.Time // Time the session expires.
	Data      map[string]interface{} // Store additional session-specific data.
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
