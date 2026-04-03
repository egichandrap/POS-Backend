package service

import (
	"context"
	"fmt"
	"time"

	"github.com/example/jwt-ddd-clean/internal/application/dto"
	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
	"github.com/example/jwt-ddd-clean/internal/domain/valueobject"
	"github.com/example/jwt-ddd-clean/internal/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// JWTProvider defines the interface for JWT operations
type JWTProvider interface {
	GenerateToken(claims *model.TokenClaims, expiresAt time.Time) (string, error)
	GenerateTokenWithDuration(userID, username string, role model.UserRole, expiration time.Duration) (string, error)
	ValidateToken(token string) (*model.TokenClaims, error)
	GetExpiration(token string) (time.Time, error)
}

// AuthService handles authentication business logic
type AuthService struct {
	userRepo    repository.UserRepository
	tokenRepo   repository.TokenRepository
	jwtProvider JWTProvider
}

// NewAuthService creates a new AuthService
func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	jwtProvider JWTProvider,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		jwtProvider: jwtProvider,
	}
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	// Find user by username
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.NewValidationError("username atau password salah")
	}

	// Check if user is active
	if !user.IsActive() {
		return nil, errors.NewValidationError("akun tidak aktif atau telah ditangguhkan")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash().String()), []byte(req.Password)); err != nil {
		return nil, errors.NewValidationError("username atau password salah")
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID()); err != nil {
		// Log error but don't fail the login
		fmt.Printf("failed to update last login: %v\n", err)
	}

	// Generate tokens
	accessToken, err := s.jwtProvider.GenerateTokenWithDuration(user.ID(), user.Username(), user.Role(), 24*time.Hour)
	if err != nil {
		return nil, errors.NewInternalError("gagal membuat access token")
	}

	refreshToken, err := s.jwtProvider.GenerateTokenWithDuration(user.ID(), user.Username(), user.Role(), 7*24*time.Hour)
	if err != nil {
		return nil, errors.NewInternalError("gagal membuat refresh token")
	}

	expiresIn := int64(24 * time.Hour.Seconds())

	userResp := dto.ToUserResponse(user)

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User:         userResp,
	}, nil
}

// Logout invalidates user tokens
func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	// Blacklist the access token
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.tokenRepo.Blacklist(ctx, accessToken, expiresAt); err != nil {
		return errors.NewInternalError("gagal melakukan logout")
	}

	return nil
}

// Register creates a new user
func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error) {
	// Convert DTO to value objects
	email, err := valueobject.NewEmail(req.Email)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	password, err := valueobject.NewPassword(req.Password)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// Check if username exists
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.NewInternalError("gagal memeriksa username")
	}
	if exists {
		return nil, errors.NewValidationError("username telah digunakan")
	}

	// Check if email exists
	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.NewInternalError("gagal memeriksa email")
	}
	if exists {
		return nil, errors.NewValidationError("email telah digunakan")
	}

	// Create user
	user, err := model.NewUser(req.Username, email, password, req.FullName, req.Role)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.NewInternalError("gagal membuat user: %v", err)
	}

	resp := dto.ToUserResponse(user)
	return &resp, nil
}

// RefreshToken generates new access token using refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	// Validate refresh token
	claims, err := s.jwtProvider.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.NewValidationError("refresh token tidak valid")
	}

	// Check if refresh token is blacklisted
	isBlacklisted, err := s.tokenRepo.IsBlacklisted(ctx, refreshToken)
	if err != nil {
		return nil, errors.NewInternalError("gagal memvalidasi refresh token")
	}
	if isBlacklisted {
		return nil, errors.NewValidationError("refresh token telah dicabut")
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.NewNotFoundError("user", "id", claims.UserID)
	}

	// Check if user is active
	if !user.IsActive() {
		return nil, errors.NewValidationError("akun tidak aktif atau telah ditangguhkan")
	}

	// Generate new access token
	accessToken, err := s.jwtProvider.GenerateTokenWithDuration(user.ID(), user.Username(), user.Role(), 24*time.Hour)
	if err != nil {
		return nil, errors.NewInternalError("gagal membuat access token")
	}

	expiresIn := int64(24 * time.Hour.Seconds())
	userResp := dto.ToUserResponse(user)

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User:         userResp,
	}, nil
}

// GetMe returns current user information
func (s *AuthService) GetMe(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user", "id", userID)
	}

	resp := dto.ToUserResponse(user)
	return &resp, nil
}

// ChangePassword changes user password
func (s *AuthService) ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.NewNotFoundError("user", "id", userID)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash().String()), []byte(req.OldPassword)); err != nil {
		return errors.NewValidationError("password lama salah")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.NewInternalError("gagal memproses password")
	}

	// Update password
	user.UpdatePassword(valueobject.Password(hashedPassword))
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.NewInternalError("gagal mengubah password")
	}

	return nil
}

// ListUsers returns paginated list of users
func (s *AuthService) ListUsers(ctx context.Context, filter repository.UserFilter) (*dto.UserListResponse, error) {
	paginatedUsers, err := s.userRepo.ListWithPagination(ctx, filter)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengambil daftar user")
	}

	userResponses := make([]dto.UserResponse, len(paginatedUsers.Users))
	for i, user := range paginatedUsers.Users {
		userResponses[i] = dto.ToUserResponse(user)
	}

	return &dto.UserListResponse{
		Users:      userResponses,
		Total:      paginatedUsers.Total,
		Limit:      paginatedUsers.Limit,
		Offset:     paginatedUsers.Offset,
		TotalPages: paginatedUsers.TotalPages,
	}, nil
}

// UpdateUser updates user information
func (s *AuthService) UpdateUser(ctx context.Context, userID string, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user", "id", userID)
	}

	// Update fields using domain methods
	if req.Email != "" {
		emailVO, err := valueobject.NewEmail(req.Email)
		if err == nil {
			user.UpdateProfile(emailVO, user.FullName())
		}
	}
	if req.FullName != "" {
		user.UpdateProfile(user.Email(), req.FullName)
	}
	if req.Role != "" {
		user.UpdateRole(req.Role)
	}
	if req.Status != "" {
		switch req.Status {
		case model.StatusActive:
			user.Activate()
		case model.StatusInactive:
			user.Deactivate()
		case model.StatusSuspended:
			user.Suspend()
		}
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errors.NewInternalError("gagal mengupdate user")
	}

	resp := dto.ToUserResponse(user)
	return &resp, nil
}

// DeleteUser deletes a user
func (s *AuthService) DeleteUser(ctx context.Context, userID string) error {
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.NewNotFoundError("user", "id", userID)
	}

	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return errors.NewInternalError("gagal menghapus user")
	}

	return nil
}
