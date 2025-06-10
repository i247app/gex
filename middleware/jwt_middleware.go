package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/i247app/gex/jwtutil"
	"github.com/i247app/gex/session"
)

var log = fmt.Println

// JwtMiddleware is a middleware that handles JWT authentication and session management.
// It checks for an existing JWT token in the Authorization header, generates a new one if none is found,
// and creates a new session if one doesn't exist.
// It also wraps the response writer to capture the response body.
//
// Deprecated: Use SessionMiddleware with NewJwtSessionProvider instead.
func JwtMiddleware(
	sessionContainer *session.Container,
	jwtToolkit *jwtutil.Toolkit,
	sessionFactory SessionFactory,
	sessionTTL time.Duration,
	sessionContextKey session.SessionRequestContextKey,
) func(http.Handler) http.Handler {
	// Create a JWT session provider
	jwtSessionProvider := NewJwtSessionProvider(sessionContainer, jwtToolkit, sessionFactory, sessionTTL)

	// Use the unified session middleware
	return SessionMiddleware(jwtSessionProvider, sessionContextKey)
}

// Legacy functions kept for backward compatibility but now using the session providers

func getSessionFromSessionKey(sessionContainer *session.Container, sessionKey string) (session.SessionStorer, bool) {
	return sessionContainer.Session(sessionKey)
}

func getOrCreateJwtToken(r *http.Request, jwtToolkit *jwtutil.Toolkit) (*JwtResult, error) {
	// This function is now used by the JWT session provider
	// Keep for backward compatibility but consider it deprecated
	provider := &JwtSessionProvider{
		jwtToolkit: jwtToolkit,
	}
	return provider.getOrCreateJwtToken(r)
}

func getAuthTokenFromJwtToken(jwtToolkit *jwtutil.Toolkit, jwtToken *jwt.Token) (string, error) {
	provider := &JwtSessionProvider{
		jwtToolkit: jwtToolkit,
	}
	return provider.getAuthTokenFromJwtToken(jwtToken)
}

// Additional legacy functions...
func getValidJwtFromRequest(r *http.Request, jwtToolkit *jwtutil.Toolkit) (*JwtResult, error) {
	provider := &JwtSessionProvider{
		jwtToolkit: jwtToolkit,
	}
	return provider.getValidJwtFromRequest(r)
}

func createNewJwtToken(jwtToolkit *jwtutil.Toolkit, sessionKey string) (*jwt.Token, error) {
	provider := &JwtSessionProvider{
		jwtToolkit: jwtToolkit,
	}
	return provider.createNewJwtToken(sessionKey)
}

func initNewSession(sessionKey string, authToken string, source string, sessionContainer *session.Container, sessionFactory SessionFactory, sessionTTL time.Duration) (session.SessionStorer, error) {
	provider := &JwtSessionProvider{
		sessionContainer: sessionContainer,
		sessionFactory:   sessionFactory,
		sessionTTL:       sessionTTL,
	}
	return provider.initNewSession(sessionKey, authToken, source)
}

func refreshSession(sess session.SessionStorer, sessionTTL time.Duration) (session.SessionStorer, error) {
	provider := &JwtSessionProvider{
		sessionTTL: sessionTTL,
	}
	return provider.refreshSession(sess)
}

func isSessionExpired(sess session.SessionStorer) (bool, error) {
	provider := &JwtSessionProvider{}
	return provider.isSessionExpired(sess)
}
