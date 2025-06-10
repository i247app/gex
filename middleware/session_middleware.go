package middleware

import (
	"bytes"
	"net/http"

	"github.com/i247app/gex/session"
)

// SessionMiddleware is a unified middleware that handles session management using a SessionProvider.
// It retrieves the session from the provider, handles auto-refresh notifications,
// and wraps the response writer to capture the response body.
func SessionMiddleware(
	sessionProvider SessionProvider,
	sessionContextKey session.SessionRequestContextKey,
) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get session with metadata from the provider
			sessionResult, err := sessionProvider.GetSessionWithMetadata(r)
			if err != nil {
				writeError(w, "error getting session", "session_middleware", err)
				return
			}

			// If session is nil (e.g., X-Skip-Session header is set), proceed without session
			if sessionResult == nil || sessionResult.Session == nil {
				next.ServeHTTP(w, r)
				return
			}

			sess := sessionResult.Session
			didAutoRefresh := sessionResult.DidAutoRefresh
			authToken := sessionResult.AuthToken

			// Wrap the response writer to capture the response body
			wr := &responseWriterWrapper{
				ResponseWriter: w,
				body:           bytes.NewBuffer(nil),
			}

			// Set X-Auth-Token response header
			if authToken != "" {
				wr.Header().Set("X-Auth-Token", authToken)
			}

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
