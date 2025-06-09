package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/i247app/gex/session"
)

func addSessionToRequestContext(r *http.Request, key session.SessionRequestContextKey, sess session.SessionStorer) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, sess))
}

func writeError(w http.ResponseWriter, tag string, origin string, err error) {
	resp := map[string]string{
		"error":  "gex panic: " + err.Error(),
		"tag":    tag,
		"origin": origin,
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// responseWriterWrapper wraps the http.ResponseWriter to capture the response body.
type responseWriterWrapper struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int // Add field to store status code
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *responseWriterWrapper) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Implement WriteHeader to capture the status code
func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}
