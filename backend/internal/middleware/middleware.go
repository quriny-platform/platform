// Package middleware provides reusable HTTP middleware for the Quriny runtime
// server. Middleware functions follow the standard Go pattern: each takes an
// http.Handler and returns a new http.Handler that wraps additional behaviour
// around the original.
package middleware

import "net/http"

// Middleware is a function that wraps an http.Handler with additional behaviour.
type Middleware func(http.Handler) http.Handler

// Chain applies a sequence of middleware to a handler. Middleware are applied
// in reverse order so the first middleware in the list is the outermost wrapper.
//
// Example:
//
//	Chain(handler, Logger, CORS)
//	// Request flow: CORS → Logger → handler → Logger → CORS
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	// Apply in reverse so the first middleware listed wraps outermost.
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
