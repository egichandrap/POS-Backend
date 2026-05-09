package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RawMaterial represents a raw material/ingredient entity
type RawMaterial struct {
	id          string
	tenantID    string
	sku         string
	name        string
	description string
	unit        string
	quantity    float64
	minStock    float64
	costPerUnit float64
	supplier    string
	location    string
	isActive    bool
	createdAt   time.Time
	updatedAt   time.Time
}

// NewRawMaterial creates a new raw material entity with validation
func NewRawMaterial(
	tenantID string,
	sku string,
	name string,
	unit string,
	quantity float64,
	costPerUnit float64,
) (*RawMaterial, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID tidak boleh kosong")
	}
	if sku == "" {
		return nil, fmt.Errorf("SKU tidak boleh kosong")
	}
	if name == "" {
		return nil, fmt.Errorf("nama bahan baku tidak boleh kosong")
	}
	if unit == "" {
		return nil, fmt.Errorf("satuan tidak boleh kosong")
	}
	if quantity < 0 {
		return nil, fmt.Errorf("jumlah tidak boleh negatif")
	}
	if costPerUnit < 0 {
		return nil, fmt.Errorf("harga per satuan tidak boleh negatif")
	}

	now := time.Now()
	return &RawMaterial{
		id:          uuid.New().String(),
		tenantID:    tenantID,
		sku:         sku,
		name:        name,
		unit:        unit,
		quantity:    quantity,
		costPerUnit: costPerUnit,
		minStock:    0,
		isActive:    true,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// ReconstructRawMaterial recreates a raw material entity from database (trusted data)
func ReconstructRawMaterial(
	id, tenantID, sku, name, description, unit, supplier, location string,
	quantity, minStock, costPerUnit float64,
	isActive bool,
	createdAt, updatedAt time.Time,
) *RawMaterial {
	return &RawMaterial{
		id:          id,
		tenantID:    tenantID,
		sku:         sku,
		name:        name,
		description: description,
		unit:        unit,
		quantity:    quantity,
		minStock:    minStock,
		costPerUnit: costPerUnit,
		supplier:    supplier,
		location:    location,
		isActive:    isActive,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// Accessor methods (read-only)

func (r *RawMaterial) ID() string            { return r.id }
func (r *RawMaterial) TenantID() string      { return r.tenantID }
func (r *RawMaterial) SKU() string           { return r.sku }
func (r *RawMaterial) Name() string          { return r.name }
func (r *RawMaterial) Description() string   { return r.description }
func (r *RawMaterial) Unit() string          { return r.unit }
func (r *RawMaterial) Quantity() float64     { return r.quantity }
func (r *RawMaterial) MinStock() float64     { return r.minStock }
func (r *RawMaterial) CostPerUnit() float64  { return r.costPerUnit }
func (r *RawMaterial) Supplier() string      { return r.supplier }
func (r *RawMaterial) Location() string      { return r.location }
func (r *RawMaterial) IsActive() bool        { return r.isActive }
func (r *RawMaterial) CreatedAt() time.Time  { return r.createdAt }
func (r *RawMaterial) UpdatedAt() time.Time  { return r.updatedAt }

// Ubiquitous language methods for raw material operations

// AddStock adds stock to raw material
func (r *RawMaterial) AddStock(quantity float64) error {
	if quantity <= 0 {
		return fmt.Errorf("jumlah stok harus lebih dari 0")
	}
	r.quantity += quantity
	r.updatedAt = time.Now()
	return nil
}

// ReduceStock reduces stock from raw material
func (r *RawMaterial) ReduceStock(quantity float64) error {
	if quantity <= 0 {
		return fmt.Errorf("jumlah stok harus lebih dari 0")
	}
	if r.quantity-quantity < 0 {
		return fmt.Errorf("stok tidak mencukupi")
	}
	r.quantity -= quantity
	r.updatedAt = time.Now()
	return nil
}

// AdjustStock adjusts stock to a specific quantity
func (r *RawMaterial) AdjustStock(quantity float64) error {
	if quantity < 0 {
		return fmt.Errorf("stok tidak boleh negatif")
	}
	r.quantity = quantity
	r.updatedAt = time.Now()
	return nil
}

// UpdateCost updates the cost per unit
func (r *RawMaterial) UpdateCost(costPerUnit float64) error {
	if costPerUnit < 0 {
		return fmt.Errorf("harga per satuan tidak boleh negatif")
	}
	r.costPerUnit = costPerUnit
	r.updatedAt = time.Now()
	return nil
}

// UpdateDetails updates raw material details
func (r *RawMaterial) UpdateDetails(name, description, unit, supplier, location string, minStock float64) error {
	if name == "" {
		return fmt.Errorf("nama bahan baku tidak boleh kosong")
	}
	if unit == "" {
		return fmt.Errorf("satuan tidak boleh kosong")
	}
	if minStock < 0 {
		return fmt.Errorf("stok minimum tidak boleh negatif")
	}

	r.name = name
	r.description = description
	r.unit = unit
	r.supplier = supplier
	r.location = location
	r.minStock = minStock
	r.updatedAt = time.Now()
	return nil
}

// Activate activates the raw material
func (r *RawMaterial) Activate() {
	r.isActive = true
	r.updatedAt = time.Now()
}

// Deactivate deactivates the raw material
func (r *RawMaterial) Deactivate() {
	r.isActive = false
	r.updatedAt = time.Now()
}

// IsLowStock checks if raw material is below minimum stock level
func (r *RawMaterial) IsLowStock() bool {
	return r.quantity <= r.minStock
}

// IsOutOfStock checks if raw material is out of stock
func (r *RawMaterial) IsOutOfStock() bool {
	return r.quantity == 0
}

// StockStatus returns the stock status as a string
func (r *RawMaterial) StockStatus() string {
	if r.IsOutOfStock() {
		return "HABIS"
	}
	if r.IsLowStock() {
		return "RENDAH"
	}
	return "NORMAL"
}

// CanProduce checks if there's enough stock to produce a certain quantity
func (r *RawMaterial) CanProduce(requiredQty float64) bool {
	return r.quantity >= requiredQty
}

// GetTotalCost calculates the total cost of current stock
func (r *RawMaterial) GetTotalCost() float64 {
	return r.quantity * r.costPerUnit
}

// ProductRecipe represents a product recipe/ingredient relationship entity
type ProductRecipe struct {
	id                string
	tenantID          string
	inventoryID       string
	rawMaterialID     string
	quantityRequired  float64
	isActive          bool
	createdAt         time.Time
	updatedAt         time.Time
}

// NewProductRecipe creates a new product recipe entity with validation
func NewProductRecipe(
	tenantID string,
	inventoryID string,
	rawMaterialID string,
	quantityRequired float64,
) (*ProductRecipe, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID tidak boleh kosong")
	}
	if inventoryID == "" {
		return nil, fmt.Errorf("inventory ID tidak boleh kosong")
	}
	if rawMaterialID == "" {
		return nil, fmt.Errorf("raw material ID tidak boleh kosong")
	}
	if quantityRequired <= 0 {
		return nil, fmt.Errorf("jumlah yang dibutuhkan harus lebih dari 0")
	}

	now := time.Now()
	return &ProductRecipe{
		id:               uuid.New().String(),
		tenantID:         tenantID,
		inventoryID:      inventoryID,
		rawMaterialID:    rawMaterialID,
		quantityRequired: quantityRequired,
		isActive:         true,
		createdAt:        now,
		updatedAt:        now,
	}, nil
}

// ReconstructProductRecipe recreates a product recipe from database (trusted data)
func ReconstructProductRecipe(
	id, tenantID, inventoryID, rawMaterialID string,
	quantityRequired float64,
	isActive bool,
	createdAt, updatedAt time.Time,
) *ProductRecipe {
	return &ProductRecipe{
		id:               id,
		tenantID:         tenantID,
		inventoryID:      inventoryID,
		rawMaterialID:    rawMaterialID,
		quantityRequired: quantityRequired,
		isActive:         isActive,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
	}
}

// Accessor methods (read-only)

func (p *ProductRecipe) ID() string              { return p.id }
func (p *ProductRecipe) TenantID() string        { return p.tenantID }
func (p *ProductRecipe) InventoryID() string     { return p.inventoryID }
func (p *ProductRecipe) RawMaterialID() string   { return p.rawMaterialID }
func (p *ProductRecipe) QuantityRequired() float64 { return p.quantityRequired }
func (p *ProductRecipe) IsActive() bool          { return p.isActive }
func (p *ProductRecipe) CreatedAt() time.Time    { return p.createdAt }
func (p *ProductRecipe) UpdatedAt() time.Time    { return p.updatedAt }

// Ubiquitous language methods for product recipe operations

// UpdateQuantity updates the quantity required
func (p *ProductRecipe) UpdateQuantity(quantity float64) error {
	if quantity <= 0 {
		return fmt.Errorf("jumlah yang dibutuhkan harus lebih dari 0")
	}
	p.quantityRequired = quantity
	p.updatedAt = time.Now()
	return nil
}

// Activate activates the product recipe
func (p *ProductRecipe) Activate() {
	p.isActive = true
	p.updatedAt = time.Now()
}

// Deactivate deactivates the product recipe
func (p *ProductRecipe) Deactivate() {
	p.isActive = false
	p.updatedAt = time.Now()
}

// MaterialAvailabilityStatus represents the availability status of a material for production
type MaterialAvailabilityStatus struct {
	RawMaterialID    string
	RawMaterialName  string
	RawMaterialSKU   string
	RequiredQuantity float64
	AvailableQuantity float64
	Unit             string
	IsSufficient     bool
	Shortage         float64
}

// ProductAvailability represents the availability status of a product based on its materials
type ProductAvailability struct {
	InventoryID         string
	TenantID            string
	ProductName         string
	ProductSKU          string
	ProductQuantity     int
	TotalIngredients    int
	AvailableIngredients int
	IsAvailable         bool
	MaterialsStatus     []MaterialAvailabilityStatus
}

// IsAvailable checks if product can be produced based on material availability
func (p *ProductAvailability) IsFullyAvailable() bool {
	return p.IsAvailable && p.AvailableIngredients == p.TotalIngredients
}

// GetMissingMaterials returns list of materials that are insufficient
func (p *ProductAvailability) GetMissingMaterials() []MaterialAvailabilityStatus {
	var missing []MaterialAvailabilityStatus
	for _, m := range p.MaterialsStatus {
		if !m.IsSufficient {
			missing = append(missing, m)
		}
	}
	return missing
}
