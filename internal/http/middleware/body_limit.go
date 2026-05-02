package http

import (
	"net/http"
)

const (
	// DefaultMaxBodySize is the default maximum request body size (1MB)
	DefaultMaxBodySize = 1 << 20 // 1MB
)

// MaxBodySizeMiddleware limits request body size
func MaxBodySizeMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBodySize
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip body size check for OPTIONS requests (CORS preflight)
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			// Use MaxBytesHandler for other requests
			http.MaxBytesHandler(next, maxBytes).ServeHTTP(w, r)
		})
	}
}

// StrictMaxBodyMiddleware for sensitive endpoints (100KB)
func StrictMaxBodyMiddleware() func(http.Handler) http.Handler {
	return MaxBodySizeMiddleware(100 << 10) // 100KB
}

// LoginMaxBodyMiddleware for login endpoint (10KB)
func LoginMaxBodyMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip body size check for OPTIONS requests (CORS preflight)
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			// Use MaxBytesHandler for other requests
			http.MaxBytesHandler(next, 10<<10).ServeHTTP(w, r)
		})
	}
}
