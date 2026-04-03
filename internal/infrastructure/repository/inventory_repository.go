package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
)

// PostgresInventoryRepository implements InventoryRepository using PostgreSQL
type PostgresInventoryRepository struct {
	db *sql.DB
}

// NewPostgresInventoryRepository creates a new PostgreSQL inventory repository
func NewPostgresInventoryRepository(db *sql.DB) repository.InventoryRepository {
	return &PostgresInventoryRepository{
		db: db,
	}
}

// Create creates a new inventory item
func (r *PostgresInventoryRepository) Create(ctx context.Context, inventory *model.Inventory) error {
	query := `
		INSERT INTO inventories (id, sku, name, description, quantity, unit, location, min_stock, max_stock, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		inventory.ID(),
		inventory.SKU(),
		inventory.Name(),
		inventory.Description(),
		inventory.Quantity(),
		inventory.Unit(),
		inventory.Location(),
		inventory.MinStock(),
		inventory.MaxStock(),
		inventory.Price(),
		inventory.CreatedAt(),
		inventory.UpdatedAt(),
	)

	return err
}

// GetByID retrieves an inventory item by its ID
func (r *PostgresInventoryRepository) GetByID(ctx context.Context, id string) (*model.Inventory, error) {
	query := `
		SELECT id, sku, name, description, quantity, unit, location, min_stock, max_stock, price, created_at, updated_at
		FROM inventories
		WHERE id = $1
	`

	var invID, sku, name, description, unit, location string
	var quantity, minStock, maxStock int
	var price float64
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&invID, &sku, &name, &description, &quantity, &unit, &location,
		&minStock, &maxStock, &price, &createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return model.ReconstructInventory(invID, sku, name, description, quantity, unit, location, minStock, maxStock, price, createdAt, updatedAt), nil
}

// GetBySKU retrieves an inventory item by its SKU
func (r *PostgresInventoryRepository) GetBySKU(ctx context.Context, sku string) (*model.Inventory, error) {
	query := `
		SELECT id, sku, name, description, quantity, unit, location, min_stock, max_stock, price, created_at, updated_at
		FROM inventories
		WHERE sku = $1
	`

	var invID, skuVal, name, description, unit, location string
	var quantity, minStock, maxStock int
	var price float64
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, sku).Scan(
		&invID, &skuVal, &name, &description, &quantity, &unit, &location,
		&minStock, &maxStock, &price, &createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return model.ReconstructInventory(invID, skuVal, name, description, quantity, unit, location, minStock, maxStock, price, createdAt, updatedAt), nil
}

// Update updates an existing inventory item
func (r *PostgresInventoryRepository) Update(ctx context.Context, inventory *model.Inventory) error {
	query := `
		UPDATE inventories
		SET sku = $1, name = $2, description = $3, quantity = $4, unit = $5, location = $6, min_stock = $7, max_stock = $8, price = $9, updated_at = $10
		WHERE id = $11
	`

	_, err := r.db.ExecContext(ctx, query,
		inventory.SKU(),
		inventory.Name(),
		inventory.Description(),
		inventory.Quantity(),
		inventory.Unit(),
		inventory.Location(),
		inventory.MinStock(),
		inventory.MaxStock(),
		inventory.Price(),
		inventory.UpdatedAt(),
		inventory.ID(),
	)

	return err
}

// Delete removes an inventory item
func (r *PostgresInventoryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM inventories WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List retrieves a list of inventory items with optional filtering
func (r *PostgresInventoryRepository) List(ctx context.Context, filter *repository.InventoryFilter) ([]*model.Inventory, error) {
	query := `
		SELECT id, sku, name, description, quantity, unit, location, min_stock, max_stock, price, created_at, updated_at
		FROM inventories
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if filter != nil {
		if filter.SKU != nil {
			query += fmt.Sprintf(" AND sku LIKE $%d", argCount)
			args = append(args, "%"+*filter.SKU+"%")
			argCount++
		}
		if filter.Name != nil {
			query += fmt.Sprintf(" AND name LIKE $%d", argCount)
			args = append(args, "%"+*filter.Name+"%")
			argCount++
		}
		if filter.Location != nil {
			query += fmt.Sprintf(" AND location = $%d", argCount)
			args = append(args, *filter.Location)
			argCount++
		}
		if filter.MinQty != nil {
			query += fmt.Sprintf(" AND quantity >= $%d", argCount)
			args = append(args, *filter.MinQty)
			argCount++
		}
		if filter.MaxQty != nil {
			query += fmt.Sprintf(" AND quantity <= $%d", argCount)
			args = append(args, *filter.MaxQty)
			argCount++
		}
	}

	query += " ORDER BY created_at DESC"

	if filter != nil && filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++

		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argCount)
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventories []*model.Inventory
	for rows.Next() {
		var invID, sku, name, description, unit, location string
		var quantity, minStock, maxStock int
		var price float64
		var createdAt, updatedAt time.Time

		err := rows.Scan(&invID, &sku, &name, &description, &quantity, &unit, &location,
			&minStock, &maxStock, &price, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		inventories = append(inventories, model.ReconstructInventory(invID, sku, name, description, quantity, unit, location, minStock, maxStock, price, createdAt, updatedAt))
	}

	return inventories, rows.Err()
}

// Count returns the total count of inventory items
func (r *PostgresInventoryRepository) Count(ctx context.Context, filter *repository.InventoryFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM inventories WHERE 1=1`

	args := []interface{}{}
	argCount := 1

	if filter != nil {
		if filter.SKU != nil {
			query += fmt.Sprintf(" AND sku LIKE $%d", argCount)
			args = append(args, "%"+*filter.SKU+"%")
			argCount++
		}
		if filter.Name != nil {
			query += fmt.Sprintf(" AND name LIKE $%d", argCount)
			args = append(args, "%"+*filter.Name+"%")
			argCount++
		}
		if filter.Location != nil {
			query += fmt.Sprintf(" AND location = $%d", argCount)
			args = append(args, *filter.Location)
			argCount++
		}
		if filter.MinQty != nil {
			query += fmt.Sprintf(" AND quantity >= $%d", argCount)
			args = append(args, *filter.MinQty)
			argCount++
		}
		if filter.MaxQty != nil {
			query += fmt.Sprintf(" AND quantity <= $%d", argCount)
			args = append(args, *filter.MaxQty)
			argCount++
		}
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// UpdateQuantity updates the quantity of an inventory item
func (r *PostgresInventoryRepository) UpdateQuantity(ctx context.Context, id string, quantity int) error {
	query := `UPDATE inventories SET quantity = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, quantity, time.Now(), id)
	return err
}

// ExistsBySKU checks if an inventory item with the given SKU exists
func (r *PostgresInventoryRepository) ExistsBySKU(ctx context.Context, sku string, excludeID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM inventories WHERE sku = $1`
	args := []interface{}{sku}

	if excludeID != "" {
		query += " AND id != $2"
		args = append(args, excludeID)
	}
	query += ")"

	var exists bool
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	return exists, err
}

// MemoryInventoryRepository is an in-memory implementation for testing
type MemoryInventoryRepository struct {
	items map[string]*model.Inventory
}

// NewMemoryInventoryRepository creates a new in-memory inventory repository
func NewMemoryInventoryRepository() repository.InventoryRepository {
	return &MemoryInventoryRepository{
		items: make(map[string]*model.Inventory),
	}
}

// Create creates a new inventory item
func (r *MemoryInventoryRepository) Create(ctx context.Context, inventory *model.Inventory) error {
	// Generate ID if not set
	if inventory.ID() == "" {
		// We can't set ID directly since it's private, so we reconstruct with an ID
		id := uuid.New().String()
		*inventory = *model.ReconstructInventory(
			id,
			inventory.SKU(),
			inventory.Name(),
			inventory.Description(),
			inventory.Quantity(),
			inventory.Unit(),
			inventory.Location(),
			inventory.MinStock(),
			inventory.MaxStock(),
			inventory.Price(),
			inventory.CreatedAt(),
			inventory.UpdatedAt(),
		)
	}
	r.items[inventory.ID()] = inventory
	return nil
}

// GetByID retrieves an inventory item by its ID
func (r *MemoryInventoryRepository) GetByID(ctx context.Context, id string) (*model.Inventory, error) {
	if inv, ok := r.items[id]; ok {
		return inv, nil
	}
	return nil, nil
}

// GetBySKU retrieves an inventory item by its SKU
func (r *MemoryInventoryRepository) GetBySKU(ctx context.Context, sku string) (*model.Inventory, error) {
	for _, inv := range r.items {
		if inv.SKU() == sku {
			return inv, nil
		}
	}
	return nil, nil
}

// Update updates an existing inventory item
func (r *MemoryInventoryRepository) Update(ctx context.Context, inventory *model.Inventory) error {
	r.items[inventory.ID()] = inventory
	return nil
}

// Delete removes an inventory item
func (r *MemoryInventoryRepository) Delete(ctx context.Context, id string) error {
	delete(r.items, id)
	return nil
}

// List retrieves a list of inventory items with optional filtering
func (r *MemoryInventoryRepository) List(ctx context.Context, filter *repository.InventoryFilter) ([]*model.Inventory, error) {
	var result []*model.Inventory

	for _, inv := range r.items {
		if filter != nil {
			if filter.SKU != nil && !strings.Contains(inv.SKU(), *filter.SKU) {
				continue
			}
			if filter.Name != nil && !strings.Contains(inv.Name(), *filter.Name) {
				continue
			}
			if filter.Location != nil && inv.Location() != *filter.Location {
				continue
			}
			if filter.MinQty != nil && inv.Quantity() < *filter.MinQty {
				continue
			}
			if filter.MaxQty != nil && inv.Quantity() > *filter.MaxQty {
				continue
			}
		}
		result = append(result, inv)
	}

	// Apply pagination
	if filter != nil && filter.Limit > 0 {
		start := filter.Offset
		end := start + filter.Limit

		if start >= len(result) {
			return []*model.Inventory{}, nil
		}
		if end > len(result) {
			end = len(result)
		}

		result = result[start:end]
	}

	return result, nil
}

// Count returns the total count of inventory items
func (r *MemoryInventoryRepository) Count(ctx context.Context, filter *repository.InventoryFilter) (int64, error) {
	// Count all items without filter, or count filtered items
	count := 0
	for _, inv := range r.items {
		if filter != nil {
			if filter.SKU != nil && !strings.Contains(inv.SKU(), *filter.SKU) {
				continue
			}
			if filter.Name != nil && !strings.Contains(inv.Name(), *filter.Name) {
				continue
			}
			if filter.Location != nil && inv.Location() != *filter.Location {
				continue
			}
			if filter.MinQty != nil && inv.Quantity() < *filter.MinQty {
				continue
			}
			if filter.MaxQty != nil && inv.Quantity() > *filter.MaxQty {
				continue
			}
		}
		count++
	}
	return int64(count), nil
}

// UpdateQuantity updates the quantity of an inventory item
func (r *MemoryInventoryRepository) UpdateQuantity(ctx context.Context, id string, quantity int) error {
	if inv, ok := r.items[id]; ok {
		// We can't directly set quantity, so reconstruct
		*inv = *model.ReconstructInventory(
			inv.ID(),
			inv.SKU(),
			inv.Name(),
			inv.Description(),
			quantity,
			inv.Unit(),
			inv.Location(),
			inv.MinStock(),
			inv.MaxStock(),
			inv.Price(),
			inv.CreatedAt(),
			time.Now(),
		)
	}
	return nil
}

// ExistsBySKU checks if an inventory item with the given SKU exists
func (r *MemoryInventoryRepository) ExistsBySKU(ctx context.Context, sku string, excludeID string) (bool, error) {
	for id, inv := range r.items {
		if inv.SKU() == sku && id != excludeID {
			return true, nil
		}
	}
	return false, nil
}
