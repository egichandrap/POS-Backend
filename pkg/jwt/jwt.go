// Package jwt provides JWT token utilities for external use
package jwt

import (
	"context"
	"time"

	"github.com/example/jwt-ddd-clean/internal/application/dto"
	"github.com/example/jwt-ddd-clean/internal/application/usecase"
	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/service"
	"github.com/example/jwt-ddd-clean/internal/handler"
	infrastructurejwt "github.com/example/jwt-ddd-clean/internal/infrastructure/jwt"
	repo "github.com/example/jwt-ddd-clean/internal/infrastructure/repository"
)

// Config holds the JWT configuration
type Config struct {
	SecretKey       string
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// JWT provides JWT token operations
type JWT struct {
	tokenHandler  *handler.TokenHandler
	tokenUsecase  usecase.TokenUsecase
	tokenService  *service.TokenService
	config        Config
}

// New creates a new JWT instance
func New(config Config) *JWT {
	// Infrastructure layer
	jwtProvider := infrastructurejwt.NewProvider(infrastructurejwt.Config{
		SecretKey: config.SecretKey,
		Issuer:    config.Issuer,
		Algorithm: "HS256",
	})

	tokenRepository := repo.NewMemoryTokenRepository()

	// Domain layer
	tokenService := service.NewTokenService(
		tokenRepository,
		jwtProvider,
		config.AccessTokenTTL,
		config.RefreshTokenTTL,
	)

	// Application layer
	tokenUsecase := usecase.NewTokenUsecase(tokenService)

	// Handler layer
	tokenHandler := handler.NewTokenHandler(tokenUsecase)

	return &JWT{
		tokenHandler: tokenHandler,
		tokenUsecase: tokenUsecase,
		tokenService: tokenService,
		config:       config,
	}
}

// GenerateToken generates a new JWT token pair for a user
func (j *JWT) GenerateToken(user *model.User) (*TokenPair, error) {
	ctx := context.Background()
	tokenPair, err := j.tokenUsecase.GenerateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  tokenPair.Token,
		RefreshToken: "", // Use RefreshToken endpoint for refresh token
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// ValidateToken validates a JWT token
func (j *JWT) ValidateToken(token string) (*TokenClaims, error) {
	ctx := context.Background()
	response, err := j.tokenHandler.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if !response.Valid {
		return nil, ErrInvalidToken
	}

	return &TokenClaims{
		UserID:   response.UserID,
		Username: response.Username,
		Role:     response.Role,
	}, nil
}

// RefreshToken refreshes an expired access token
func (j *JWT) RefreshToken(refreshToken string) (*TokenPair, error) {
	ctx := context.Background()
	response, err := j.tokenHandler.RefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  response.Token,
		RefreshToken: "",
		ExpiresIn:    response.ExpiresIn,
	}, nil
}

// RevokeToken revokes a token
func (j *JWT) RevokeToken(token string) error {
	ctx := context.Background()
	return j.tokenHandler.RevokeToken(ctx, token)
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	UserID   string
	Username string
	Role     string
}

// TokenResponse represents token response
type TokenResponse = dto.TokenResponse

// ErrInvalidToken is returned when a token is invalid
var ErrInvalidToken = model.ErrInvalidToken
