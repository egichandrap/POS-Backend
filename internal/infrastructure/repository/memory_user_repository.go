package repository

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
	"github.com/example/jwt-ddd-clean/internal/domain/valueobject"
	"github.com/google/uuid"
)

// MemoryUserRepository implements repository.UserRepository for testing
type MemoryUserRepository struct {
	users map[string]*model.User
	mu    sync.RWMutex
}

// NewMemoryUserRepository creates a new MemoryUserRepository
func NewMemoryUserRepository() *MemoryUserRepository {
	repo := &MemoryUserRepository{
		users: make(map[string]*model.User),
	}

	// Seed default users
	repo.seedDefaultUsers()

	return repo
}

// seedDefaultUsers creates default users for development
func (r *MemoryUserRepository) seedDefaultUsers() {
	now := time.Now()
	hashedPassword := valueobject.Password("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy") // admin123
	email1, _ := valueobject.NewEmail("superadmin@pos.local")
	email2, _ := valueobject.NewEmail("admin@pos.local")
	email3, _ := valueobject.NewEmail("cashier@pos.local")

	// Super Admin
	superAdmin := model.ReconstructUser(
		uuid.New().String(),
		"", // No tenant for super admin
		"superadmin",
		email1,
		hashedPassword,
		"Super Administrator",
		model.RoleSuperAdmin,
		model.StatusActive,
		now, now, nil,
	)
	r.users[superAdmin.ID()] = superAdmin

	// Admin
	admin := model.ReconstructUser(
		uuid.New().String(),
		"", // No tenant for default admin
		"admin",
		email2,
		hashedPassword,
		"Administrator",
		model.RoleAdmin,
		model.StatusActive,
		now, now, nil,
	)
	r.users[admin.ID()] = admin

	// Cashier
	cashier := model.ReconstructUser(
		uuid.New().String(),
		"", // No tenant for default cashier
		"cashier",
		email3,
		hashedPassword,
		"Cashier User",
		model.RoleCashier,
		model.StatusActive,
		now, now, nil,
	)
	r.users[cashier.ID()] = cashier
}

// FindByID retrieves a user by their ID
func (r *MemoryUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// FindByUsername retrieves a user by their username
func (r *MemoryUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if strings.EqualFold(user.Username(), username) {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

// FindByEmail retrieves a user by their email
func (r *MemoryUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if strings.EqualFold(user.Email().String(), email) {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

// Create creates a new user
func (r *MemoryUserRepository) Create(ctx context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID()] = user
	return nil
}

// Update updates an existing user
func (r *MemoryUserRepository) Update(ctx context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.users[user.ID()]
	if !exists {
		return fmt.Errorf("user not found")
	}

	r.users[user.ID()] = user
	return nil
}

// Delete removes a user
func (r *MemoryUserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.users[id]
	if !exists {
		return fmt.Errorf("user not found")
	}

	delete(r.users, id)
	return nil
}

// List retrieves users with filtering
func (r *MemoryUserRepository) List(ctx context.Context, filter repository.UserFilter) ([]*model.User, error) {
	paginated, err := r.ListWithPagination(ctx, filter)
	if err != nil {
		return nil, err
	}
	return paginated.Users, nil
}

// Count returns total number of users with optional filter
func (r *MemoryUserRepository) Count(ctx context.Context, filter repository.UserFilter) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := int64(0)
	for _, user := range r.users {
		if matchesUserFilter(user, filter) {
			count++
		}
	}

	return count, nil
}

// ListWithPagination retrieves users with pagination
func (r *MemoryUserRepository) ListWithPagination(ctx context.Context, filter repository.UserFilter) (*repository.PaginatedUsers, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Filter users
	var filtered []*model.User
	for _, user := range r.users {
		if matchesUserFilter(user, filter) {
			filtered = append(filtered, user)
		}
	}

	total := int64(len(filtered))

	// Set defaults
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// Apply pagination
	start := filter.Offset
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + filter.Limit
	if end > len(filtered) {
		end = len(filtered)
	}

	paginated := filtered[start:end]
	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))

	return &repository.PaginatedUsers{
		Users:      paginated,
		Total:      total,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
		TotalPages: totalPages,
	}, nil
}

// ExistsByUsername checks if a username already exists
func (r *MemoryUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if strings.EqualFold(user.Username(), username) {
			return true, nil
		}
	}

	return false, nil
}

// ExistsByEmail checks if an email already exists
func (r *MemoryUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if strings.EqualFold(user.Email().String(), email) {
			return true, nil
		}
	}

	return false, nil
}

// UpdatePassword updates user password
func (r *MemoryUserRepository) UpdatePassword(ctx context.Context, id, hashedPassword string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.UpdatePassword(valueobject.Password(hashedPassword))
	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *MemoryUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.RecordLogin()
	return nil
}

// matchesUserFilter checks if a user matches the filter
func matchesUserFilter(user *model.User, filter repository.UserFilter) bool {
	if filter.Role != "" && user.Role() != filter.Role {
		return false
	}
	if filter.Status != "" && user.Status() != filter.Status {
		return false
	}
	if filter.Search != "" {
		search := strings.ToLower(filter.Search)
		if !strings.Contains(strings.ToLower(user.Username()), search) &&
			!strings.Contains(strings.ToLower(user.Email().String()), search) &&
			!strings.Contains(strings.ToLower(user.FullName()), search) {
			return false
		}
	}
	return true
}

// Ensure interface implementation
var _ repository.UserRepository = (*MemoryUserRepository)(nil)
