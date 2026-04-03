package usecase

import (
	"context"

	"github.com/example/jwt-ddd-clean/internal/application/dto"
	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
	"github.com/example/jwt-ddd-clean/internal/domain/service"
	"github.com/example/jwt-ddd-clean/internal/domain/valueobject"
	"github.com/example/jwt-ddd-clean/internal/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// AuthUsecase defines the authentication usecase interface
type AuthUsecase interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	Logout(ctx context.Context, accessToken string) error
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error)
	GetMe(ctx context.Context, userID string) (*dto.UserResponse, error)
	ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) error
	ListUsers(ctx context.Context, filter repository.UserFilter) (*dto.UserListResponse, error)
	UpdateUser(ctx context.Context, userID string, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, userID string) error
}

type authUsecase struct {
	userRepo    repository.UserRepository
	tokenRepo   repository.TokenRepository
	authService *service.AuthService
}

// NewAuthUsecase creates a new AuthUsecase
func NewAuthUsecase(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	authService *service.AuthService,
) AuthUsecase {
	return &authUsecase{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		authService: authService,
	}
}

func (u *authUsecase) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	return u.authService.Login(ctx, req)
}

func (u *authUsecase) Logout(ctx context.Context, accessToken string) error {
	return u.authService.Logout(ctx, accessToken)
}

func (u *authUsecase) Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error) {
	// Convert DTO to value objects
	email, password, err := req.ToUserValueObjects()
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// Check if username exists
	exists, err := u.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.NewInternalError("gagal memeriksa username")
	}
	if exists {
		return nil, errors.NewValidationError("username telah digunakan")
	}

	// Check if email exists
	exists, err = u.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.NewInternalError("gagal memeriksa email")
	}
	if exists {
		return nil, errors.NewValidationError("email telah digunakan")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password.String()), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.NewInternalError("gagal memproses password")
	}

	// Create user using domain entity factory
	user, err := model.NewUser(req.Username, email, valueobject.Password(hashedPassword), req.FullName, req.Role)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, errors.NewInternalError("gagal membuat user: %v", err)
	}

	resp := dto.ToUserResponse(user)
	return &resp, nil
}

func (u *authUsecase) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	return u.authService.RefreshToken(ctx, refreshToken)
}

func (u *authUsecase) GetMe(ctx context.Context, userID string) (*dto.UserResponse, error) {
	return u.authService.GetMe(ctx, userID)
}

func (u *authUsecase) ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) error {
	return u.authService.ChangePassword(ctx, userID, req)
}

func (u *authUsecase) ListUsers(ctx context.Context, filter repository.UserFilter) (*dto.UserListResponse, error) {
	return u.authService.ListUsers(ctx, filter)
}

func (u *authUsecase) UpdateUser(ctx context.Context, userID string, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	return u.authService.UpdateUser(ctx, userID, req)
}

func (u *authUsecase) DeleteUser(ctx context.Context, userID string) error {
	return u.authService.DeleteUser(ctx, userID)
}
