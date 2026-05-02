package http

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{}, // Empty = deny all by default
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           3600,
	}
}

// CORSMiddleware handles CORS preflight and requests
func CORSMiddleware(config CORSConfig) func(http.Handler) http.Handler {
	// Normalize origins to lowercase
	for i, origin := range config.AllowedOrigins {
		config.AllowedOrigins[i] = strings.ToLower(origin)
	}

	log.Printf("[CORS] Middleware initialized with allowed origins: %v", config.AllowedOrigins)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			originLower := strings.ToLower(origin)

			log.Printf("[CORS] Request: %s %s, Origin: %q", r.Method, r.URL.Path, origin)

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range config.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == originLower {
					allowed = true
					break
				}
			}

			// If origin not in allowed list, deny request
			if origin != "" && !allowed {
				log.Printf("[CORS] Origin NOT allowed: %q", origin)
				http.Error(w, "CORS: Origin not allowed", http.StatusForbidden)
				return
			}

			// Set CORS headers
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				// If no Origin header but "*" is allowed, set wildcard
				for _, allowedOrigin := range config.AllowedOrigins {
					if allowedOrigin == "*" {
						w.Header().Set("Access-Control-Allow-Origin", "*")
						break
					}
				}
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ","))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ","))
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ","))
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))

			log.Printf("[CORS] Headers set, Allow-Origin: %q", w.Header().Get("Access-Control-Allow-Origin"))

			// Handle preflight
			if r.Method == "OPTIONS" {
				log.Printf("[CORS] Preflight request, returning 204")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
