package dto

import "github.com/example/jwt-ddd-clean/internal/domain/model"

// CreateInventoryRequest represents the request to create inventory
type CreateInventoryRequest struct {
	SKU         string  `json:"sku" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description,omitempty"`
	Quantity    int     `json:"quantity" validate:"required,min=0"`
	Unit        string  `json:"unit" validate:"required"`
	Location    string  `json:"location,omitempty"`
	MinStock    int     `json:"min_stock" validate:"min=0"`
	MaxStock    int     `json:"max_stock" validate:"min=0"`
	Price       float64 `json:"price" validate:"required,min=0"`
}

// UpdateInventoryRequest represents the request to update inventory
type UpdateInventoryRequest struct {
	Name        string `json:"name" validate:"omitempty"`
	Description string `json:"description,omitempty"`
	Unit        string `json:"unit" validate:"omitempty"`
	Location    string `json:"location,omitempty"`
	MinStock    int    `json:"min_stock" validate:"min=0"`
	MaxStock    int    `json:"max_stock" validate:"min=0"`
}

// UpdateStockRequest represents the request to update stock
type UpdateStockRequest struct {
	Quantity int `json:"quantity" validate:"required,min=0"`
}

// AdjustStockRequest represents the request to adjust stock
type AdjustStockRequest struct {
	Quantity int `json:"quantity" validate:"required,min=0"`
}

// InventoryResponse represents the inventory response
type InventoryResponse struct {
	ID          string  `json:"id"`
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	Unit        string  `json:"unit"`
	Location    string  `json:"location"`
	MinStock    int     `json:"min_stock"`
	MaxStock    int     `json:"max_stock"`
	Price       float64 `json:"price"`
	StockStatus string  `json:"stock_status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// InventoryListResponse represents paginated inventory list
type InventoryListResponse struct {
	Items      []InventoryResponse `json:"items"`
	Total      int64               `json:"total"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
	TotalPages int                 `json:"total_pages"`
}

// ToInventoryResponse converts Inventory model to DTO
func ToInventoryResponse(inv *model.Inventory) *InventoryResponse {
	return &InventoryResponse{
		ID:          inv.ID(),
		SKU:         inv.SKU(),
		Name:        inv.Name(),
		Description: inv.Description(),
		Quantity:    inv.Quantity(),
		Unit:        inv.Unit(),
		Location:    inv.Location(),
		MinStock:    inv.MinStock(),
		MaxStock:    inv.MaxStock(),
		Price:       inv.Price(),
		StockStatus: inv.StockStatus(),
		CreatedAt:   inv.CreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   inv.UpdatedAt().Format("2006-01-02T15:04:05Z"),
	}
}
