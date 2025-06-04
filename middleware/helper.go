package middleware

import (
	"bytes"
	"net/http"
)

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
	// w.ResponseWriter.WriteHeader(statusCode)
}
