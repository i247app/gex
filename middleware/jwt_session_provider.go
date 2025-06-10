package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/i247app/gex/jwtutil"
	"github.com/i247app/gex/session"
	"github.com/i247app/gex/util"
)

// JwtSessionProvider implements SessionProvider for JWT-based authentication
type JwtSessionProvider struct {
	sessionContainer *session.Container
	jwtToolkit       *jwtutil.Toolkit
	sessionFactory   SessionFactory
	sessionTTL       time.Duration
}

// NewJwtSessionProvider creates a new JWT session provider
func NewJwtSessionProvider(
	sessionContainer *session.Container,
	jwtToolkit *jwtutil.Toolkit,
	sessionFactory SessionFactory,
	sessionTTL time.Duration,
) *JwtSessionProvider {
	return &JwtSessionProvider{
		sessionContainer: sessionContainer,
		jwtToolkit:       jwtToolkit,
		sessionFactory:   sessionFactory,
		sessionTTL:       sessionTTL,
	}
}

// GetSession implements SessionProvider interface for JWT authentication
func (j *JwtSessionProvider) GetSession(r *http.Request) (session.SessionStorer, error) {
	result, err := j.GetSessionWithMetadata(r)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.Session, nil
}

// GetSessionWithMetadata implements SessionProvider interface with additional metadata
func (j *JwtSessionProvider) GetSessionWithMetadata(r *http.Request) (*SessionResult, error) {
	var didAutoRefresh bool

	// Skip session handling if this header is set
	if r.Header.Get("X-Skip-Session") == "true" {
		return nil, nil
	}

	// 1. Get or create a JWT token
	jwtResult, err := j.getOrCreateJwtToken(r)
	if err != nil {
		return nil, fmt.Errorf("error getting or creating jwt: %w", err)
	}

	sessionKey := jwtResult.SessionKey
	authToken := jwtResult.AuthToken

	// 2. Get or create session
	sess, ok := j.sessionContainer.Session(sessionKey)
	if sess == nil || !ok {
		sess, err = j.initNewSession(sessionKey, authToken, "gex.jwt_session_provider")
		if err != nil {
			return nil, fmt.Errorf("error initializing new session: %w", err)
		}
	}

	// 3. Check for expired session and refresh if needed
	isExpired, err := j.isSessionExpired(sess)
	if isExpired || err != nil {
		didAutoRefresh = true
		log(">> JwtSessionProvider: session expired, auto-refreshing...")
		sess, err = j.refreshSession(sess)
		if err != nil {
			return nil, fmt.Errorf("error refreshing expired session: %w", err)
		}
	}

	// 4. Update session touched_at
	sess.Put("touched_at", time.Now())

	// 5. Set auth token in request header for downstream handlers
	if r.Header.Get("Authorization") == "" {
		r.Header.Add("Authorization", "Bearer "+authToken)
	}

	return &SessionResult{
		Session:        sess,
		DidAutoRefresh: didAutoRefresh,
		AuthToken:      authToken,
	}, nil
}

func (j *JwtSessionProvider) getOrCreateJwtToken(r *http.Request) (*JwtResult, error) {
	jwtResult, err := j.getValidJwtFromRequest(r)
	if jwtResult != nil && err == nil {
		return jwtResult, nil
	}

	// Failed to get a valid JWT token from the request
	if err == ErrMalformedJwt {
		log(">> JwtSessionProvider: WARNING ignoring your jwt token - totally malformed JWT token")
	} else if jwtResult == nil || jwtResult.JwtToken == nil || jwtResult.SessionKey == "" || jwtResult.AuthToken == "" || err != nil {
		log(">> JwtSessionProvider: WARNING ignoring your jwt token - error getting JWT from request:", err)
	} else {
		log(">> JwtSessionProvider: jwt ok")
	}

	// Create a new JWT token with a new session key
	sessionKey := util.GenerateSessionKey()
	jwtToken, err := j.createNewJwtToken(sessionKey)
	if jwtToken == nil || err != nil {
		return nil, fmt.Errorf("error creating new JWT token: %v", err)
	}

	authToken, err := j.getAuthTokenFromJwtToken(jwtToken)
	if authToken == "" || err != nil {
		return nil, fmt.Errorf("error getting authToken from JWT token: %v", err)
	}

	return &JwtResult{
		JwtToken:   jwtToken,
		SessionKey: sessionKey,
		AuthToken:  authToken,
	}, nil
}

func (j *JwtSessionProvider) getAuthTokenFromJwtToken(jwtToken *jwt.Token) (string, error) {
	authToken, err := j.jwtToolkit.SignToken(jwtToken)
	if authToken == "" || err != nil {
		log(">> JwtSessionProvider: error signing JWT:", err)
		return "", err
	}
	return authToken, nil
}

func (j *JwtSessionProvider) getValidJwtFromRequest(r *http.Request) (*JwtResult, error) {
	// Validate Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("no Authorization header found")
	}

	if authHeader == "Bearer " {
		return nil, fmt.Errorf("no JWT token found in Authorization header")
	}

	// Get JWT Token
	jwtToken, err := j.jwtToolkit.GetAuthorizationHeaderJwt(authHeader)
	if jwtToken == nil || err != nil {
		return nil, ErrMalformedJwt
	}

	// Validate JWT Token
	claims, ok := jwtToken.Claims.(*jwtutil.CustomClaims)
	if claims == nil || !ok {
		return nil, fmt.Errorf("invalid JWT token")
	}

	// Validate Session Key
	if claims.SessionKey == "" {
		return nil, fmt.Errorf("no session key found in JWT token")
	}

	// Validate authToken
	authToken, err := j.getAuthTokenFromJwtToken(jwtToken)
	if authToken == "" || err != nil {
		return nil, fmt.Errorf("error getting authToken from JWT token: %v", err)
	}

	return &JwtResult{
		JwtToken:   jwtToken,
		SessionKey: claims.SessionKey,
		AuthToken:  authToken,
	}, nil
}

func (j *JwtSessionProvider) createNewJwtToken(sessionKey string) (*jwt.Token, error) {
	// Create a new JWT token with a new session key
	claims := jwtutil.NewClaims(sessionKey)
	jwtToken, err := j.jwtToolkit.GenerateJwt(claims)
	if jwtToken == nil || err != nil {
		return nil, fmt.Errorf("error generating JWT: %v", err)
	}

	return jwtToken, nil
}

func (j *JwtSessionProvider) initNewSession(sessionKey string, authToken string, source string) (session.SessionStorer, error) {
	sess, _ := j.sessionContainer.InitSession(sessionKey, j.sessionFactory())
	sess.Put("key", sessionKey)
	sess.Put("source", source)
	sess.Put("token", authToken)
	sess.Put("is_secure", false)

	now := time.Now()
	sess.Put("created_at", now)
	sess.Put("expires_at", now.Add(j.sessionTTL))
	sess.Put("touched_at", now)

	return sess, nil
}

func (j *JwtSessionProvider) refreshSession(sess session.SessionStorer) (session.SessionStorer, error) {
	now := time.Now()
	sess.Put("expires_at", now.Add(j.sessionTTL))
	sess.Put("touched_at", now)

	// Increment refresh count
	refreshCountRaw, ok := sess.Get("refresh_count")
	if !ok {
		sess.Put("refresh_count", 1)
	} else {
		refreshCount, ok := refreshCountRaw.(int)
		if !ok {
			sess.Put("refresh_count", 1)
		} else {
			sess.Put("refresh_count", refreshCount+1)
		}
	}

	return sess, nil
}

func (j *JwtSessionProvider) isSessionExpired(sess session.SessionStorer) (bool, error) {
	expiresAtRaw, ok := sess.Get("expires_at")
	if !ok {
		return false, fmt.Errorf("no expires_at found in session")
	}

	expiresAt, ok := expiresAtRaw.(time.Time)
	if !ok {
		return false, fmt.Errorf("error converting expires_at to time.Time")
	}

	return expiresAt.Before(time.Now()), nil
}
