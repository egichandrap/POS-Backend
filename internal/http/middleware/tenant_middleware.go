package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	apperrors "github.com/example/jwt-ddd-clean/internal/pkg/errors"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
)

// TenantMiddleware represents tenant identification and isolation middleware
type TenantMiddleware struct {
	tenantRepo repository.TenantRepository
}

// NewTenantMiddleware creates a new TenantMiddleware
func NewTenantMiddleware(tenantRepo repository.TenantRepository) *TenantMiddleware {
	return &TenantMiddleware{
		tenantRepo: tenantRepo,
	}
}

// TenantFromHeader extracts tenant from custom header
func (m *TenantMiddleware) TenantFromHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract tenant slug from header
		tenantSlug := r.Header.Get("X-Tenant-Slug")
		if tenantSlug == "" {
			// Try X-Tenant-ID header as fallback
			tenantID := r.Header.Get("X-Tenant-ID")
			if tenantID != "" {
				// Verify tenant exists
				tenant, err := m.tenantRepo.GetByID(r.Context(), tenantID)
				if err == nil && tenant.IsActive() {
					ctx := context.WithValue(r.Context(), TenantIDKey, tenant.ID())
					ctx = context.WithValue(ctx, TenantSlugKey, tenant.CompanySlug())
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			m.sendTenantError(w, apperrors.ErrUnauthenticatedErr.WithDetails("Missing tenant identification"))
			return
		}

		// Get tenant by slug
		tenant, err := m.tenantRepo.GetByCompanySlug(r.Context(), tenantSlug)
		if err != nil {
			m.sendTenantError(w, apperrors.ErrNotFoundErr.WithDetails("Tenant not found"))
			return
		}

		if !tenant.IsActive() {
			m.sendTenantError(w, apperrors.ErrForbiddenErr.WithDetails("Tenant is not active"))
			return
		}

		// Add tenant info to context
		ctx := context.WithValue(r.Context(), TenantIDKey, tenant.ID())
		ctx = context.WithValue(ctx, TenantSlugKey, tenant.CompanySlug())

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TenantFromSubdomain extracts tenant from subdomain
// Example: company.yourdomain.com -> tenant slug: company
func (m *TenantMiddleware) TenantFromSubdomain(baseDomain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := r.Host
			// Remove port if present
			if idx := strings.Index(host, ":"); idx != -1 {
				host = host[:idx]
			}

			// Check if it's a subdomain
			if !strings.HasSuffix(host, "."+baseDomain) {
				// Not a subdomain, maybe using header instead
				m.TenantFromHeader(next).ServeHTTP(w, r)
				return
			}

			// Extract subdomain
			subdomain := strings.TrimSuffix(host, "."+baseDomain)
			if subdomain == "" || subdomain == baseDomain {
				// No subdomain, try header
				m.TenantFromHeader(next).ServeHTTP(w, r)
				return
			}

			// Get tenant by slug
			tenant, err := m.tenantRepo.GetByCompanySlug(r.Context(), subdomain)
			if err != nil {
				m.sendTenantError(w, apperrors.ErrNotFoundErr.WithDetails("Tenant not found"))
				return
			}

			if !tenant.IsActive() {
				m.sendTenantError(w, apperrors.ErrForbiddenErr.WithDetails("Tenant is not active"))
				return
			}

			// Add tenant info to context
			ctx := context.WithValue(r.Context(), TenantIDKey, tenant.ID())
			ctx = context.WithValue(ctx, TenantSlugKey, tenant.CompanySlug())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TenantFromUser extracts tenant from authenticated user
// This should be used after authentication middleware
func (m *TenantMiddleware) TenantFromUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context (set by auth middleware)
		userID, ok := r.Context().Value(UserIDKey).(string)
		if !ok || userID == "" {
			m.sendTenantError(w, apperrors.ErrUnauthenticatedErr.WithDetails("User not authenticated"))
			return
		}

		// Get tenant by user ID
		tenant, err := m.tenantRepo.GetByUserID(r.Context(), userID)
		if err != nil {
			m.sendTenantError(w, apperrors.ErrNotFoundErr.WithDetails("Tenant not found for user"))
			return
		}

		if !tenant.IsActive() {
			m.sendTenantError(w, apperrors.ErrForbiddenErr.WithDetails("Tenant is not active"))
			return
		}

		// Add tenant info to context
		ctx := context.WithValue(r.Context(), TenantIDKey, tenant.ID())
		ctx = context.WithValue(ctx, TenantSlugKey, tenant.CompanySlug())

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireSubscription checks if tenant has an active subscription with required features
func (m *TenantMiddleware) RequireSubscription(features ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get tenant ID from context
			tenantID, ok := r.Context().Value(TenantIDKey).(string)
			if !ok || tenantID == "" {
				m.sendTenantError(w, apperrors.ErrUnauthenticatedErr.WithDetails("Tenant not found in context"))
				return
			}

			// Get tenant
			tenant, err := m.tenantRepo.GetByID(r.Context(), tenantID)
			if err != nil {
				m.sendTenantError(w, apperrors.ErrNotFoundErr.WithDetails("Tenant not found"))
				return
			}

			// Check if subscription is active
			if !tenant.IsSubscriptionActive() && !tenant.IsTrial() {
				m.sendTenantError(w, apperrors.ErrForbiddenErr.WithDetails("Subscription is not active"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalTenant is a middleware that optionally extracts tenant if available
// but doesn't require it. Useful for public endpoints that may behave differently for tenants.
func (m *TenantMiddleware) OptionalTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Try to get tenant from header
		tenantSlug := r.Header.Get("X-Tenant-Slug")
		if tenantSlug != "" {
			if tenant, err := m.tenantRepo.GetByCompanySlug(r.Context(), tenantSlug); err == nil && tenant.IsActive() {
				ctx = context.WithValue(ctx, TenantIDKey, tenant.ID())
				ctx = context.WithValue(ctx, TenantSlugKey, tenant.CompanySlug())
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Try X-Tenant-ID header
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID != "" {
			if tenant, err := m.tenantRepo.GetByID(r.Context(), tenantID); err == nil && tenant.IsActive() {
				ctx = context.WithValue(ctx, TenantIDKey, tenant.ID())
				ctx = context.WithValue(ctx, TenantSlugKey, tenant.CompanySlug())
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTenantFromContext retrieves tenant information from the context
func GetTenantFromContext(ctx context.Context) (tenantID, tenantSlug string, ok bool) {
	tenantID, ok = ctx.Value(TenantIDKey).(string)
	if !ok {
		return "", "", false
	}

	tenantSlug, ok = ctx.Value(TenantSlugKey).(string)
	if !ok {
		tenantSlug = ""
	}

	return tenantID, tenantSlug, true
}

// sendTenantError sends a tenant-related error response
func (m *TenantMiddleware) sendTenantError(w http.ResponseWriter, err *apperrors.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.GetHTTPStatus())
	json.NewEncoder(w).Encode(err.ToResponse())
}
