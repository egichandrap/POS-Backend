package model

import (
	"fmt"
	"time"

	"github.com/example/jwt-ddd-clean/internal/domain/valueobject"
	"github.com/google/uuid"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	RoleSuperAdmin UserRole = "SUPER_ADMIN"
	RoleAdmin      UserRole = "ADMIN"
	RoleCashier    UserRole = "CASHIER"
	RoleViewer     UserRole = "VIEWER"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	StatusActive    UserStatus = "ACTIVE"
	StatusInactive  UserStatus = "INACTIVE"
	StatusSuspended UserStatus = "SUSPENDED"
)

// User represents a user entity in the domain
type User struct {
	id           string
	tenantID     string // Multi-tenant: user belongs to a tenant
	username     string
	passwordHash valueobject.Password
	email        valueobject.Email
	fullName     string
	role         UserRole
	status       UserStatus
	createdAt    time.Time
	updatedAt    time.Time
	lastLoginAt  *time.Time
}

// NewUser creates a new user entity with full validation
func NewUser(
	username string,
	email valueobject.Email,
	password valueobject.Password,
	fullName string,
	role UserRole,
) (*User, error) {
	if username == "" {
		return nil, fmt.Errorf("username tidak boleh kosong")
	}
	if fullName == "" {
		return nil, fmt.Errorf("nama lengkap tidak boleh kosong")
	}
	if password.IsEmpty() {
		return nil, fmt.Errorf("password tidak boleh kosong")
	}

	now := time.Now()
	return &User{
		id:           uuid.New().String(),
		username:     username,
		email:        email,
		passwordHash: password,
		fullName:     fullName,
		role:         role,
		status:       StatusActive,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// ReconstructUser recreates a user entity from database (trusted data)
func ReconstructUser(
	id string,
	tenantID string,
	username string,
	email valueobject.Email,
	password valueobject.Password,
	fullName string,
	role UserRole,
	status UserStatus,
	createdAt time.Time,
	updatedAt time.Time,
	lastLoginAt *time.Time,
) *User {
	return &User{
		id:           id,
		tenantID:     tenantID,
		username:     username,
		email:        email,
		passwordHash: password,
		fullName:     fullName,
		role:         role,
		status:       status,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		lastLoginAt:  lastLoginAt,
	}
}

// Accessor methods (read-only)

// ID returns the user ID
func (u *User) ID() string {
	return u.id
}

// TenantID returns the tenant ID
func (u *User) TenantID() string {
	return u.tenantID
}

// SetTenantID sets the tenant ID (used during user creation/update)
func (u *User) SetTenantID(tenantID string) {
	u.tenantID = tenantID
	u.updatedAt = time.Now()
}

// Username returns the username
func (u *User) Username() string {
	return u.username
}

// PasswordHash returns the hashed password
func (u *User) PasswordHash() valueobject.Password {
	return u.passwordHash
}

// Email returns the email
func (u *User) Email() valueobject.Email {
	return u.email
}

// FullName returns the full name
func (u *User) FullName() string {
	return u.fullName
}

// Role returns the user role
func (u *User) Role() UserRole {
	return u.role
}

// Status returns the user status
func (u *User) Status() UserStatus {
	return u.status
}

// CreatedAt returns the creation time
func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

// UpdatedAt returns the last update time
func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// LastLoginAt returns the last login time
func (u *User) LastLoginAt() *time.Time {
	return u.lastLoginAt
}

// Ubiquitous language methods for user operations

// Activate activates the user account
func (u *User) Activate() {
	u.status = StatusActive
	u.updatedAt = time.Now()
}

// Deactivate deactivates the user account
func (u *User) Deactivate() {
	u.status = StatusInactive
	u.updatedAt = time.Now()
}

// Suspend suspends the user account
func (u *User) Suspend() {
	u.status = StatusSuspended
	u.updatedAt = time.Now()
}

// UpdatePassword updates the user password
func (u *User) UpdatePassword(newPassword valueobject.Password) {
	u.passwordHash = newPassword
	u.updatedAt = time.Now()
}

// RecordLogin records the user login timestamp
func (u *User) RecordLogin() {
	now := time.Now()
	u.lastLoginAt = &now
	u.updatedAt = now
}

// UpdateProfile updates the user profile information
func (u *User) UpdateProfile(email valueobject.Email, fullName string) {
	u.email = email
	u.fullName = fullName
	u.updatedAt = time.Now()
}

// UpdateRole updates the user role
func (u *User) UpdateRole(newRole UserRole) {
	u.role = newRole
	u.updatedAt = time.Now()
}

// Permission checks

// CanManageInventory checks if user has permission to manage inventory
func (u *User) CanManageInventory() bool {
	return u.role == RoleSuperAdmin || u.role == RoleAdmin
}

// CanAccessPOS checks if user can access POS features
func (u *User) CanAccessPOS() bool {
	return u.role == RoleSuperAdmin || u.role == RoleAdmin || u.role == RoleCashier
}

// CanManageUsers checks if user can manage other users
func (u *User) CanManageUsers() bool {
	return u.role == RoleSuperAdmin
}

// IsActive checks if user account is active
func (u *User) IsActive() bool {
	return u.status == StatusActive
}
