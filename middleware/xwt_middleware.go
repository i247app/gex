package middleware

import (
	"net/http"
	"time"

	"github.com/i247app/gex/jwtutil"
	"github.com/i247app/gex/session"
)

// XwtMiddleware is a middleware that handles XWT token authentication and session management.
// It behaves similarly to JwtMiddleware, but uses the signed JWT token string as the session key.
//
// Deprecated: Use SessionMiddleware with NewXwtSessionProvider instead.
func XwtMiddleware(
	sessionContainer *session.Container,
	jwtToolkit *jwtutil.Toolkit,
	sessionFactory SessionFactory,
	sessionTTL time.Duration,
	sessionContextKey session.SessionRequestContextKey,
) func(http.Handler) http.Handler {
	// Create an XWT session provider
	xwtSessionProvider := NewXwtSessionProvider(sessionContainer, jwtToolkit, sessionFactory, sessionTTL)

	// Use the unified session middleware
	return SessionMiddleware(xwtSessionProvider, sessionContextKey)
}

// Legacy functions kept for backward compatibility

func getOrCreateXwtToken(r *http.Request, jwtToolkit *jwtutil.Toolkit) (*XwtResult, error) {
	provider := &XwtSessionProvider{
		jwtToolkit: jwtToolkit,
	}
	return provider.getOrCreateXwtToken(r)
}

func getValidXwtFromRequest(r *http.Request) (*XwtResult, error) {
	provider := &XwtSessionProvider{}
	return provider.getValidXwtFromRequest(r)
}

func createNewXwtToken(jwtToolkit *jwtutil.Toolkit) (*XwtResult, error) {
	provider := &XwtSessionProvider{
		jwtToolkit: jwtToolkit,
	}
	return provider.createNewXwtToken()
}
