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
				newAuthToken   *string
				didAutoRefresh bool
			)

			// Skip entire token and session handling if this header is set
			if r.Header.Get("X-Skip-Session") == "true" {
				next.ServeHTTP(w, r)
				return
			}

			// JWT Middleware Flow:
			//
			// 1. Check for valid JWT token based on the Authorization header, set result to thereIsAnIncomingJwt
			// 2a. If JWT token found, just get the session key
			// 2b. If JWT token NOT found, create a new JWT token with session key
			// 3. Check for a session linked to the JWT token's sessionKey
			// 3a. If session is found, do nothing
			// 3b. 🚨 If session is not found AND thereIsAnIncomingJwt == true, return unauthorized error
			// 3c. If session is not found AND thereIsAnIncomingJwt == false, create a new session
			// 4. Check session expiration
			// 4a. 🚨 If session is expired, return unauthorized error
			// 4b. If session is not expired, do nothing
			// 5. Update session touched_at
			// 6. Determine current authToken
			// 7. Set the session token in the Authorization header

			// 1. Check for valid JWT token based on the Authorization header
			var inJwtSessionKey *string
			jwtToken, err := getJwtFromRequest(r, jwtToolkit)
			if err == ErrMalformedJwt {
				log(">> JwtMiddleware: totally malformed JWT token, returning unauthorized...")
				writeError(w, fmt.Errorf("Unauthorized"))
				return
			} else if jwtToken == nil || err != nil {
				log(">> JwtMiddleware: error getting JWT from request:", err)
			} else {
				claims, ok := jwtToken.Claims.(*jwtutil.CustomClaims)
				if ok {
					inJwtSessionKey = &claims.SessionKey
				}
			}
			isValidIncomingJwt := err == nil && jwtToken != nil && inJwtSessionKey != nil

			if isValidIncomingJwt {
				// 2a. If JWT token found, just get the session key

				log(">> JwtMiddleware: found valid JWT token, using session key from claims...")

				sessionKey = *inJwtSessionKey
			} else {
				// 2b. If JWT token NOT found, create a new JWT token with session key

				log(">> JwtMiddleware: no valid JWT token found, creating a new one...")

				key := util.GenerateSessionKey()
				claims := jwtutil.NewClaims(key)
				tmp, err := jwtToolkit.GenerateJwt(claims)
				if tmp == nil || err != nil {
					log(">> JwtMiddleware: error generating JWT:", err)

					// !! Major error, just return unauthorized
					writeError(w, fmt.Errorf("Unauthorized"))
					return
				}

				sessionKey = key
				jwtToken = tmp
			}

			// 3. Check for a session linked to the JWT token's sessionKey
			sess, ok := sessionContainer.Session(sessionKey)
			if sess != nil && ok {
				// 3a. If session is found, do nothing
			} else {
				if isValidIncomingJwt {
					// 3b. 🚨 If session is not found AND thereIsAnIncomingJwt == true, return unauthorized error

					// NOTE: For now expired sessions will be auto-refreshed,
					// so just create a session same as 3c
					didAutoRefresh = true

					log(">> JwtMiddleware: JWT token exists but no session found, for now just auto-creating a new session...")
				}
				// 3c. If session is not found AND thereIsAnIncomingJwt == false, create a new session

				log(">> JwtMiddleware: no session found (most likely no client JWT detected)")

				// Get the signed Auth token
				authToken, err := jwtToolkit.SignToken(jwtToken)
				if err != nil {
					log(">> JwtMiddleware: error signing JWT:", err)
				}
				newAuthToken = &authToken

				// Initialize a new session
				if authToken != "" {
					sess, _ = initNewSession(sessionKey, authToken, sessionContainer, sessionFactory, sessionTTL)
					if sess == nil {
						log(">> JwtMiddleware: error initializing new session")
					}
				}
			}

			if sess == nil {
				// !! Major error, just return unauthorized
				writeError(w, fmt.Errorf("Unauthorized"))
				return
			}

			// 4. Check session expiration
			isSessionExpired, err := isSessionExpired(sess)
			if isSessionExpired || err != nil {
				didAutoRefresh = true

				log(">> JwtMiddleware: session expired, for now just auto-refreshing...")

				// Initialize a new session
				sess, _ = refreshSession(sess, sessionTTL)
				if sess == nil {
					log(">> JwtMiddleware: error initializing new session")
				}
			} else {
				// 4b. If session is not expired, do nothing
			}

			// 5. Update session touched_at
			if sess != nil {
				sess.Put("touched_at", time.Now())
			}

			// 6. Determine current authToken
			var authToken string
			if newAuthToken != nil {
				authToken = *newAuthToken
			} else {
				authTokenRaw, ok := sess.Get("token")
				if !ok {
					log(">> JwtMiddleware: error getting authToken from session")
				}

				authToken = authTokenRaw.(string)
			}

			rwrap := &responseWriterWrapper{
				ResponseWriter: w,
				body:           bytes.NewBuffer(nil),
			}

			// 6. Set the authToken in the Authorization request header and X-Auth-Token response header

			// TODO hacky but for now we inject an Authorization header if its missing
			if r.Header.Get("Authorization") == "" {
				r.Header.Add("Authorization", "Bearer "+authToken)
			}
			rwrap.Header().Set("X-Auth-Token", authToken)

			next.ServeHTTP(rwrap, r)

			// Notify the client that the session was auto-refreshed
			if didAutoRefresh {
				w.Header().Add("GEX-Session-Auto-Refreshed", "true")
			}

			w.Write(rwrap.body.Bytes())
		})
	}
}

func writeError(w http.ResponseWriter, err error) {
	resp := map[string]string{
		"error": err.Error(),
	}
	json.NewEncoder(w).Encode(resp)

	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Set("Content-Type", "application/json")
}

func getJwtFromRequest(r *http.Request, jwtToolkit *jwtutil.Toolkit) (*jwt.Token, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("no Authorization header found")
	}

	// log(">> JwtMiddleware: found JWT token in Authorization header")
	jwtToken, err := jwtToolkit.GetAuthorizationHeaderJwt(authHeader)
	if err != nil {
		log(">> JwtMiddleware: error converting Authorization header to JWT:", err)
		return nil, ErrMalformedJwt
	}

	return jwtToken, err
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
