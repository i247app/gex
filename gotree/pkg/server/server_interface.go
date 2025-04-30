// pkg/server/server_interface.go`:
package server

import (
	"context"
	"net/http"
)

// Server defines the interface for a server.  This interface
// allows you to swap out different server implementations
// (e.g., net/http, or a testing server) without changing
// your application code.
type Server interface {
	// Start starts the server.  It should block until the server
	// is ready to serve requests (or encounters an error).
	Start() error

	// Stop gracefully shuts down the server.  It should wait for
	// any in-flight requests to complete (up to a timeout).
	Stop(ctx context.Context) error

	// Router returns the underlying HTTP handler (e.g., *http.ServeMux,
	// or a router from a framework like chi or mux).  This allows
	// the application to register its handlers.
	Router() http.Handler
	// Addr returns the address the server is listening on.
	Addr() string
}
