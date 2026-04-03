package model

import (
	"fmt"
	"time"
)

// Inventory represents a warehouse inventory item entity
type Inventory struct {
	id          string
	sku         string
	name        string
	description string
	quantity    int
	unit        string
	location    string
	minStock    int
	maxStock    int
	price       float64
	createdAt   time.Time
	updatedAt   time.Time
}

// NewInventory creates a new inventory entity with validation
func NewInventory(
	sku string,
	name string,
	description string,
	quantity int,
	unit string,
	location string,
	minStock int,
	maxStock int,
	price float64,
) (*Inventory, error) {
	if sku == "" {
		return nil, fmt.Errorf("SKU tidak boleh kosong")
	}
	if name == "" {
		return nil, fmt.Errorf("nama produk tidak boleh kosong")
	}
	if unit == "" {
		return nil, fmt.Errorf("satuan tidak boleh kosong")
	}
	if price < 0 {
		return nil, fmt.Errorf("harga tidak boleh negatif")
	}
	if quantity < 0 {
		return nil, fmt.Errorf("stok tidak boleh negatif")
	}
	if minStock < 0 {
		return nil, fmt.Errorf("stok minimum tidak boleh negatif")
	}
	if maxStock < 0 {
		return nil, fmt.Errorf("stok maksimum tidak boleh negatif")
	}
	if minStock > maxStock && maxStock > 0 {
		return nil, fmt.Errorf("stok minimum tidak boleh lebih besar dari stok maksimum")
	}

	now := time.Now()
	return &Inventory{
		sku:         sku,
		name:        name,
		description: description,
		quantity:    quantity,
		unit:        unit,
		location:    location,
		minStock:    minStock,
		maxStock:    maxStock,
		price:       price,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// ReconstructInventory recreates an inventory entity from database (trusted data)
func ReconstructInventory(
	id, sku, name, description string,
	quantity int,
	unit, location string,
	minStock, maxStock int,
	price float64,
	createdAt, updatedAt time.Time,
) *Inventory {
	return &Inventory{
		id:          id,
		sku:         sku,
		name:        name,
		description: description,
		quantity:    quantity,
		unit:        unit,
		location:    location,
		minStock:    minStock,
		maxStock:    maxStock,
		price:       price,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// Accessor methods (read-only)

// ID returns the inventory ID
func (i *Inventory) ID() string {
	return i.id
}

// SKU returns the SKU
func (i *Inventory) SKU() string {
	return i.sku
}

// Name returns the product name
func (i *Inventory) Name() string {
	return i.name
}

// Description returns the description
func (i *Inventory) Description() string {
	return i.description
}

// Quantity returns the current quantity
func (i *Inventory) Quantity() int {
	return i.quantity
}

// Unit returns the unit of measurement
func (i *Inventory) Unit() string {
	return i.unit
}

// Location returns the storage location
func (i *Inventory) Location() string {
	return i.location
}

// MinStock returns the minimum stock level
func (i *Inventory) MinStock() int {
	return i.minStock
}

// MaxStock returns the maximum stock level
func (i *Inventory) MaxStock() int {
	return i.maxStock
}

// Price returns the price
func (i *Inventory) Price() float64 {
	return i.price
}

// CreatedAt returns the creation time
func (i *Inventory) CreatedAt() time.Time {
	return i.createdAt
}

// UpdatedAt returns the last update time
func (i *Inventory) UpdatedAt() time.Time {
	return i.updatedAt
}

// Ubiquitous language methods for inventory operations

// AddStock adds stock to inventory
func (i *Inventory) AddStock(quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("jumlah stok harus lebih dari 0")
	}

	if i.maxStock > 0 && i.quantity+quantity > i.maxStock {
		return fmt.Errorf("stok melebihi batas maksimum (%d)", i.maxStock)
	}

	i.quantity += quantity
	i.updatedAt = time.Now()
	return nil
}

// ReduceStock reduces stock from inventory
func (i *Inventory) ReduceStock(quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("jumlah stok harus lebih dari 0")
	}

	if i.quantity-quantity < 0 {
		return fmt.Errorf("stok tidak mencukupi")
	}

	i.quantity -= quantity
	i.updatedAt = time.Now()
	return nil
}

// AdjustStock adjusts stock to a specific quantity
func (i *Inventory) AdjustStock(quantity int) error {
	if quantity < 0 {
		return fmt.Errorf("stok tidak boleh negatif")
	}

	if i.maxStock > 0 && quantity > i.maxStock {
		return fmt.Errorf("stok melebihi batas maksimum (%d)", i.maxStock)
	}

	i.quantity = quantity
	i.updatedAt = time.Now()
	return nil
}

// UpdatePrice updates the product price
func (i *Inventory) UpdatePrice(price float64) error {
	if price < 0 {
		return fmt.Errorf("harga tidak boleh negatif")
	}

	i.price = price
	i.updatedAt = time.Now()
	return nil
}

// UpdateDetails updates inventory details
func (i *Inventory) UpdateDetails(name, description, location string, minStock, maxStock int) error {
	if name == "" {
		return fmt.Errorf("nama produk tidak boleh kosong")
	}
	if minStock < 0 {
		return fmt.Errorf("stok minimum tidak boleh negatif")
	}
	if maxStock < 0 {
		return fmt.Errorf("stok maksimum tidak boleh negatif")
	}
	if minStock > maxStock && maxStock > 0 {
		return fmt.Errorf("stok minimum tidak boleh lebih besar dari stok maksimum")
	}

	i.name = name
	i.description = description
	i.location = location
	i.minStock = minStock
	i.maxStock = maxStock
	i.updatedAt = time.Now()
	return nil
}

// IsLowStock checks if inventory is below minimum stock level
func (i *Inventory) IsLowStock() bool {
	return i.quantity <= i.minStock
}

// IsOutOfStock checks if inventory is out of stock
func (i *Inventory) IsOutOfStock() bool {
	return i.quantity == 0
}

// IsOverstocked checks if inventory exceeds maximum stock level
func (i *Inventory) IsOverstocked() bool {
	return i.maxStock > 0 && i.quantity > i.maxStock
}

// StockStatus returns the stock status as a string
func (i *Inventory) StockStatus() string {
	if i.IsOutOfStock() {
		return "HABIS"
	}
	if i.IsLowStock() {
		return "RENDAH"
	}
	if i.IsOverstocked() {
		return "BERLEBIHAN"
	}
	return "NORMAL"
}
