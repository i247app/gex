package middleware

import (
	"bytes"
	"net/http"
)

// responseWriterWrapper wraps the http.ResponseWriter to capture the response body.
type responseWriterWrapper struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	return w.body.Write(b)
}
