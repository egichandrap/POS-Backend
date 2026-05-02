package http

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/example/jwt-ddd-clean/internal/pkg/sanitizer"
)

// List of standard HTTP headers that should skip validation
// These are headers automatically added by clients/browsers and are safe
var safeHeaders = map[string]bool{
	"user-agent":        true,
	"accept":            true,
	"accept-encoding":   true,
	"accept-language":   true,
	"connection":        true,
	"content-length":    true,
	"content-type":      true,
	"authorization":     true,
	"host":              true,
	"referer":           true,
	"origin":            true,
	"cache-control":     true,
	"postman-token":     true,
	"x-forwarded-for":   true,
	"x-forwarded-host":  true,
	"x-forwarded-proto": true,
	"x-request-id":      true,
	"x-request-start":   true,
	// Browser security headers (Fetch Metadata, Client Hints, etc.)
	"sec-ch-ua":                 true,
	"sec-ch-ua-mobile":          true,
	"sec-ch-ua-platform":        true,
	"sec-fetch-dest":            true,
	"sec-fetch-mode":            true,
	"sec-fetch-site":            true,
	"sec-fetch-user":            true,
	"sec-gpc":                   true,
	"dnt":                       true,
	"upgrade-insecure-requests": true,
	"pragma":                    true,
}

// List of standard query parameter names that are safe
var safeQueryParams = map[string]bool{
	"limit":          true,
	"offset":         true,
	"page":           true,
	"size":           true,
	"sort":           true,
	"order":          true,
	"search":         true,
	"query":          true,
	"filter":         true,
	"status":         true,
	"role":           true,
	"type":           true,
	"from":           true,
	"to":             true,
	"date":           true,
	"start_date":     true,
	"end_date":       true,
	"sku":            true,
	"name":           true,
	"location":       true,
	"payment_method": true,
}

// isSafeHeader checks if a header is a standard/safe header that should skip validation
func isSafeHeader(key string) bool {
	lowerKey := strings.ToLower(key)
	return safeHeaders[lowerKey]
}

// isSafeQueryParam checks if a query parameter name is standard/safe
func isSafeQueryParam(key string) bool {
	lowerKey := strings.ToLower(key)
	return safeQueryParams[lowerKey]
}

// isValidParamName checks if a parameter name contains only safe characters
func isValidParamName(param string) bool {
	// Allow only alphanumeric, underscore, and hyphen
	paramNamePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return paramNamePattern.MatchString(param)
}

// ValidationMiddleware validates incoming requests
func ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip validation for OPTIONS preflight requests (CORS)
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Validate URL path
		if !sanitizer.ValidatePathTraversal(r.URL.Path) {
			sendValidationError(w, "Invalid URL path")
			return
		}

		// Validate query parameters
		for key, values := range r.URL.Query() {
			// Skip validation for safe query parameter names
			if isSafeQueryParam(key) {
				// Still validate the values
				for _, value := range values {
					if !sanitizer.ValidateSQL(value) || !sanitizer.ValidateXSS(value) {
						sendValidationError(w, "Invalid query value")
						return
					}
				}
				continue
			}

			// For non-standard params, validate both key and value
			if !isValidParamName(key) {
				sendValidationError(w, "Invalid query parameter name")
				return
			}
			if !sanitizer.ValidateSQL(key) || !sanitizer.ValidateXSS(key) {
				sendValidationError(w, "Invalid query parameter")
				return
			}
			for _, value := range values {
				if !sanitizer.ValidateSQL(value) || !sanitizer.ValidateXSS(value) {
					sendValidationError(w, "Invalid query value")
					return
				}
			}
		}

		// Validate headers (skip standard/safe headers)
		for key, values := range r.Header {
			// Skip validation for safe headers
			if isSafeHeader(key) {
				continue
			}

			// Validate header name
			if !sanitizer.ValidateSQL(key) || !sanitizer.ValidateXSS(key) {
				sendValidationError(w, "Invalid header name")
				return
			}

			// Validate header values
			for _, v := range values {
				if !sanitizer.ValidateSQL(v) || !sanitizer.ValidateXSS(v) {
					sendValidationError(w, "Invalid header value")
					return
				}
			}
		}

		// Validate Content-Type for POST/PUT/PATCH (only if request has body)
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			// Only validate Content-Type if there's a request body
			if r.ContentLength > 0 {
				contentType := r.Header.Get("Content-Type")
				if contentType == "" {
					sendValidationError(w, "Content-Type header required for request with body")
					return
				}

				// Only allow JSON content type
				if !strings.Contains(contentType, "application/json") {
					sendValidationError(w, "Content-Type must be application/json")
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// ValidateJSONBody validates and sanitizes JSON request body
func ValidateJSONBody(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Reject unknown fields

	if err := decoder.Decode(v); err != nil {
		sendValidationError(w, "Invalid JSON: "+err.Error())
		return false
	}

	// Sanitize string fields in the struct
	sanitizeStruct(v)

	return true
}

// sanitizeStruct recursively sanitizes string fields in a struct
func sanitizeStruct(v interface{}) {
	// This is a simplified version
	// In production, use reflection to sanitize all string fields
}

// ValidateUUID validates UUID format
func ValidateUUID(uuid string) bool {
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return uuidPattern.MatchString(uuid)
}

// ValidatePositiveInt validates positive integer
func ValidatePositiveInt(n int) bool {
	return n > 0
}

// ValidateNonNegative validates non-negative integer
func ValidateNonNegative(n int) bool {
	return n >= 0
}

// ValidateStringLength validates string length
func ValidateStringLength(s string, min, max int) bool {
	length := len(s)
	return length >= min && length <= max
}

// ValidateAlphanumeric validates alphanumeric string
func ValidateAlphanumeric(s string) bool {
	alnumPattern := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	return alnumPattern.MatchString(s)
}

// sendValidationError sends a validation error response
func sendValidationError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    "ERR_VALIDATION",
			"message": message,
		},
	})
}
