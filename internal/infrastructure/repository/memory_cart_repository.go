package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
)

// MemoryCartRepository implements repository.CartRepository for testing
type MemoryCartRepository struct {
	carts map[string]*model.Cart
	mu    sync.RWMutex
}

// NewMemoryCartRepository creates a new MemoryCartRepository
func NewMemoryCartRepository() *MemoryCartRepository {
	return &MemoryCartRepository{
		carts: make(map[string]*model.Cart),
	}
}

// Create creates a new cart
func (r *MemoryCartRepository) Create(ctx context.Context, cart *model.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.carts[cart.ID()] = cart
	return nil
}

// GetByID retrieves a cart by ID
func (r *MemoryCartRepository) GetByID(ctx context.Context, id string) (*model.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cart, exists := r.carts[id]
	if !exists {
		return nil, fmt.Errorf("cart not found")
	}

	return cart, nil
}

// GetByUserID retrieves active cart by user ID
func (r *MemoryCartRepository) GetByUserID(ctx context.Context, userID string) (*model.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, cart := range r.carts {
		if cart.UserID() == userID {
			return cart, nil
		}
	}

	return nil, fmt.Errorf("cart not found")
}

// Update updates an existing cart
func (r *MemoryCartRepository) Update(ctx context.Context, cart *model.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.carts[cart.ID()] = cart
	return nil
}

// Delete removes a cart
func (r *MemoryCartRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.carts, id)
	return nil
}

// ClearItems removes all items from a cart
func (r *MemoryCartRepository) ClearItems(ctx context.Context, cartID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cart, exists := r.carts[cartID]
	if !exists {
		return fmt.Errorf("cart not found")
	}

	cart.Clear()
	return nil
}

// Ensure interface implementation
var _ repository.CartRepository = (*MemoryCartRepository)(nil)
