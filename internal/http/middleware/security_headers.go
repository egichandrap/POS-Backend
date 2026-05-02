package http

import (
	"log"
	"net/http"
)

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[SecurityHeaders] Request: %s %s", r.Method, r.URL.Path)

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (strict)
		w.Header().Set(
			"Content-Security-Policy",
			"default-src 'none'; frame-ancestors 'none'; base-uri 'none';",
		)

		// Permissions Policy (restrict browser features)
		w.Header().Set(
			"Permissions-Policy",
			"camera=(), microphone=(), geolocation=(), payment=()",
		)

		// Strict Transport Security (HSTS)
		w.Header().Set(
			"Strict-Transport-Security",
			"max-age=63072000; includeSubDomains; preload",
		)

		// Remove server header
		w.Header().Del("Server")

		next.ServeHTTP(w, r)
	})
}
