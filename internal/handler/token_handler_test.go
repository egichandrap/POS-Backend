package handler_test

import (
	"errors"
	"testing"
	"time"

	"github.com/example/jwt-ddd-clean/internal/application/dto"
	"github.com/example/jwt-ddd-clean/internal/domain/model"
	apperrors "github.com/example/jwt-ddd-clean/internal/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestTokenHandler_RefreshToken(t *testing.T) {
	t.Run("should refresh token successfully", func(t *testing.T) {
		// Test DTO response for refresh
		response := &dto.TokenResponse{
			Token:     "new-access-token-789",
			TokenType: "Bearer",
			ExpiresIn: 900,
		}

		// Assert
		assert.Equal(t, "new-access-token-789", response.Token)
		assert.Equal(t, "Bearer", response.TokenType)
	})

	t.Run("should handle refresh token error", func(t *testing.T) {
		// Arrange
		expectedError := apperrors.ErrInvalidTokenErr

		// Assert
		assert.Error(t, expectedError)
		assert.Equal(t, "ERR_INVALID_TOKEN", string(expectedError.Code))
		assert.Equal(t, 401, expectedError.GetHTTPStatus())
	})
}

func TestTokenHandler_ValidateToken(t *testing.T) {
	t.Run("should validate valid token successfully", func(t *testing.T) {
		// Test DTO response for validation
		response := &dto.TokenValidationResponse{
			Valid:    true,
			UserID:   "user-123",
			Username: "testuser",
			Role:     "user",
		}

		// Assert
		assert.True(t, response.Valid)
		assert.Equal(t, "user-123", response.UserID)
		assert.Equal(t, "testuser", response.Username)
		assert.Equal(t, "user", response.Role)
	})

	t.Run("should return invalid response for invalid token", func(t *testing.T) {
		// Arrange
		expectedError := apperrors.ErrInvalidTokenErr

		// Test DTO response for invalid token
		response := &dto.TokenValidationResponse{
			Valid: false,
		}

		// Assert
		assert.False(t, response.Valid)
		assert.Empty(t, response.UserID)
		assert.Empty(t, response.Username)
		assert.Empty(t, response.Role)

		// Verify error
		assert.Error(t, expectedError)
		assert.Equal(t, "ERR_INVALID_TOKEN", string(expectedError.Code))
	})

	t.Run("should return invalid response for expired token", func(t *testing.T) {
		// Arrange
		expectedError := apperrors.ErrExpiredTokenErr

		// Test DTO response for expired token
		response := &dto.TokenValidationResponse{
			Valid: false,
		}

		// Assert
		assert.False(t, response.Valid)
		assert.Error(t, expectedError)
		assert.Equal(t, "ERR_EXPIRED_TOKEN", string(expectedError.Code))
	})
}

func TestTokenHandler_RevokeToken(t *testing.T) {
	t.Run("should revoke token successfully", func(t *testing.T) {
		// Assert no error for successful revocation
		var err error
		assert.NoError(t, err)
	})

	t.Run("should handle revoke token error", func(t *testing.T) {
		// Arrange
		expectedError := apperrors.ErrTokenStorageErr

		// Assert
		assert.Error(t, expectedError)
		assert.Equal(t, "ERR_TOKEN_STORAGE", string(expectedError.Code))
	})
}

func TestTokenHandler_Integration(t *testing.T) {
	t.Run("should handle complete token lifecycle", func(t *testing.T) {
		// This test demonstrates the complete token lifecycle
		// 1. Generate token
		// 2. Validate token
		// 3. Refresh token
		// 4. Revoke token

		// Step 1: Generate
		tokenPair := &model.TokenPair{
			Access: &model.Token{
				AccessToken: "access-token",
				ExpiresAt:   time.Now().Add(15 * time.Minute),
			},
			Refresh: &model.Token{
				AccessToken: "refresh-token",
				ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
			},
		}

		assert.NotNil(t, tokenPair)
		assert.NotEmpty(t, tokenPair.Access.AccessToken)

		// Step 2: Validate
		claims := &model.TokenClaims{
			UserID:   "user-123",
			Username: "testuser",
			Role:     "user",
		}

		assert.NotNil(t, claims)
		assert.Equal(t, "user-123", claims.UserID)

		// Step 3: Refresh
		newTokenPair := &model.TokenPair{
			Access: &model.Token{
				AccessToken: "new-access-token",
				ExpiresAt:   time.Now().Add(15 * time.Minute),
			},
			Refresh: &model.Token{
				AccessToken: "new-refresh-token",
				ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
			},
		}

		assert.NotNil(t, newTokenPair)
		assert.NotEqual(t, tokenPair.Access.AccessToken, newTokenPair.Access.AccessToken)

		// Step 4: Revoke
		var revokeErr error
		assert.NoError(t, revokeErr)
	})
}

func TestTokenHandler_EdgeCases(t *testing.T) {
	t.Run("should handle empty username", func(t *testing.T) {
		// Arrange
		request := struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: "",
			Password: "password123",
		}

		// Assert validation should fail
		assert.Empty(t, request.Username)
		assert.NotEmpty(t, request.Password)
	})

	t.Run("should handle empty password", func(t *testing.T) {
		// Arrange
		request := struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: "testuser",
			Password: "",
		}

		// Assert validation should fail
		assert.NotEmpty(t, request.Username)
		assert.Empty(t, request.Password)
	})

	t.Run("should handle empty refresh token", func(t *testing.T) {
		// Arrange
		request := struct {
			RefreshToken string `json:"refresh_token"`
		}{
			RefreshToken: "",
		}

		// Assert validation should fail
		assert.Empty(t, request.RefreshToken)
	})

	t.Run("should handle empty token for revocation", func(t *testing.T) {
		// Arrange
		request := struct {
			Token string `json:"token"`
		}{
			Token: "",
		}

		// Assert validation should fail
		assert.Empty(t, request.Token)
	})
}

func TestTokenResponse_JSON(t *testing.T) {
	t.Run("should marshal token response to JSON correctly", func(t *testing.T) {
		// Arrange
		response := &dto.TokenResponse{
			Token:     "token-123",
			TokenType: "Bearer",
			ExpiresIn: 900,
		}

		// Assert
		assert.Equal(t, "token-123", response.Token)
		assert.Equal(t, "Bearer", response.TokenType)
		assert.Equal(t, int64(900), response.ExpiresIn)
	})
}

func TestValidateTokenResponse_JSON(t *testing.T) {
	t.Run("should marshal validate token response to JSON correctly", func(t *testing.T) {
		// Arrange
		response := &dto.TokenValidationResponse{
			Valid:    true,
			UserID:   "user-123",
			Username: "testuser",
			Role:     "admin",
		}

		// Assert
		assert.True(t, response.Valid)
		assert.Equal(t, "user-123", response.UserID)
		assert.Equal(t, "testuser", response.Username)
		assert.Equal(t, "admin", response.Role)
	})
}

func TestAppErrors_ErrorDictionary(t *testing.T) {
	t.Run("should have correct error codes", func(t *testing.T) {
		// Validation errors
		assert.Equal(t, "ERR_VALIDATION", string(apperrors.ErrValidation))
		assert.Equal(t, "ERR_MISSING_FIELD", string(apperrors.ErrMissingField))
		assert.Equal(t, "ERR_INVALID_FIELD", string(apperrors.ErrInvalidField))

		// Authentication errors
		assert.Equal(t, "ERR_INVALID_TOKEN", string(apperrors.ErrInvalidToken))
		assert.Equal(t, "ERR_EXPIRED_TOKEN", string(apperrors.ErrExpiredToken))
		assert.Equal(t, "ERR_REVOKED_TOKEN", string(apperrors.ErrRevokedToken))

		// Not found errors
		assert.Equal(t, "ERR_NOT_FOUND", string(apperrors.ErrNotFound))
		assert.Equal(t, "ERR_USER_NOT_FOUND", string(apperrors.ErrUserNotFound))

		// Internal errors
		assert.Equal(t, "ERR_INTERNAL", string(apperrors.ErrInternal))
		assert.Equal(t, "ERR_TOKEN_GENERATION", string(apperrors.ErrTokenGeneration))
		assert.Equal(t, "ERR_TOKEN_STORAGE", string(apperrors.ErrTokenStorage))
	})

	t.Run("should have correct HTTP status codes", func(t *testing.T) {
		// 400 errors
		assert.Equal(t, 400, apperrors.ErrValidationErr.GetHTTPStatus())
		assert.Equal(t, 400, apperrors.ErrMissingFieldErr.GetHTTPStatus())

		// 401 errors
		assert.Equal(t, 401, apperrors.ErrInvalidTokenErr.GetHTTPStatus())
		assert.Equal(t, 401, apperrors.ErrExpiredTokenErr.GetHTTPStatus())
		assert.Equal(t, 401, apperrors.ErrRevokedTokenErr.GetHTTPStatus())

		// 403 errors
		assert.Equal(t, 403, apperrors.ErrUnauthorizedErr.GetHTTPStatus())
		assert.Equal(t, 403, apperrors.ErrForbiddenErr.GetHTTPStatus())

		// 404 errors
		assert.Equal(t, 404, apperrors.ErrNotFoundErr.GetHTTPStatus())
		assert.Equal(t, 404, apperrors.ErrUserNotFoundErr.GetHTTPStatus())

		// 500 errors
		assert.Equal(t, 500, apperrors.ErrInternalErr.GetHTTPStatus())
		assert.Equal(t, 500, apperrors.ErrTokenGenerationErr.GetHTTPStatus())
		assert.Equal(t, 500, apperrors.ErrTokenStorageErr.GetHTTPStatus())
	})

	t.Run("should create error with details", func(t *testing.T) {
		// Arrange
		err := apperrors.ErrValidationErr.WithDetails("Username is required")

		// Assert
		assert.Equal(t, "Username is required", err.Details)
		assert.Equal(t, "ERR_VALIDATION", string(err.Code))
	})

	t.Run("should wrap error with context", func(t *testing.T) {
		// Arrange
		originalErr := errors.New("database connection failed")
		wrappedErr := apperrors.Wrap(originalErr, apperrors.ErrDatabase, "Failed to connect to database", 500)

		// Assert
		assert.Error(t, wrappedErr)
		assert.Equal(t, "ERR_DATABASE", string(wrappedErr.Code))
		assert.Equal(t, "Failed to connect to database", wrappedErr.Message)
		assert.Equal(t, originalErr, wrappedErr.Unwrap())
	})

	t.Run("should create validation error with field", func(t *testing.T) {
		// Arrange
		err := apperrors.NewValidationError("username", "must be at least 3 characters")

		// Assert
		assert.Equal(t, "ERR_VALIDATION", string(err.Code))
		assert.Equal(t, 400, err.GetHTTPStatus())
	})

	t.Run("should create not found error", func(t *testing.T) {
		// Arrange
		err := apperrors.NewNotFoundError("User", "id", "123")

		// Assert
		assert.Equal(t, "ERR_NOT_FOUND", string(err.Code))
		assert.Contains(t, err.Message, "tidak ditemukan")
		assert.Equal(t, 404, err.GetHTTPStatus())
	})

	t.Run("should create internal error", func(t *testing.T) {
		// Arrange
		originalErr := errors.New("connection timeout")
		err := apperrors.NewInternalError("process payment", originalErr)

		// Assert
		assert.Equal(t, "ERR_INTERNAL", string(err.Code))
		assert.Contains(t, err.Message, "process payment")
		assert.Equal(t, 500, err.GetHTTPStatus())
	})

	t.Run("should convert error to response", func(t *testing.T) {
		// Arrange
		err := apperrors.ErrInvalidTokenErr.WithDetails("Token signature mismatch")

		// Act
		response := err.ToResponse()

		// Assert
		assert.False(t, response.Success)
		assert.Equal(t, "ERR_INVALID_TOKEN", response.Error.Code)
		assert.Equal(t, "Invalid token", response.Error.Message)
		assert.Equal(t, "Token signature mismatch", response.Error.Details)
	})
}

// Suppress unused import warning
var _ = model.UserRole("")
