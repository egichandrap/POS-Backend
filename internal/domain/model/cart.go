package model

import (
	"fmt"
	"time"
)

// CartItem represents an item in the shopping cart
type CartItem struct {
	productID   string
	productName string
	sku         string
	quantity    int
	unitPrice   float64
	subtotal    float64
}

// NewCartItem creates a new cart item
func NewCartItem(productID, productName, sku string, quantity int, unitPrice float64) (*CartItem, error) {
	if productID == "" {
		return nil, fmt.Errorf("product ID tidak boleh kosong")
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity harus lebih dari 0")
	}
	if unitPrice < 0 {
		return nil, fmt.Errorf("harga tidak boleh negatif")
	}

	return &CartItem{
		productID:   productID,
		productName: productName,
		sku:         sku,
		quantity:    quantity,
		unitPrice:   unitPrice,
		subtotal:    float64(quantity) * unitPrice,
	}, nil
}

// ReconstructCartItem recreates a cart item from database
func ReconstructCartItem(
	productID, productName, sku string,
	quantity int,
	unitPrice, subtotal float64,
) *CartItem {
	return &CartItem{
		productID:   productID,
		productName: productName,
		sku:         sku,
		quantity:    quantity,
		unitPrice:   unitPrice,
		subtotal:    subtotal,
	}
}

// Accessors
func (i *CartItem) ProductID() string     { return i.productID }
func (i *CartItem) ProductName() string   { return i.productName }
func (i *CartItem) SKU() string           { return i.sku }
func (i *CartItem) Quantity() int         { return i.quantity }
func (i *CartItem) UnitPrice() float64    { return i.unitPrice }
func (i *CartItem) Subtotal() float64     { return i.subtotal }

// UpdateQuantity updates the quantity and recalculates subtotal
func (i *CartItem) UpdateQuantity(quantity int) {
	if quantity > 0 {
		i.quantity = quantity
		i.subtotal = float64(quantity) * i.unitPrice
	}
}

// Cart represents a shopping cart aggregate root
type Cart struct {
	id           string
	userID       string
	customerName string
	items        []CartItem
	totalAmount  float64
	createdAt    time.Time
	updatedAt    time.Time
}

// NewCart creates a new cart entity
func NewCart(userID string, customerName string) (*Cart, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID tidak boleh kosong")
	}

	now := time.Now()
	return &Cart{
		userID:       userID,
		customerName: customerName,
		items:        make([]CartItem, 0),
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// ReconstructCart recreates a cart from database
func ReconstructCart(
	id, userID, customerName string,
	items []CartItem,
	totalAmount float64,
	createdAt, updatedAt time.Time,
) *Cart {
	return &Cart{
		id:           id,
		userID:       userID,
		customerName: customerName,
		items:        items,
		totalAmount:  totalAmount,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// Accessors
func (c *Cart) ID() string           { return c.id }
func (c *Cart) UserID() string       { return c.userID }
func (c *Cart) CustomerName() string { return c.customerName }
func (c *Cart) Items() []CartItem    { return c.items }
func (c *Cart) TotalAmount() float64 { return c.totalAmount }
func (c *Cart) CreatedAt() time.Time { return c.createdAt }
func (c *Cart) UpdatedAt() time.Time { return c.updatedAt }

// AddItem adds or updates an item in the cart
func (c *Cart) AddItem(productID, productName, sku string, quantity int, unitPrice float64) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity harus lebih dari 0")
	}

	// Check if item already exists
	for i, item := range c.items {
		if item.productID == productID {
			newQuantity := item.quantity + quantity
			c.items[i].UpdateQuantity(newQuantity)
			c.recalculateTotal()
			c.updatedAt = time.Now()
			return nil
		}
	}

	// Add new item
	newItem, err := NewCartItem(productID, productName, sku, quantity, unitPrice)
	if err != nil {
		return err
	}
	c.items = append(c.items, *newItem)
	c.recalculateTotal()
	c.updatedAt = time.Now()
	return nil
}

// RemoveItem removes an item from the cart by product ID
func (c *Cart) RemoveItem(productID string) error {
	for i, item := range c.items {
		if item.productID == productID {
			c.items = append(c.items[:i], c.items[i+1:]...)
			c.recalculateTotal()
			c.updatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("item dengan product ID %s tidak ditemukan", productID)
}

// UpdateItemQuantity updates the quantity of an item
func (c *Cart) UpdateItemQuantity(productID string, quantity int) error {
	if quantity < 0 {
		return fmt.Errorf("quantity tidak boleh negatif")
	}

	// If quantity is 0, remove the item
	if quantity == 0 {
		return c.RemoveItem(productID)
	}

	for i, item := range c.items {
		if item.productID == productID {
			c.items[i].UpdateQuantity(quantity)
			c.recalculateTotal()
			c.updatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("item dengan product ID %s tidak ditemukan", productID)
}

// Clear removes all items from the cart
func (c *Cart) Clear() {
	c.items = make([]CartItem, 0)
	c.totalAmount = 0
	c.updatedAt = time.Now()
}

// IsEmpty checks if cart has no items
func (c *Cart) IsEmpty() bool {
	return len(c.items) == 0
}

// ItemCount returns the number of items in the cart
func (c *Cart) ItemCount() int {
	return len(c.items)
}

// recalculateTotal recalculates the total amount
func (c *Cart) recalculateTotal() {
	c.totalAmount = 0
	for _, item := range c.items {
		c.totalAmount += item.subtotal
	}
}
