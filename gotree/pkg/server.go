// pkg/server/server.go
package server

import (
	"yourmodule/internal/server" // Import the internal server
	"yourmodule/internal/session"
	"database/sql"
)

// Server is a wrapper around the internal server.
type Server = server.Server

// NewServer creates a new server instance.
func NewServer(addr string, sm *session.SessionManager, db *sql.DB) *Server { // Add db
	return server.NewServer(addr, sm, db) // Call the internal NewServer
}

// RegisterService registers a service handler.
func (s *Server) RegisterService(name string, handler server.AuthenticatedServiceHandlerFunc) {
	s.RegisterService(name, handler)
}

// Start starts the server.
func (s *Server) Start() error {
	return s.Start()
}

type AuthenticatedServiceHandlerFunc = server.AuthenticatedServiceHandlerFunc

// GetDB returns the database connection pool.  // addded getter
func (s *Server) GetDB() *sql.DB{
	return s.GetDB()
}
