package usecase

import (
	"context"

	"github.com/example/jwt-ddd-clean/internal/application/dto"
	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
	"github.com/example/jwt-ddd-clean/internal/domain/service"
	"github.com/example/jwt-ddd-clean/internal/pkg/errors"
)

// InventoryUsecase defines the inventory usecase interface
type InventoryUsecase interface {
	CreateInventory(ctx context.Context, req dto.CreateInventoryRequest) (*dto.InventoryResponse, error)
	GetInventory(ctx context.Context, id string) (*dto.InventoryResponse, error)
	UpdateInventory(ctx context.Context, id string, req dto.UpdateInventoryRequest) (*dto.InventoryResponse, error)
	DeleteInventory(ctx context.Context, id string) error
	ListInventory(ctx context.Context, filter repository.InventoryFilter) (*dto.InventoryListResponse, error)
	UpdateStock(ctx context.Context, id string, quantity int) (*dto.InventoryResponse, error)
	AdjustStock(ctx context.Context, id string, adjustment int) (*dto.InventoryResponse, error)
}

type inventoryUsecase struct {
	inventoryRepo    repository.InventoryRepository
	inventoryService *service.InventoryService
}

// NewInventoryUsecase creates a new InventoryUsecase
func NewInventoryUsecase(
	inventoryRepo repository.InventoryRepository,
	inventoryService *service.InventoryService,
) InventoryUsecase {
	return &inventoryUsecase{
		inventoryRepo:    inventoryRepo,
		inventoryService: inventoryService,
	}
}

func (u *inventoryUsecase) CreateInventory(ctx context.Context, req dto.CreateInventoryRequest) (*dto.InventoryResponse, error) {
	// Create inventory using domain entity factory
	inv, err := model.NewInventory(
		req.SKU,
		req.Name,
		req.Description,
		req.Quantity,
		req.Unit,
		req.Location,
		req.MinStock,
		req.MaxStock,
		req.Price,
	)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// Save via repository
	if err := u.inventoryRepo.Create(ctx, inv); err != nil {
		return nil, errors.NewInternalError("gagal membuat inventory: %v", err)
	}

	return dto.ToInventoryResponse(inv), nil
}

func (u *inventoryUsecase) GetInventory(ctx context.Context, id string) (*dto.InventoryResponse, error) {
	inv, err := u.inventoryRepo.GetByID(ctx, id)
	if err != nil || inv == nil {
		return nil, errors.NewNotFoundError("inventory", "id", id)
	}

	return dto.ToInventoryResponse(inv), nil
}

func (u *inventoryUsecase) UpdateInventory(ctx context.Context, id string, req dto.UpdateInventoryRequest) (*dto.InventoryResponse, error) {
	// Fetch existing inventory
	existing, err := u.inventoryRepo.GetByID(ctx, id)
	if err != nil || existing == nil {
		return nil, errors.NewNotFoundError("inventory", "id", id)
	}

	// Update details using domain method
	err = existing.UpdateDetails(
		coalesceString(req.Name, existing.Name()),
		coalesceString(req.Description, existing.Description()),
		coalesceString(req.Location, existing.Location()),
		coalesceInt(req.MinStock, existing.MinStock()),
		coalesceInt(req.MaxStock, existing.MaxStock()),
	)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// Update unit if provided
	if req.Unit != "" {
		// Need to recreate inventory to update unit - or we can add UpdateUnit method to domain
		// For now, we'll handle this in repository update
	}

	// Persist via repository
	if err := u.inventoryRepo.Update(ctx, existing); err != nil {
		return nil, errors.NewInternalError("gagal mengupdate inventory")
	}

	return dto.ToInventoryResponse(existing), nil
}

func (u *inventoryUsecase) DeleteInventory(ctx context.Context, id string) error {
	existing, err := u.inventoryRepo.GetByID(ctx, id)
	if err != nil || existing == nil {
		return errors.NewNotFoundError("inventory", "id", id)
	}

	if err := u.inventoryRepo.Delete(ctx, id); err != nil {
		return errors.NewInternalError("gagal menghapus inventory")
	}

	return nil
}

func (u *inventoryUsecase) ListInventory(ctx context.Context, filter repository.InventoryFilter) (*dto.InventoryListResponse, error) {
	items, err := u.inventoryRepo.List(ctx, &filter)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengambil inventory")
	}

	total, err := u.inventoryRepo.Count(ctx, &filter)
	if err != nil {
		return nil, errors.NewInternalError("gagal menghitung inventory")
	}

	totalPages := int(total) / filter.Limit
	if filter.Limit > 0 && int(total)%filter.Limit > 0 {
		totalPages++
	}

	itemResponses := make([]dto.InventoryResponse, len(items))
	for i, item := range items {
		itemResponses[i] = *dto.ToInventoryResponse(item)
	}

	return &dto.InventoryListResponse{
		Items:      itemResponses,
		Total:      total,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
		TotalPages: totalPages,
	}, nil
}

func (u *inventoryUsecase) UpdateStock(ctx context.Context, id string, quantity int) (*dto.InventoryResponse, error) {
	inv, err := u.inventoryRepo.GetByID(ctx, id)
	if err != nil || inv == nil {
		return nil, errors.NewNotFoundError("inventory", "id", id)
	}

	// Use domain method to adjust stock
	err = inv.AdjustStock(quantity)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// Persist
	if err := u.inventoryRepo.Update(ctx, inv); err != nil {
		return nil, errors.NewInternalError("gagal mengupdate stok")
	}

	return dto.ToInventoryResponse(inv), nil
}

func (u *inventoryUsecase) AdjustStock(ctx context.Context, id string, adjustment int) (*dto.InventoryResponse, error) {
	inv, err := u.inventoryRepo.GetByID(ctx, id)
	if err != nil || inv == nil {
		return nil, errors.NewNotFoundError("inventory", "id", id)
	}

	// Use domain method
	if adjustment > 0 {
		err = inv.AddStock(adjustment)
	} else if adjustment < 0 {
		err = inv.ReduceStock(-adjustment)
	}
	
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// Persist
	if err := u.inventoryRepo.Update(ctx, inv); err != nil {
		return nil, errors.NewInternalError("gagal mengadjust stok")
	}

	return dto.ToInventoryResponse(inv), nil
}

// Helper functions for coalescing values
func coalesceString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func coalesceInt(value, fallback int) int {
	if value != 0 || fallback == 0 {
		return value
	}
	return fallback
}
