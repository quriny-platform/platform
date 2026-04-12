package middleware

import "net/http"

// CORS adds Cross-Origin Resource Sharing headers to every response. This is
// required so the frontend dev server (typically running on a different port
// like :5173) can call the backend API without browser security errors.
//
// For development, we allow all origins. In production this should be scoped
// to specific domains.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from any origin during development.
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Allow common HTTP methods used by CRUD operations.
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// Allow common headers sent by frontend clients.
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS requests — browsers send these before the
		// actual request to check if CORS is allowed.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
