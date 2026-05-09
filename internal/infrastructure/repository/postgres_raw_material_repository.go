package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
)

type PostgresRawMaterialRepository struct {
	db *sql.DB
}

func NewPostgresRawMaterialRepository(db *sql.DB) repository.RawMaterialRepository {
	return &PostgresRawMaterialRepository{db: db}
}

func (r *PostgresRawMaterialRepository) Create(ctx context.Context, material *model.RawMaterial) error {
	query := `
		INSERT INTO raw_materials (
			id, tenant_id, sku, name, description, unit, quantity,
			min_stock, cost_per_unit, supplier, location, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		material.ID(), material.TenantID(), material.SKU(), material.Name(),
		material.Description(), material.Unit(), material.Quantity(),
		material.MinStock(), material.CostPerUnit(), material.Supplier(),
		material.Location(), material.IsActive(),
	)
	if err != nil {
		return fmt.Errorf("failed to create raw material: %w", err)
	}

	return nil
}

func (r *PostgresRawMaterialRepository) GetByID(ctx context.Context, id string) (*model.RawMaterial, error) {
	query := `
		SELECT id, tenant_id, sku, name, description, unit, quantity,
			min_stock, cost_per_unit, supplier, location, is_active,
			created_at, updated_at
		FROM raw_materials
		WHERE id = $1
	`

	var materialID, tenantID, sku, name, unit string
	var description, supplier, location sql.NullString
	var quantity, minStock, costPerUnit float64
	var isActive bool
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&materialID, &tenantID, &sku, &name, &description, &unit, &quantity,
		&minStock, &costPerUnit, &supplier, &location, &isActive,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("raw material not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get raw material: %w", err)
	}

	return model.ReconstructRawMaterial(
		materialID, tenantID, sku, name, description.String, unit, supplier.String, location.String,
		quantity, minStock, costPerUnit,
		isActive, createdAt, updatedAt,
	), nil
}

func (r *PostgresRawMaterialRepository) GetBySKU(ctx context.Context, tenantID, sku string) (*model.RawMaterial, error) {
	query := `
		SELECT id, tenant_id, sku, name, description, unit, quantity,
			min_stock, cost_per_unit, supplier, location, is_active,
			created_at, updated_at
		FROM raw_materials
		WHERE tenant_id = $1 AND sku = $2
	`

	var id, dbTenantID, sku2, name, unit string
	var description, supplier, location sql.NullString
	var quantity, minStock, costPerUnit float64
	var isActive bool
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, tenantID, sku).Scan(
		&id, &dbTenantID, &sku2, &name, &description, &unit, &quantity,
		&minStock, &costPerUnit, &supplier, &location, &isActive,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("raw material not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get raw material: %w", err)
	}

	return model.ReconstructRawMaterial(
		id, dbTenantID, sku2, name, description.String, unit, supplier.String, location.String,
		quantity, minStock, costPerUnit,
		isActive, createdAt, updatedAt,
	), nil
}

func (r *PostgresRawMaterialRepository) List(ctx context.Context, filter repository.RawMaterialFilter) ([]*model.RawMaterial, error) {
	query := `
		SELECT id, tenant_id, sku, name, description, unit, quantity,
			min_stock, cost_per_unit, supplier, location, is_active,
			created_at, updated_at
		FROM raw_materials
		WHERE tenant_id = $1
	`
	args := []interface{}{filter.TenantID}
	argCount := 2

	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argCount)
		args = append(args, *filter.IsActive)
		argCount++
	}

	if filter.Search != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR sku ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+filter.Search+"%")
		argCount++
	}

	if filter.Supplier != "" {
		query += fmt.Sprintf(" AND supplier = $%d", argCount)
		args = append(args, filter.Supplier)
		argCount++
	}

	if filter.LowStock {
		query += fmt.Sprintf(" AND quantity <= min_stock AND quantity > 0")
	}

	if filter.OutOfStock {
		query += fmt.Sprintf(" AND quantity = 0")
	}

	query += " ORDER BY name ASC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list raw materials: %w", err)
	}
	defer rows.Close()

	var materials []*model.RawMaterial
	for rows.Next() {
		var id, tenantID, sku, name, unit string
		var description, supplier, location sql.NullString
		var quantity, minStock, costPerUnit float64
		var isActive bool
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&id, &tenantID, &sku, &name, &description, &unit, &quantity,
			&minStock, &costPerUnit, &supplier, &location, &isActive,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan raw material: %w", err)
		}

		materials = append(materials, model.ReconstructRawMaterial(
			id, tenantID, sku, name, description.String, unit, supplier.String, location.String,
			quantity, minStock, costPerUnit,
			isActive, createdAt, updatedAt,
		))
	}

	return materials, nil
}

func (r *PostgresRawMaterialRepository) Update(ctx context.Context, material *model.RawMaterial) error {
	query := `
		UPDATE raw_materials SET
			sku = $2,
			name = $3,
			description = $4,
			unit = $5,
			quantity = $6,
			min_stock = $7,
			cost_per_unit = $8,
			supplier = $9,
			location = $10,
			is_active = $11
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		material.ID(), material.SKU(), material.Name(), material.Description(),
		material.Unit(), material.Quantity(), material.MinStock(),
		material.CostPerUnit(), material.Supplier(), material.Location(),
		material.IsActive(),
	)
	if err != nil {
		return fmt.Errorf("failed to update raw material: %w", err)
	}

	return nil
}

func (r *PostgresRawMaterialRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM raw_materials WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete raw material: %w", err)
	}
	return nil
}

func (r *PostgresRawMaterialRepository) UpdateStock(ctx context.Context, id string, quantity float64) error {
	query := `UPDATE raw_materials SET quantity = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, quantity)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}
	return nil
}

func (r *PostgresRawMaterialRepository) AdjustStock(ctx context.Context, id string, delta float64) error {
	query := `UPDATE raw_materials SET quantity = quantity + $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, delta)
	if err != nil {
		return fmt.Errorf("failed to adjust stock: %w", err)
	}
	return nil
}

func (r *PostgresRawMaterialRepository) GetLowStockMaterials(ctx context.Context, tenantID string) ([]*model.RawMaterial, error) {
	query := `
		SELECT id, tenant_id, sku, name, description, unit, quantity,
			min_stock, cost_per_unit, supplier, location, is_active,
			created_at, updated_at
		FROM raw_materials
		WHERE tenant_id = $1 AND quantity <= min_stock AND quantity > 0 AND is_active = true
		ORDER BY quantity ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock materials: %w", err)
	}
	defer rows.Close()

	var materials []*model.RawMaterial
	for rows.Next() {
		var id, tenantID, sku, name, unit string
		var description, supplier, location sql.NullString
		var quantity, minStock, costPerUnit float64
		var isActive bool
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&id, &tenantID, &sku, &name, &description, &unit, &quantity,
			&minStock, &costPerUnit, &supplier, &location, &isActive,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan raw material: %w", err)
		}

		materials = append(materials, model.ReconstructRawMaterial(
			id, tenantID, sku, name, description.String, unit, supplier.String, location.String,
			quantity, minStock, costPerUnit,
			isActive, createdAt, updatedAt,
		))
	}

	return materials, nil
}

func (r *PostgresRawMaterialRepository) GetOutOfStockMaterials(ctx context.Context, tenantID string) ([]*model.RawMaterial, error) {
	query := `
		SELECT id, tenant_id, sku, name, description, unit, quantity,
			min_stock, cost_per_unit, supplier, location, is_active,
			created_at, updated_at
		FROM raw_materials
		WHERE tenant_id = $1 AND quantity = 0 AND is_active = true
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get out of stock materials: %w", err)
	}
	defer rows.Close()

	var materials []*model.RawMaterial
	for rows.Next() {
		var id, tenantID, sku, name, unit string
		var description, supplier, location sql.NullString
		var quantity, minStock, costPerUnit float64
		var isActive bool
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&id, &tenantID, &sku, &name, &description, &unit, &quantity,
			&minStock, &costPerUnit, &supplier, &location, &isActive,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan raw material: %w", err)
		}

		materials = append(materials, model.ReconstructRawMaterial(
			id, tenantID, sku, name, description.String, unit, supplier.String, location.String,
			quantity, minStock, costPerUnit,
			isActive, createdAt, updatedAt,
		))
	}

	return materials, nil
}

func (r *PostgresRawMaterialRepository) ExistsBySKU(ctx context.Context, tenantID, sku string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM raw_materials WHERE tenant_id = $1 AND sku = $2)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, tenantID, sku).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check SKU existence: %w", err)
	}
	return exists, nil
}

func (r *PostgresRawMaterialRepository) Count(ctx context.Context, tenantID string) (int64, error) {
	query := `SELECT COUNT(*) FROM raw_materials WHERE tenant_id = $1`
	var count int64
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count raw materials: %w", err)
	}
	return count, nil
}

// PostgresProductRecipeRepository implements product recipe repository
type PostgresProductRecipeRepository struct {
	db *sql.DB
}

func NewPostgresProductRecipeRepository(db *sql.DB) repository.ProductRecipeRepository {
	return &PostgresProductRecipeRepository{db: db}
}

func (r *PostgresProductRecipeRepository) Create(ctx context.Context, recipe *model.ProductRecipe) error {
	query := `
		INSERT INTO product_recipes (
			id, tenant_id, inventory_id, raw_material_id, quantity_required, is_active
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		recipe.ID(), recipe.TenantID(), recipe.InventoryID(),
		recipe.RawMaterialID(), recipe.QuantityRequired(), recipe.IsActive(),
	)
	if err != nil {
		return fmt.Errorf("failed to create product recipe: %w", err)
	}

	return nil
}

func (r *PostgresProductRecipeRepository) GetByID(ctx context.Context, id string) (*model.ProductRecipe, error) {
	query := `
		SELECT id, tenant_id, inventory_id, raw_material_id, quantity_required, is_active,
			created_at, updated_at
		FROM product_recipes
		WHERE id = $1
	`

	var recipeID, tenantID, inventoryID, rawMaterialID string
	var quantityRequired float64
	var isActive bool
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&recipeID, &tenantID, &inventoryID, &rawMaterialID,
		&quantityRequired, &isActive, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product recipe not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product recipe: %w", err)
	}

	return model.ReconstructProductRecipe(
		recipeID, tenantID, inventoryID, rawMaterialID,
		quantityRequired, isActive, createdAt, updatedAt,
	), nil
}

func (r *PostgresProductRecipeRepository) GetByProductID(ctx context.Context, inventoryID string) ([]*model.ProductRecipe, error) {
	query := `
		SELECT pr.id, pr.tenant_id, pr.inventory_id, pr.raw_material_id,
			pr.quantity_required, pr.is_active, pr.created_at, pr.updated_at
		FROM product_recipes pr
		WHERE pr.inventory_id = $1 AND pr.is_active = true
		ORDER BY pr.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, inventoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product recipes: %w", err)
	}
	defer rows.Close()

	var recipes []*model.ProductRecipe
	for rows.Next() {
		var recipeID, tenantID, invID, rawMaterialID string
		var quantityRequired float64
		var isActive bool
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&recipeID, &tenantID, &invID, &rawMaterialID,
			&quantityRequired, &isActive, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product recipe: %w", err)
		}

		recipes = append(recipes, model.ReconstructProductRecipe(
			recipeID, tenantID, invID, rawMaterialID,
			quantityRequired, isActive, createdAt, updatedAt,
		))
	}

	return recipes, nil
}

func (r *PostgresProductRecipeRepository) GetByProductAndMaterial(ctx context.Context, inventoryID, rawMaterialID string) (*model.ProductRecipe, error) {
	query := `
		SELECT id, tenant_id, inventory_id, raw_material_id, quantity_required, is_active,
			created_at, updated_at
		FROM product_recipes
		WHERE inventory_id = $1 AND raw_material_id = $2
	`

	var recipeID, tenantID, invID, matID string
	var quantityRequired float64
	var isActive bool
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, inventoryID, rawMaterialID).Scan(
		&recipeID, &tenantID, &invID, &matID,
		&quantityRequired, &isActive, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product recipe not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product recipe: %w", err)
	}

	return model.ReconstructProductRecipe(
		recipeID, tenantID, invID, matID,
		quantityRequired, isActive, createdAt, updatedAt,
	), nil
}

func (r *PostgresProductRecipeRepository) List(ctx context.Context, filter repository.ProductRecipeFilter) ([]*model.ProductRecipe, error) {
	query := `
		SELECT pr.id, pr.tenant_id, pr.inventory_id, pr.raw_material_id,
			pr.quantity_required, pr.is_active, pr.created_at, pr.updated_at
		FROM product_recipes pr
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if filter.TenantID != "" {
		query += fmt.Sprintf(" AND pr.tenant_id = $%d", argCount)
		args = append(args, filter.TenantID)
		argCount++
	}

	if filter.InventoryID != "" {
		query += fmt.Sprintf(" AND pr.inventory_id = $%d", argCount)
		args = append(args, filter.InventoryID)
		argCount++
	}

	if filter.RawMaterialID != "" {
		query += fmt.Sprintf(" AND pr.raw_material_id = $%d", argCount)
		args = append(args, filter.RawMaterialID)
		argCount++
	}

	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND pr.is_active = $%d", argCount)
		args = append(args, *filter.IsActive)
		argCount++
	}

	query += " ORDER BY pr.created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list product recipes: %w", err)
	}
	defer rows.Close()

	var recipes []*model.ProductRecipe
	for rows.Next() {
		var recipeID, tenantID, invID, rawMaterialID string
		var quantityRequired float64
		var isActive bool
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&recipeID, &tenantID, &invID, &rawMaterialID,
			&quantityRequired, &isActive, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product recipe: %w", err)
		}

		recipes = append(recipes, model.ReconstructProductRecipe(
			recipeID, tenantID, invID, rawMaterialID,
			quantityRequired, isActive, createdAt, updatedAt,
		))
	}

	return recipes, nil
}

func (r *PostgresProductRecipeRepository) Update(ctx context.Context, recipe *model.ProductRecipe) error {
	query := `
		UPDATE product_recipes SET
			quantity_required = $2,
			is_active = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		recipe.ID(), recipe.QuantityRequired(), recipe.IsActive(),
	)
	if err != nil {
		return fmt.Errorf("failed to update product recipe: %w", err)
	}

	return nil
}

func (r *PostgresProductRecipeRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM product_recipes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product recipe: %w", err)
	}
	return nil
}

func (r *PostgresProductRecipeRepository) DeleteByProductID(ctx context.Context, inventoryID string) error {
	query := `DELETE FROM product_recipes WHERE inventory_id = $1`
	_, err := r.db.ExecContext(ctx, query, inventoryID)
	if err != nil {
		return fmt.Errorf("failed to delete product recipes: %w", err)
	}
	return nil
}

func (r *PostgresProductRecipeRepository) GetProductAvailability(ctx context.Context, inventoryID string) (*model.ProductAvailability, error) {
	query := `
		SELECT
			i.id AS inventory_id,
			i.tenant_id,
			i.name AS product_name,
			i.sku AS product_sku,
			i.quantity AS product_quantity,
			COUNT(pr.id) AS total_ingredients,
			COALESCE(SUM(CASE WHEN rm.quantity >= pr.quantity_required THEN 1 ELSE 0 END), 0) AS available_ingredients,
			CASE
				WHEN COUNT(pr.id) = 0 THEN true
				WHEN COALESCE(SUM(CASE WHEN rm.quantity >= pr.quantity_required THEN 1 ELSE 0 END), 0) = COUNT(pr.id) THEN true
				ELSE false
			END AS is_available
		FROM inventories i
		LEFT JOIN product_recipes pr ON i.id = pr.inventory_id AND pr.is_active = true
		LEFT JOIN raw_materials rm ON pr.raw_material_id = rm.id AND rm.is_active = true
		WHERE i.id = $1
		GROUP BY i.id, i.tenant_id, i.name, i.sku, i.quantity
	`

	var scanInventoryID, tenantID sql.NullString
	var productName, productSKU string
	var productQuantity, totalIngredients, availableIngredients int
	var isAvailable bool

	err := r.db.QueryRowContext(ctx, query, inventoryID).Scan(
		&scanInventoryID,
		&tenantID,
		&productName,
		&productSKU,
		&productQuantity,
		&totalIngredients,
		&availableIngredients,
		&isAvailable,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product availability: %w", err)
	}

	availability := &model.ProductAvailability{
		InventoryID:         scanInventoryID.String,
		TenantID:            tenantID.String,
		ProductName:         productName,
		ProductSKU:          productSKU,
		ProductQuantity:     productQuantity,
		TotalIngredients:    totalIngredients,
		AvailableIngredients: availableIngredients,
		IsAvailable:         isAvailable,
	}

	// Get detailed material status
	materialQuery := `
		SELECT
			rm.id AS material_id,
			rm.name AS material_name,
			rm.sku AS material_sku,
			pr.quantity_required AS required,
			rm.quantity AS available,
			rm.unit AS unit,
			rm.quantity >= pr.quantity_required AS is_sufficient
		FROM product_recipes pr
		INNER JOIN raw_materials rm ON pr.raw_material_id = rm.id
		WHERE pr.inventory_id = $1 AND pr.is_active = true AND rm.is_active = true
		ORDER BY rm.name ASC
	`

	rows, err := r.db.QueryContext(ctx, materialQuery, inventoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get material status: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status model.MaterialAvailabilityStatus
		err := rows.Scan(
			&status.RawMaterialID,
			&status.RawMaterialName,
			&status.RawMaterialSKU,
			&status.RequiredQuantity,
			&status.AvailableQuantity,
			&status.Unit,
			&status.IsSufficient,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan material status: %w", err)
		}

		if !status.IsSufficient {
			status.Shortage = status.RequiredQuantity - status.AvailableQuantity
		}

		availability.MaterialsStatus = append(availability.MaterialsStatus, status)
	}

	return availability, nil
}

func (r *PostgresProductRecipeRepository) GetBatchProductAvailability(ctx context.Context, inventoryIDs []string) ([]*model.ProductAvailability, error) {
	// Handle empty input
	if len(inventoryIDs) == 0 {
		return []*model.ProductAvailability{}, nil
	}

	// Build query with IN clause
	query := `
		SELECT
			i.id AS inventory_id,
			i.tenant_id,
			i.name AS product_name,
			i.sku AS product_sku,
			i.quantity AS product_quantity,
			COUNT(pr.id) AS total_ingredients,
			COALESCE(SUM(CASE WHEN rm.quantity >= pr.quantity_required THEN 1 ELSE 0 END), 0) AS available_ingredients,
			CASE
				WHEN COUNT(pr.id) = 0 THEN true
				WHEN COALESCE(SUM(CASE WHEN rm.quantity >= pr.quantity_required THEN 1 ELSE 0 END), 0) = COUNT(pr.id) THEN true
				ELSE false
			END AS is_available
		FROM inventories i
		LEFT JOIN product_recipes pr ON i.id = pr.inventory_id AND pr.is_active = true
		LEFT JOIN raw_materials rm ON pr.raw_material_id = rm.id AND rm.is_active = true
		WHERE i.id = ANY($1)
		GROUP BY i.id, i.tenant_id, i.name, i.sku, i.quantity
	`

	rows, err := r.db.QueryContext(ctx, query, inventoryIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch product availability: %w", err)
	}
	defer rows.Close()

	var availabilities []*model.ProductAvailability
	for rows.Next() {
		var availability model.ProductAvailability
		err := rows.Scan(
			&availability.InventoryID,
			&availability.TenantID,
			&availability.ProductName,
			&availability.ProductSKU,
			&availability.ProductQuantity,
			&availability.TotalIngredients,
			&availability.AvailableIngredients,
			&availability.IsAvailable,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product availability: %w", err)
		}

		availabilities = append(availabilities, &availability)
	}

	return availabilities, nil
}

func (r *PostgresProductRecipeRepository) GetMaterialsNeededForProduction(ctx context.Context, inventoryID string, quantity int) ([]model.MaterialAvailabilityStatus, error) {
	query := `
		SELECT
			rm.id AS material_id,
			rm.name AS material_name,
			rm.sku AS material_sku,
			(pr.quantity_required * $2) AS required,
			rm.quantity AS available,
			rm.unit AS unit,
			rm.quantity >= (pr.quantity_required * $2) AS is_sufficient
		FROM product_recipes pr
		INNER JOIN raw_materials rm ON pr.raw_material_id = rm.id
		WHERE pr.inventory_id = $1 AND pr.is_active = true AND rm.is_active = true
		ORDER BY rm.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, inventoryID, quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to get materials needed: %w", err)
	}
	defer rows.Close()

	var materials []model.MaterialAvailabilityStatus
	for rows.Next() {
		var status model.MaterialAvailabilityStatus
		err := rows.Scan(
			&status.RawMaterialID,
			&status.RawMaterialName,
			&status.RawMaterialSKU,
			&status.RequiredQuantity,
			&status.AvailableQuantity,
			&status.Unit,
			&status.IsSufficient,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan material status: %w", err)
		}

		if !status.IsSufficient {
			status.Shortage = status.RequiredQuantity - status.AvailableQuantity
		}

		materials = append(materials, status)
	}

	return materials, nil
}
