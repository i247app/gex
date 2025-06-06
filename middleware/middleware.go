package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/i247app/gex/jwtutil"
	"github.com/i247app/gex/session"
	"github.com/i247app/gex/util"
)

var log = fmt.Println

var (
	ErrMalformedJwt = errors.New("invalid or malformed JWT")

	DefaultSessionTTL = time.Second * 10
)

type SessionFactory func() session.SessionStorer

// JwtMiddleware is a middleware that handles JWT authentication and session management.
// It checks for an existing JWT token in the Authorization header, generates a new one if none is found,
// and creates a new session if one doesn't exist.
// It also wraps the response writer to capture the response body.
func JwtMiddleware(
	sessionContainer *session.Container,
	jwtToolkit *jwtutil.Toolkit,
	sessionFactory SessionFactory,
	sessionTTL time.Duration,
) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				sessionKey     string
				authToken      string
				didAutoRefresh bool
			)

			// Skip entire token and session handling if this header is set
			if r.Header.Get("X-Skip-Session") == "true" {
				next.ServeHTTP(w, r)
				return
			}

			// 1. Get or create a JWT token
			jwtResult, err := getOrCreateJwtToken(r, jwtToolkit)
			var isValidIncomingJwt bool = err == nil && jwtResult.JwtToken != nil && jwtResult.SessionKey != ""
			if !isValidIncomingJwt {
				writeError(w, "error getting or creating jwt", fmt.Errorf("Unauthorized"))
				return
			}
			sessionKey = jwtResult.SessionKey
			authToken = jwtResult.AuthToken

			// 2. Get session
			sess, ok := getSessionFromSessionKey(sessionContainer, sessionKey)
			if sess == nil || !ok {
				tmp, err := initNewSession(sessionKey, authToken, sessionContainer, sessionFactory, sessionTTL)
				if tmp == nil || err != nil {
					log(">> JwtMiddleware: error initializing new session")
					writeError(w, "error initializing new session", fmt.Errorf("Unauthorized"))
					return
				}
				sess = tmp
			}

			// 3. Check for expired session
			isSessionExpired, err := isSessionExpired(sess)
			if isSessionExpired || err != nil {
				didAutoRefresh = true

				log(">> JwtMiddleware: session expired, for now just auto-refreshing...")
				sess, _ = refreshSession(sess, sessionTTL)
				if sess == nil {
					log(">> JwtMiddleware: error refreshing expired session")
					writeError(w, "error refreshing expired session", fmt.Errorf("Unauthorized"))
					return
				}
			}

			// 4. Update session touched_at
			sess.Put("touched_at", time.Now())

			// 5. Set the authToken in the Authorization request header and X-Auth-Token response header

			// Wrap the response writer to capture the response body
			wr := &responseWriterWrapper{
				ResponseWriter: w,
				body:           bytes.NewBuffer(nil),
			}

			// TODO hacky but for now we inject an Authorization header if its missing
			if r.Header.Get("Authorization") == "" {
				r.Header.Add("Authorization", "Bearer "+authToken)
			}
			wr.Header().Set("X-Auth-Token", authToken)

			next.ServeHTTP(wr, r)

			// Notify the client that the session was auto-refreshed
			if didAutoRefresh {
				w.Header().Add("GEX-Session-Auto-Refreshed", "true")
			}

			if wr.statusCode != 0 {
				w.WriteHeader(wr.statusCode)
			}
			w.Write(wr.body.Bytes())
		})
	}
}

func getSessionFromSessionKey(sessionContainer *session.Container, sessionKey string) (session.SessionStorer, bool) {
	return sessionContainer.Session(sessionKey)
}

func writeError(w http.ResponseWriter, tag string, err error) {
	resp := map[string]string{
		"error":  "gex panic: " + err.Error(),
		"tag":    tag,
		"origin": "jwt_middleware",
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type JwtResult struct {
	JwtToken   *jwt.Token
	SessionKey string
	AuthToken  string
}

func getOrCreateJwtToken(r *http.Request, jwtToolkit *jwtutil.Toolkit) (*JwtResult, error) {
	jwtResult, err := getValidJwtFromRequest(r, jwtToolkit)
	if jwtResult != nil && err == nil {
		return &JwtResult{
			JwtToken:   jwtResult.JwtToken,
			SessionKey: jwtResult.SessionKey,
			AuthToken:  jwtResult.AuthToken,
		}, nil
	}

	// Failed to get a valid JWT token from the request
	if err == ErrMalformedJwt {
		log(">> JwtMiddleware: WARNING ignoring your jwt token - totally malformed JWT token")
	} else if jwtResult == nil || jwtResult.JwtToken == nil || jwtResult.SessionKey == "" || jwtResult.AuthToken == "" || err != nil {
		log(">> JwtMiddleware: WARNING ignoring your jwt token - error getting JWT from request:", err)
	} else {
		log(">> JwtMiddleware: jwt ok")
	}

	// Create a new JWT token with a new session key
	sessionKey := util.GenerateSessionKey()
	jwtToken, err := createNewJwtToken(jwtToolkit, sessionKey)
	if jwtToken == nil || err != nil {
		return nil, fmt.Errorf("error creating new JWT token: %v", err)
	}

	authToken, err := getAuthTokenFromJwtToken(jwtToolkit, jwtToken)
	if authToken == "" || err != nil {
		return nil, fmt.Errorf("error getting authToken from JWT token: %v", err)
	}

	return &JwtResult{
		JwtToken:   jwtToken,
		SessionKey: sessionKey,
		AuthToken:  authToken,
	}, nil
}

func getAuthTokenFromJwtToken(jwtToolkit *jwtutil.Toolkit, jwtToken *jwt.Token) (string, error) {
	authToken, err := jwtToolkit.SignToken(jwtToken)
	if authToken == "" || err != nil {
		log(">> JwtMiddleware: error signing JWT:", err)
		return "", err
	}
	return authToken, nil

}

func getValidJwtFromRequest(r *http.Request, jwtToolkit *jwtutil.Toolkit) (*JwtResult, error) {
	// Validate Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("no Authorization header found")
	}

	if authHeader == "Bearer " {
		return nil, fmt.Errorf("no JWT token found in Authorization header")
	}

	// Get JWT Token
	jwtToken, err := jwtToolkit.GetAuthorizationHeaderJwt(authHeader)
	if jwtToken == nil || err != nil {
		// log(">> JwtMiddleware: error converting Authorization header to JWT:", err)
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
	authToken, err := getAuthTokenFromJwtToken(jwtToolkit, jwtToken)
	if authToken == "" || err != nil {
		return nil, fmt.Errorf("error getting authToken from JWT token: %v", err)
	}

	return &JwtResult{
		JwtToken:   jwtToken,
		SessionKey: claims.SessionKey,
		AuthToken:  authToken,
	}, nil
}

func createNewJwtToken(jwtToolkit *jwtutil.Toolkit, sessionKey string) (*jwt.Token, error) {
	// Create a new JWT token with a new session key
	claims := jwtutil.NewClaims(sessionKey)
	jwtToken, err := jwtToolkit.GenerateJwt(claims)
	if jwtToken == nil || err != nil {
		return nil, fmt.Errorf("error generating JWT: %v", err)
	}

	return jwtToken, nil
}

func initNewSession(sessionKey string, authToken string, sessionContainer *session.Container, sessionFactory SessionFactory, sessionTTL time.Duration) (session.SessionStorer, error) {
	sess, _ := sessionContainer.InitSession(sessionKey, sessionFactory())
	sess.Put("source", "gex.jwt_middleware")
	sess.Put("token", authToken)
	sess.Put("is_secure", false)

	now := time.Now()
	sess.Put("created_at", now)
	sess.Put("expires_at", now.Add(sessionTTL))
	sess.Put("touched_at", now)

	return sess, nil
}

func refreshSession(sess session.SessionStorer, sessionTTL time.Duration) (session.SessionStorer, error) {
	now := time.Now()
	sess.Put("expires_at", now.Add(sessionTTL))
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

func isSessionExpired(sess session.SessionStorer) (bool, error) {
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
