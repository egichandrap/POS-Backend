package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	apperrors "github.com/example/jwt-ddd-clean/internal/pkg/errors"
	"github.com/example/jwt-ddd-clean/internal/domain/service"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
	// UsernameKey is the context key for username
	UsernameKey contextKey = "username"
	// UserRoleKey is the context key for user role
	UserRoleKey contextKey = "user_role"
	// TenantIDKey is the context key for tenant ID
	TenantIDKey contextKey = "tenant_id"
	// TenantSlugKey is the context key for tenant slug
	TenantSlugKey contextKey = "tenant_slug"
)

// AuthMiddleware represents JWT authentication middleware
type AuthMiddleware struct {
	tokenService *service.TokenService
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(tokenService *service.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// Authenticate is a middleware that validates JWT tokens
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token := extractTokenFromHeader(r)
		if token == "" {
			m.sendAuthError(w, apperrors.ErrUnauthenticatedErr.WithDetails("Missing or empty Authorization header"))
			return
		}

		// Validate the token
		claims, err := m.tokenService.ValidateToken(r.Context(), token)
		if err != nil {
			var appErr *apperrors.AppError
			if errors.As(err, &appErr) {
				switch appErr.Code {
				case apperrors.ErrExpiredToken:
					m.sendAuthError(w, apperrors.ErrExpiredTokenErr)
					return
				case apperrors.ErrRevokedToken:
					m.sendAuthError(w, apperrors.ErrRevokedTokenErr)
					return
				default:
					m.sendAuthError(w, apperrors.ErrInvalidTokenErr)
					return
				}
			}
			m.sendAuthError(w, apperrors.ErrInvalidTokenErr)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UsernameKey, claims.Username)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole creates a middleware that requires a specific role
func (m *AuthMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user role from context (set by Authenticate middleware)
			userRole, ok := r.Context().Value(UserRoleKey).(string)
			if !ok || userRole == "" {
				m.sendAuthError(w, apperrors.ErrUnauthenticatedErr)
				return
			}

			// Check if user has required role
			hasRole := false
			for _, role := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				m.sendAuthError(w, apperrors.ErrForbiddenErr.WithDetails("Required role: "+strings.Join(roles, " or ")))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractTokenFromHeader extracts the JWT token from the Authorization header
func extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Expect "Bearer <token>" format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// sendAuthError sends an authentication error response
func (m *AuthMiddleware) sendAuthError(w http.ResponseWriter, err *apperrors.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.GetHTTPStatus())
	json.NewEncoder(w).Encode(err.ToResponse())
}

// GetUserFromContext retrieves user information from the context
func GetUserFromContext(ctx context.Context) (userID, username, role string, ok bool) {
	userID, ok = ctx.Value(UserIDKey).(string)
	if !ok {
		return "", "", "", false
	}

	username, ok = ctx.Value(UsernameKey).(string)
	if !ok {
		return "", "", "", false
	}

	role, ok = ctx.Value(UserRoleKey).(string)
	if !ok {
		return "", "", "", false
	}

	return userID, username, role, true
}

// CanManageInventory checks if the user can manage inventory based on role
func CanManageInventory(ctx context.Context) bool {
	_, _, role, ok := GetUserFromContext(ctx)
	if !ok {
		return false
	}
	return role == "SUPER_ADMIN" || role == "ADMIN"
}

// CanAccessPOS checks if the user can access POS features
func CanAccessPOS(ctx context.Context) bool {
	_, _, role, ok := GetUserFromContext(ctx)
	if !ok {
		return false
	}
	return role == "SUPER_ADMIN" || role == "ADMIN" || role == "CASHIER"
}

// CanManageUsers checks if the user can manage other users
func CanManageUsers(ctx context.Context) bool {
	_, _, role, ok := GetUserFromContext(ctx)
	if !ok {
		return false
	}
	return role == "SUPER_ADMIN"
}

// OptionalAuth is a middleware that optionally authenticates if a token is provided
// but doesn't require it. Useful for endpoints that behave differently for authenticated users.
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractTokenFromHeader(r)
		if token != "" {
			claims, err := m.tokenService.ValidateToken(r.Context(), token)
			if err == nil {
				// Token is valid, add to context
				ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
				ctx = context.WithValue(ctx, UsernameKey, claims.Username)
				ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
				r = r.WithContext(ctx)
			}
			// If token is invalid, continue without authentication
		}

		next.ServeHTTP(w, r)
	})
}
