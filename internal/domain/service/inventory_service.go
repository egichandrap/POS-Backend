package service

import (
	"context"

	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
)

// InventoryService handles business logic for inventory operations
type InventoryService struct {
	inventoryRepo repository.InventoryRepository
}

// NewInventoryService creates a new InventoryService
func NewInventoryService(inventoryRepo repository.InventoryRepository) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
	}
}

// CreateInventory creates a new inventory item
func (s *InventoryService) CreateInventory(ctx context.Context, inv *model.Inventory) error {
	return s.inventoryRepo.Create(ctx, inv)
}

// GetInventory retrieves an inventory item by ID
func (s *InventoryService) GetInventory(ctx context.Context, id string) (*model.Inventory, error) {
	return s.inventoryRepo.GetByID(ctx, id)
}

// UpdateInventory updates an existing inventory item
func (s *InventoryService) UpdateInventory(ctx context.Context, inv *model.Inventory) error {
	return s.inventoryRepo.Update(ctx, inv)
}

// DeleteInventory deletes an inventory item
func (s *InventoryService) DeleteInventory(ctx context.Context, id string) error {
	return s.inventoryRepo.Delete(ctx, id)
}

// ListInventory retrieves a paginated list of inventory items
func (s *InventoryService) ListInventory(ctx context.Context, filter *repository.InventoryFilter) ([]*model.Inventory, error) {
	return s.inventoryRepo.List(ctx, filter)
}

// CountInventory returns total count of inventory items
func (s *InventoryService) CountInventory(ctx context.Context, filter *repository.InventoryFilter) (int64, error) {
	return s.inventoryRepo.Count(ctx, filter)
}

// UpdateStock updates the quantity of an inventory item
func (s *InventoryService) UpdateStock(ctx context.Context, id string, quantity int) error {
	return s.inventoryRepo.UpdateQuantity(ctx, id, quantity)
}
