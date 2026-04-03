package handler

import (
	"context"

	"github.com/example/jwt-ddd-clean/internal/application/dto"
	"github.com/example/jwt-ddd-clean/internal/application/usecase"
)

// TokenHandler handles token operations (non-HTTP layer)
type TokenHandler struct {
	tokenUsecase usecase.TokenUsecase
}

// NewTokenHandler creates a new TokenHandler
func NewTokenHandler(tokenUsecase usecase.TokenUsecase) *TokenHandler {
	return &TokenHandler{
		tokenUsecase: tokenUsecase,
	}
}

// RefreshToken handles token refresh requests
func (h *TokenHandler) RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error) {
	return h.tokenUsecase.RefreshToken(ctx, refreshToken)
}

// ValidateToken handles token validation requests
func (h *TokenHandler) ValidateToken(ctx context.Context, token string) (*dto.TokenValidationResponse, error) {
	return h.tokenUsecase.ValidateToken(ctx, token)
}

// RevokeToken handles token revocation requests
func (h *TokenHandler) RevokeToken(ctx context.Context, token string) error {
	return h.tokenUsecase.RevokeToken(ctx, token)
}
