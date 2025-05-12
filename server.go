package gex

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/cors"
)

type HostConfig struct {
	ServerHost    string
	ServerPort    string
	HttpsCertFile string
	HttpsKeyFile  string
}

type App struct {
	HostConfig HostConfig

	mux        *gexMux
	server     *http.Server
	onShutdown []func()
}

type Middleware func(http.Handler) http.Handler

func NewApp(hostConfig HostConfig, defaultRoute http.HandlerFunc) *App {
	// Create the mux
	mux := &gexMux{
		mux:            http.NewServeMux(),
		defaultHandler: defaultRoute,
	}

	// Create the server
	address := fmt.Sprintf("%s:%s", hostConfig.ServerHost, hostConfig.ServerPort)
	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	return &App{
		HostConfig: hostConfig,
		mux:        mux,
		server:     server,
	}
}

func (a *App) AddRoute(path string, handler http.HandlerFunc, middleware ...Middleware) {
	a.mux.addRoute(path, handler, middleware...)
}

func (a *App) RegisterMiddleware(middleware Middleware) {
	a.server.Handler = middleware(a.server.Handler)
}

func (a *App) SetupServerCORS() {
	a.server.Handler = cors.AllowAll().Handler(a.server.Handler)
}

func (a *App) OnShutdown(cleanupFunc func()) {
	a.onShutdown = append(a.onShutdown, cleanupFunc)
}

/**
 * Runs the server while listening for shutdown signals
 */
func (a *App) Start() error {
	// Listen for shutdown signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start the server in separate goroutine
	go func() {
		fmt.Printf("Server running on %s\n", a.server.Addr)

		var err error
		if a.HostConfig.HttpsCertFile == "" || a.HostConfig.HttpsKeyFile == "" {
			fmt.Println("WARNING: Starting server without TLS")
			err = a.server.ListenAndServe()
		} else {
			err = a.server.ListenAndServeTLS(a.HostConfig.HttpsCertFile, a.HostConfig.HttpsKeyFile)
		}
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server failed: %v\n", err)
			stop()
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Call all shutdown hooks
	for _, hook := range a.onShutdown {
		hook()
	}

	// Shutdown gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %v", err)
	}

	fmt.Println("Server shutdown successfully")
	return nil
}
