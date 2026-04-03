package usecase

import (
	"context"
	"time"

	"github.com/example/jwt-ddd-clean/internal/application/dto"
	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/service"
)

// TokenUsecase defines the token usecase interface
type TokenUsecase interface {
	GenerateTokens(ctx context.Context, user *model.User) (*dto.TokenResponse, error)
	ValidateToken(ctx context.Context, token string) (*dto.TokenValidationResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error)
	RevokeToken(ctx context.Context, token string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
}

type tokenUsecase struct {
	tokenService *service.TokenService
}

// NewTokenUsecase creates a new TokenUsecase
func NewTokenUsecase(tokenService *service.TokenService) TokenUsecase {
	return &tokenUsecase{
		tokenService: tokenService,
	}
}

func (u *tokenUsecase) GenerateTokens(ctx context.Context, user *model.User) (*dto.TokenResponse, error) {
	tokenPair, err := u.tokenService.GenerateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	expiresIn := int64(time.Until(tokenPair.Access.ExpiresAt).Seconds())

	return &dto.TokenResponse{
		Token:     tokenPair.Access.AccessToken,
		TokenType: "Bearer",
		ExpiresIn: expiresIn,
	}, nil
}

func (u *tokenUsecase) ValidateToken(ctx context.Context, token string) (*dto.TokenValidationResponse, error) {
	claims, err := u.tokenService.ValidateToken(ctx, token)
	if err != nil {
		return &dto.TokenValidationResponse{
			Valid: false,
		}, nil
	}

	return &dto.TokenValidationResponse{
		Valid:    true,
		UserID:   claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}

func (u *tokenUsecase) RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error) {
	tokenPair, err := u.tokenService.RefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	expiresIn := int64(time.Until(tokenPair.Access.ExpiresAt).Seconds())

	return &dto.TokenResponse{
		Token:     tokenPair.Access.AccessToken,
		TokenType: "Bearer",
		ExpiresIn: expiresIn,
	}, nil
}

func (u *tokenUsecase) RevokeToken(ctx context.Context, token string) error {
	return u.tokenService.RevokeToken(ctx, token)
}

func (u *tokenUsecase) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return u.tokenService.RevokeAllUserTokens(ctx, userID)
}
