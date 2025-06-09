package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/i247app/gex/jwtutil"
	"github.com/i247app/gex/session"
)

// XwtMiddleware is a middleware that handles XWT token authentication and session management.
// It behaves similarly to JwtMiddleware, but uses the signed JWT token string as the session key.
func XwtMiddleware(
	sessionContainer *session.Container,
	jwtToolkit *jwtutil.Toolkit,
	sessionFactory SessionFactory,
	sessionTTL time.Duration,
	sessionContextKey session.SessionRequestContextKey,
) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				didAutoRefresh bool = false
			)

			// Skip entire token and session handling if this header is set
			if r.Header.Get("X-Skip-Session") == "true" {
				next.ServeHTTP(w, r)
				return
			}

			// 1. Get or create a XWT token
			xwtResult, err := getOrCreateXwtToken(r, jwtToolkit)
			var isValidXwt bool = err == nil && xwtResult.XwtToken != "" && xwtResult.SessionKey != ""
			if !isValidXwt {
				writeError(w, "error getting or creating xwt", "xwt_middleware", fmt.Errorf("Unauthorized"))
				return
			}
			authToken := xwtResult.XwtToken
			sessionKey := authToken

			// 2. Get session
			sess, ok := getSessionFromSessionKey(sessionContainer, sessionKey)
			if sess == nil || !ok {
				tmp, err := initNewSession(sessionKey, authToken, "gex.xwt_middleware", sessionContainer, sessionFactory, sessionTTL)
				if tmp == nil || err != nil {
					log(">> XwtMiddleware: error initializing new session")
					writeError(w, "error initializing new session", "xwt_middleware", fmt.Errorf("Unauthorized"))
					return
				}
				sess = tmp
			}

			// 3. Check for expired session
			isSessionExpired, err := isSessionExpired(sess)
			if isSessionExpired || err != nil {
				didAutoRefresh = true

				log(">> XwtMiddleware: session expired, for now just auto-refreshing...")
				sess, _ = refreshSession(sess, sessionTTL)
				if sess == nil {
					log(">> XwtMiddleware: error refreshing expired session")
					writeError(w, "error refreshing expired session", "xwt_middleware", fmt.Errorf("Unauthorized"))
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

			// Add session to request context
			r = addSessionToRequestContext(r, sessionContextKey, sess)

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

type XwtResult struct {
	XwtToken   string
	SessionKey string
}

func getOrCreateXwtToken(r *http.Request, jwtToolkit *jwtutil.Toolkit) (*XwtResult, error) {
	xwtResult, err := getValidXwtFromRequest(r)
	if xwtResult != nil && err == nil {
		return xwtResult, nil
	}

	// Failed to get a valid XWT token from the request
	if xwtResult == nil || xwtResult.XwtToken == "" || xwtResult.SessionKey == "" || err != nil {
		log(">> XwtMiddleware: WARNING ignoring your xwt token - error getting XWT from request:", err)
	} else {
		log(">> XwtMiddleware: xwt ok")
	}

	// Create a new XWT token with a new session key
	xwtResult, err = createNewXwtToken(jwtToolkit)
	if err != nil {
		return nil, fmt.Errorf("error creating new XWT token: %v", err)
	}

	return xwtResult, nil
}

func getValidXwtFromRequest(r *http.Request) (*XwtResult, error) {
	// Validate Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("no Authorization header found")
	}

	if authHeader == "Bearer " {
		return nil, fmt.Errorf("no XWT token found in Authorization header")
	}

	// Get XWT Token
	xwtToken := strings.TrimPrefix(authHeader, "Bearer ")
	sessionKey := xwtToken

	return &XwtResult{
		XwtToken:   xwtToken,
		SessionKey: sessionKey,
	}, nil
}

func createNewXwtToken(jwtToolkit *jwtutil.Toolkit) (*XwtResult, error) {
	sessionKey := "n/a" // util.GenerateSessionKey()
	claims := jwtutil.NewClaims(sessionKey)
	jwtToken, err := jwtToolkit.GenerateJwt(claims)
	if err != nil {
		return nil, fmt.Errorf("error creating new XWT token: %v", err)
	}

	signedToken, err := jwtToolkit.SignToken(jwtToken)
	if err != nil {
		return nil, fmt.Errorf("error signing new XWT token: %v", err)
	}

	return &XwtResult{
		XwtToken:   signedToken,
		SessionKey: signedToken,
	}, nil
}
