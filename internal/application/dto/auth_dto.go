package dto

import (
	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/valueobject"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Username string           `json:"username" validate:"required,min=3,max=50"`
	Email    string           `json:"email" validate:"required,email"`
	Password string           `json:"password" validate:"required,min=6"`
	FullName string           `json:"full_name" validate:"required"`
	Role     model.UserRole   `json:"role" validate:"required,oneof=SUPER_ADMIN ADMIN CASHIER VIEWER"`
}

// UpdateUserRequest represents the user update request payload
type UpdateUserRequest struct {
	Email    string           `json:"email" validate:"omitempty,email"`
	FullName string           `json:"full_name" validate:"omitempty"`
	Role     model.UserRole   `json:"role" validate:"omitempty,oneof=SUPER_ADMIN ADMIN CASHIER VIEWER"`
	Status   model.UserStatus `json:"status" validate:"omitempty,oneof=ACTIVE INACTIVE SUSPENDED"`
}

// ChangePasswordRequest represents the change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
	User         UserResponse `json:"user"`
}

// UserResponse represents the user response (without sensitive data)
type UserResponse struct {
	ID          string           `json:"id"`
	Username    string           `json:"username"`
	Email       string           `json:"email"`
	FullName    string           `json:"full_name"`
	Role        model.UserRole   `json:"role"`
	Status      model.UserStatus `json:"status"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
	LastLoginAt *string          `json:"last_login_at,omitempty"`
}

// UserListResponse represents the paginated user list response
type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
	TotalPages int            `json:"total_pages"`
}

// ToUserResponse converts a User model to UserResponse DTO
func ToUserResponse(user *model.User) UserResponse {
	resp := UserResponse{
		ID:        user.ID(),
		Username:  user.Username(),
		Email:     user.Email().String(),
		FullName:  user.FullName(),
		Role:      user.Role(),
		Status:    user.Status(),
		CreatedAt: user.CreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt().Format("2006-01-02T15:04:05Z"),
	}

	if user.LastLoginAt() != nil {
		t := user.LastLoginAt().Format("2006-01-02T15:04:05Z")
		resp.LastLoginAt = &t
	}

	return resp
}

// ToUserValueObjects converts DTO fields to domain value objects
func (r *RegisterRequest) ToUserValueObjects() (valueobject.Email, valueobject.Password, error) {
	email, err := valueobject.NewEmail(r.Email)
	if err != nil {
		return "", "", err
	}

	password, err := valueobject.NewPassword(r.Password)
	if err != nil {
		return "", "", err
	}

	return email, password, nil
}
