package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// statusWriter wraps http.ResponseWriter to capture the status code written
// by the downstream handler. This lets the logger report the response status
// without interfering with the response itself.
type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before delegating to the real writer.
func (sw *statusWriter) WriteHeader(code int) {
	sw.statusCode = code
	sw.ResponseWriter.WriteHeader(code)
}

// RequestLogger logs every HTTP request with structured fields:
//   - method:   the HTTP method (GET, POST, etc.)
//   - path:     the request URL path
//   - status:   the response status code
//   - duration: how long the request took to handle
//
// This uses Go 1.21's slog package for structured, levelled logging.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the writer so we can capture the status code.
		sw := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Pass the request down the middleware chain.
		next.ServeHTTP(sw, r)

		// Log the completed request with structured fields.
		slog.Info("request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", sw.statusCode),
			slog.Duration("duration", time.Since(start)),
		)
	})
}
