package service

import (
	"context"
	"errors"
	"time"

	apperrors "github.com/example/jwt-ddd-clean/internal/pkg/errors"
	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
)

// TokenService handles business logic for token operations
type TokenService struct {
	tokenRepo       repository.TokenRepository
	jwtProvider     JWTProvider
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewTokenService creates a new TokenService
func NewTokenService(
	tokenRepo repository.TokenRepository,
	jwtProvider JWTProvider,
	accessTokenTTL,
	refreshTokenTTL time.Duration,
) *TokenService {
	return &TokenService{
		tokenRepo:       tokenRepo,
		jwtProvider:     jwtProvider,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// GenerateTokens generates a new access and refresh token pair for a user
func (s *TokenService) GenerateTokens(ctx context.Context, user *model.User) (*model.TokenPair, error) {
	now := time.Now()

	claims := &model.TokenClaims{
		UserID:   user.ID(),
		Username: user.Username(),
		Role:     string(user.Role()),
	}

	// Generate access token
	accessExpiresAt := now.Add(s.accessTokenTTL)
	accessToken, err := s.jwtProvider.GenerateToken(claims, accessExpiresAt)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrTokenGeneration, "Failed to generate access token", apperrors.ErrTokenGenerationErr.GetHTTPStatus())
	}

	// Generate refresh token
	refreshExpiresAt := now.Add(s.refreshTokenTTL)
	refreshToken, err := s.jwtProvider.GenerateToken(claims, refreshExpiresAt)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrTokenGeneration, "Failed to generate refresh token", apperrors.ErrTokenGenerationErr.GetHTTPStatus())
	}

	// Store refresh token for later use
	if err := s.tokenRepo.Store(ctx, user.ID(), refreshToken, "refresh", refreshExpiresAt); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrTokenStorage, "Failed to store refresh token", apperrors.ErrTokenStorageErr.GetHTTPStatus())
	}

	return &model.TokenPair{
		Access: &model.Token{
			AccessToken: accessToken,
			ExpiresAt:   accessExpiresAt,
		},
		Refresh: &model.Token{
			AccessToken: refreshToken,
			ExpiresAt:   refreshExpiresAt,
		},
	}, nil
}

// ValidateToken validates a token and returns its claims
func (s *TokenService) ValidateToken(ctx context.Context, token string) (*model.TokenClaims, error) {
	// Check if token is blacklisted
	isBlacklisted, err := s.tokenRepo.IsBlacklisted(ctx, token)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "Failed to check token blacklist status", apperrors.ErrInternalErr.GetHTTPStatus())
	}
	if isBlacklisted {
		return nil, apperrors.ErrRevokedTokenErr
	}

	// Validate JWT signature and expiration
	claims, err := s.jwtProvider.ValidateToken(token)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == apperrors.ErrExpiredToken {
				return nil, apperrors.ErrExpiredTokenErr
			}
			if appErr.Code == apperrors.ErrInvalidToken {
				return nil, apperrors.ErrInvalidTokenErr
			}
		}
		return nil, apperrors.ErrInvalidTokenErr
	}

	return claims, nil
}

// RefreshToken generates a new token pair using a refresh token
func (s *TokenService) RefreshToken(ctx context.Context, refreshToken string) (*model.TokenPair, error) {
	// Validate the refresh token
	claims, err := s.ValidateToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	// Blacklist the old refresh token
	expiration, _ := s.jwtProvider.GetExpiration(refreshToken)
	if err := s.tokenRepo.Blacklist(ctx, refreshToken, expiration); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrTokenStorage, "Failed to blacklist old refresh token", apperrors.ErrTokenStorageErr.GetHTTPStatus())
	}

	// Generate new token pair using reconstructed user
	reconstructedUser := model.ReconstructUser(
		claims.UserID,
		claims.Username,
		"",
		"",
		"",
		model.UserRole(claims.Role),
		model.StatusActive,
		time.Now(),
		time.Now(),
		nil,
	)
	return s.GenerateTokens(ctx, reconstructedUser)
}

// RevokeToken revokes (blacklists) a token
func (s *TokenService) RevokeToken(ctx context.Context, token string) error {
	expiration, err := s.jwtProvider.GetExpiration(token)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrInvalidToken, "Failed to get token expiration", apperrors.ErrInvalidTokenErr.GetHTTPStatus())
	}

	if err := s.tokenRepo.Blacklist(ctx, token, expiration); err != nil {
		return apperrors.Wrap(err, apperrors.ErrTokenStorage, "Failed to blacklist token", apperrors.ErrTokenStorageErr.GetHTTPStatus())
	}

	return nil
}

// RevokeAllUserTokens revokes all tokens for a specific user
func (s *TokenService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	if err := s.tokenRepo.Delete(ctx, userID, "refresh"); err != nil {
		return apperrors.Wrap(err, apperrors.ErrTokenStorage, "Failed to delete user tokens", apperrors.ErrTokenStorageErr.GetHTTPStatus())
	}
	return nil
}
