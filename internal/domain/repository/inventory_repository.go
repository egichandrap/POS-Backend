package repository

import (
	"context"

	"github.com/example/jwt-ddd-clean/internal/domain/model"
)

// InventoryFilter defines filter options for inventory queries
type InventoryFilter struct {
	SKU      *string
	Name     *string
	Location *string
	MinQty   *int
	MaxQty   *int
	Limit    int
	Offset   int
}

// InventoryRepository defines the interface for inventory data operations
type InventoryRepository interface {
	// Create creates a new inventory item
	Create(ctx context.Context, inventory *model.Inventory) error

	// GetByID retrieves an inventory item by its ID
	GetByID(ctx context.Context, id string) (*model.Inventory, error)

	// GetBySKU retrieves an inventory item by its SKU
	GetBySKU(ctx context.Context, sku string) (*model.Inventory, error)

	// Update updates an existing inventory item
	Update(ctx context.Context, inventory *model.Inventory) error

	// Delete removes an inventory item
	Delete(ctx context.Context, id string) error

	// List retrieves a list of inventory items with optional filtering
	List(ctx context.Context, filter *InventoryFilter) ([]*model.Inventory, error)

	// Count returns the total count of inventory items
	Count(ctx context.Context, filter *InventoryFilter) (int64, error)

	// UpdateQuantity updates the quantity of an inventory item
	UpdateQuantity(ctx context.Context, id string, quantity int) error

	// ExistsBySKU checks if an inventory item with the given SKU exists
	ExistsBySKU(ctx context.Context, sku string, excludeID string) (bool, error)
}
