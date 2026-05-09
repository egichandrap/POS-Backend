package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
)

// MemoryTenantRepository implements TenantRepository for in-memory storage
type MemoryTenantRepository struct {
	tenants map[string]*model.Tenant
	mu      sync.RWMutex
}

func NewMemoryTenantRepository() repository.TenantRepository {
	return &MemoryTenantRepository{
		tenants: make(map[string]*model.Tenant),
	}
}

func (r *MemoryTenantRepository) Create(ctx context.Context, tenant *model.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[tenant.ID()] = tenant
	return nil
}

func (r *MemoryTenantRepository) GetByID(ctx context.Context, id string) (*model.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tenant, ok := r.tenants[id]
	if !ok {
		return nil, fmt.Errorf("tenant not found")
	}
	return tenant, nil
}

func (r *MemoryTenantRepository) GetByCompanySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, tenant := range r.tenants {
		if tenant.CompanySlug() == slug {
			return tenant, nil
		}
	}
	return nil, fmt.Errorf("tenant not found")
}

func (r *MemoryTenantRepository) GetByDomain(ctx context.Context, domain string) (*model.Tenant, error) {
	if domain == "" {
		return nil, fmt.Errorf("tenant not found")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, tenant := range r.tenants {
		if tenant.Domain() == domain {
			return tenant, nil
		}
	}
	return nil, fmt.Errorf("tenant not found")
}

func (r *MemoryTenantRepository) GetByUserID(ctx context.Context, userID string) (*model.Tenant, error) {
	// In a real implementation, you would have a user-tenant mapping
	// For now, return the first tenant
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, tenant := range r.tenants {
		return tenant, nil
	}
	return nil, fmt.Errorf("tenant not found")
}

func (r *MemoryTenantRepository) List(ctx context.Context, filter repository.TenantFilter) ([]*model.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tenants []*model.Tenant
	for _, tenant := range r.tenants {
		tenants = append(tenants, tenant)
	}
	return tenants, nil
}

func (r *MemoryTenantRepository) Update(ctx context.Context, tenant *model.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[tenant.ID()] = tenant
	return nil
}

func (r *MemoryTenantRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tenants, id)
	return nil
}

func (r *MemoryTenantRepository) ExistsByCompanySlug(ctx context.Context, slug string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, tenant := range r.tenants {
		if tenant.CompanySlug() == slug {
			return true, nil
		}
	}
	return false, nil
}

func (r *MemoryTenantRepository) ExistsByDomain(ctx context.Context, domain string) (bool, error) {
	if domain == "" {
		return false, nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, tenant := range r.tenants {
		if tenant.Domain() == domain {
			return true, nil
		}
	}
	return false, nil
}

func (r *MemoryTenantRepository) Count(ctx context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return int64(len(r.tenants)), nil
}

// MemorySubscriptionPlanRepository implements SubscriptionPlanRepository
type MemorySubscriptionPlanRepository struct {
	plans map[string]*model.SubscriptionPlanDetail
	mu    sync.RWMutex
}

func NewMemorySubscriptionPlanRepository() repository.SubscriptionPlanRepository {
	repo := &MemorySubscriptionPlanRepository{
		plans: make(map[string]*model.SubscriptionPlanDetail),
	}
	// Initialize with default plans
	repo.seedPlans()
	return repo
}

func (r *MemorySubscriptionPlanRepository) seedPlans() {
	now := time.Now()

	plans := []*model.SubscriptionPlanDetail{
		model.ReconstructSubscriptionPlanDetail(
			"plus",
			"Plus Plan",
			"Perfect for small businesses",
			499000, 4990000,
			5, 1, 100, 500,
			`{"pos":true,"inventory_management":true,"basic_reports":true,"multi_user":true,"qr_ordering":false,"advanced_reports":false,"api_access":false,"custom_branding":false,"multi_store":false,"raw_material_management":false}`,
			true,
			now, now,
		),
		model.ReconstructSubscriptionPlanDetail(
			"pro",
			"Pro Plan",
			"For growing businesses",
			1499000, 14990000,
			20, 3, 500, 5000,
			`{"pos":true,"inventory_management":true,"basic_reports":true,"multi_user":true,"qr_ordering":true,"advanced_reports":true,"api_access":true,"custom_branding":true,"multi_store":true,"raw_material_management":true}`,
			true,
			now, now,
		),
		model.ReconstructSubscriptionPlanDetail(
			"enterprise",
			"Enterprise Plan",
			"For large organizations",
			4999000, 49990000,
			-1, -1, -1, -1,
			`{"pos":true,"inventory_management":true,"basic_reports":true,"multi_user":true,"qr_ordering":true,"advanced_reports":true,"api_access":true,"custom_branding":true,"multi_store":true,"raw_material_management":true,"priority_support":true,"custom_integrations":true,"dedicated_server":true}`,
			true,
			now, now,
		),
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	for _, plan := range plans {
		r.plans[plan.ID()] = plan
	}
}

func (r *MemorySubscriptionPlanRepository) GetByID(ctx context.Context, id string) (*model.SubscriptionPlanDetail, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	plan, ok := r.plans[id]
	if !ok {
		return nil, fmt.Errorf("subscription plan not found")
	}
	return plan, nil
}

func (r *MemorySubscriptionPlanRepository) List(ctx context.Context) ([]*model.SubscriptionPlanDetail, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var plans []*model.SubscriptionPlanDetail
	for _, plan := range r.plans {
		plans = append(plans, plan)
	}
	return plans, nil
}

func (r *MemorySubscriptionPlanRepository) ListActive(ctx context.Context) ([]*model.SubscriptionPlanDetail, error) {
	return r.List(ctx)
}

// MemoryRawMaterialRepository implements RawMaterialRepository
type MemoryRawMaterialRepository struct {
	materials map[string]*model.RawMaterial
	mu        sync.RWMutex
}

func NewMemoryRawMaterialRepository() repository.RawMaterialRepository {
	return &MemoryRawMaterialRepository{
		materials: make(map[string]*model.RawMaterial),
	}
}

func (r *MemoryRawMaterialRepository) Create(ctx context.Context, material *model.RawMaterial) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.materials[material.ID()] = material
	return nil
}

func (r *MemoryRawMaterialRepository) GetByID(ctx context.Context, id string) (*model.RawMaterial, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	material, ok := r.materials[id]
	if !ok {
		return nil, fmt.Errorf("raw material not found")
	}
	return material, nil
}

func (r *MemoryRawMaterialRepository) GetBySKU(ctx context.Context, tenantID, sku string) (*model.RawMaterial, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, material := range r.materials {
		if material.TenantID() == tenantID && material.SKU() == sku {
			return material, nil
		}
	}
	return nil, fmt.Errorf("raw material not found")
}

func (r *MemoryRawMaterialRepository) List(ctx context.Context, filter repository.RawMaterialFilter) ([]*model.RawMaterial, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var materials []*model.RawMaterial
	for _, material := range r.materials {
		if material.TenantID() == filter.TenantID {
			materials = append(materials, material)
		}
	}
	return materials, nil
}

func (r *MemoryRawMaterialRepository) Update(ctx context.Context, material *model.RawMaterial) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.materials[material.ID()] = material
	return nil
}

func (r *MemoryRawMaterialRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.materials, id)
	return nil
}

func (r *MemoryRawMaterialRepository) UpdateStock(ctx context.Context, id string, quantity float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if material, ok := r.materials[id]; ok {
		// This is a simplified version - in reality you'd use the domain method
		r.materials[id] = model.ReconstructRawMaterial(
			material.ID(), material.TenantID(), material.SKU(), material.Name(),
			material.Description(), material.Unit(), material.Supplier(), material.Location(),
			quantity, material.MinStock(), material.CostPerUnit(),
			material.IsActive(), material.CreatedAt(), time.Now(),
		)
	}
	return nil
}

func (r *MemoryRawMaterialRepository) AdjustStock(ctx context.Context, id string, delta float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if material, ok := r.materials[id]; ok {
		newQuantity := material.Quantity() + delta
		r.materials[id] = model.ReconstructRawMaterial(
			material.ID(), material.TenantID(), material.SKU(), material.Name(),
			material.Description(), material.Unit(), material.Supplier(), material.Location(),
			newQuantity, material.MinStock(), material.CostPerUnit(),
			material.IsActive(), material.CreatedAt(), time.Now(),
		)
	}
	return nil
}

func (r *MemoryRawMaterialRepository) GetLowStockMaterials(ctx context.Context, tenantID string) ([]*model.RawMaterial, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var materials []*model.RawMaterial
	for _, material := range r.materials {
		if material.TenantID() == tenantID && material.IsLowStock() {
			materials = append(materials, material)
		}
	}
	return materials, nil
}

func (r *MemoryRawMaterialRepository) GetOutOfStockMaterials(ctx context.Context, tenantID string) ([]*model.RawMaterial, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var materials []*model.RawMaterial
	for _, material := range r.materials {
		if material.TenantID() == tenantID && material.IsOutOfStock() {
			materials = append(materials, material)
		}
	}
	return materials, nil
}

func (r *MemoryRawMaterialRepository) ExistsBySKU(ctx context.Context, tenantID, sku string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, material := range r.materials {
		if material.TenantID() == tenantID && material.SKU() == sku {
			return true, nil
		}
	}
	return false, nil
}

func (r *MemoryRawMaterialRepository) Count(ctx context.Context, tenantID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := int64(0)
	for _, material := range r.materials {
		if material.TenantID() == tenantID {
			count++
		}
	}
	return count, nil
}

// MemoryProductRecipeRepository implements ProductRecipeRepository
type MemoryProductRecipeRepository struct {
	recipes map[string]*model.ProductRecipe
	mu      sync.RWMutex
}

func NewMemoryProductRecipeRepository() repository.ProductRecipeRepository {
	return &MemoryProductRecipeRepository{
		recipes: make(map[string]*model.ProductRecipe),
	}
}

func (r *MemoryProductRecipeRepository) Create(ctx context.Context, recipe *model.ProductRecipe) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recipes[recipe.ID()] = recipe
	return nil
}

func (r *MemoryProductRecipeRepository) GetByID(ctx context.Context, id string) (*model.ProductRecipe, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	recipe, ok := r.recipes[id]
	if !ok {
		return nil, fmt.Errorf("product recipe not found")
	}
	return recipe, nil
}

func (r *MemoryProductRecipeRepository) GetByProductID(ctx context.Context, inventoryID string) ([]*model.ProductRecipe, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var recipes []*model.ProductRecipe
	for _, recipe := range r.recipes {
		if recipe.InventoryID() == inventoryID && recipe.IsActive() {
			recipes = append(recipes, recipe)
		}
	}
	return recipes, nil
}

func (r *MemoryProductRecipeRepository) GetByProductAndMaterial(ctx context.Context, inventoryID, rawMaterialID string) (*model.ProductRecipe, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, recipe := range r.recipes {
		if recipe.InventoryID() == inventoryID && recipe.RawMaterialID() == rawMaterialID {
			return recipe, nil
		}
	}
	return nil, fmt.Errorf("product recipe not found")
}

func (r *MemoryProductRecipeRepository) List(ctx context.Context, filter repository.ProductRecipeFilter) ([]*model.ProductRecipe, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var recipes []*model.ProductRecipe
	for _, recipe := range r.recipes {
		recipes = append(recipes, recipe)
	}
	return recipes, nil
}

func (r *MemoryProductRecipeRepository) Update(ctx context.Context, recipe *model.ProductRecipe) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recipes[recipe.ID()] = recipe
	return nil
}

func (r *MemoryProductRecipeRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.recipes, id)
	return nil
}

func (r *MemoryProductRecipeRepository) DeleteByProductID(ctx context.Context, inventoryID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, recipe := range r.recipes {
		if recipe.InventoryID() == inventoryID {
			delete(r.recipes, id)
		}
	}
	return nil
}

func (r *MemoryProductRecipeRepository) GetProductAvailability(ctx context.Context, inventoryID string) (*model.ProductAvailability, error) {
	// Simplified implementation - returns available by default
	return &model.ProductAvailability{
		InventoryID:          inventoryID,
		ProductName:          "Sample Product",
		ProductSKU:           "SAMPLE-001",
		ProductQuantity:      100,
		TotalIngredients:     0,
		AvailableIngredients: 0,
		IsAvailable:          true,
		MaterialsStatus:      []model.MaterialAvailabilityStatus{},
	}, nil
}

func (r *MemoryProductRecipeRepository) GetBatchProductAvailability(ctx context.Context, inventoryIDs []string) ([]*model.ProductAvailability, error) {
	var availabilities []*model.ProductAvailability
	for _, id := range inventoryIDs {
		avail, _ := r.GetProductAvailability(ctx, id)
		availabilities = append(availabilities, avail)
	}
	return availabilities, nil
}

func (r *MemoryProductRecipeRepository) GetMaterialsNeededForProduction(ctx context.Context, inventoryID string, quantity int) ([]model.MaterialAvailabilityStatus, error) {
	return []model.MaterialAvailabilityStatus{}, nil
}

// MemoryTenantUsageRepository implements TenantUsageRepository
type MemoryTenantUsageRepository struct {
	usage map[string]*repository.TenantUsage
	mu    sync.RWMutex
}

func NewMemoryTenantUsageRepository() repository.TenantUsageRepository {
	return &MemoryTenantUsageRepository{
		usage: make(map[string]*repository.TenantUsage),
	}
}

func (r *MemoryTenantUsageRepository) GetByTenantID(ctx context.Context, tenantID string) (*repository.TenantUsage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if usage, ok := r.usage[tenantID]; ok {
		return usage, nil
	}
	// Return default usage if not found
	return &repository.TenantUsage{
		TenantID:           tenantID,
		CurrentUsers:       0,
		CurrentStores:      0,
		CurrentProducts:    0,
		TransactionsToday:  0,
		LastResetDate:      time.Now(),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}, nil
}

func (r *MemoryTenantUsageRepository) Create(ctx context.Context, usage *repository.TenantUsage) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.usage[usage.TenantID] = usage
	return nil
}

func (r *MemoryTenantUsageRepository) Update(ctx context.Context, usage *repository.TenantUsage) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.usage[usage.TenantID] = usage
	return nil
}

func (r *MemoryTenantUsageRepository) IncrementUserCount(ctx context.Context, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if usage, ok := r.usage[tenantID]; ok {
		usage.CurrentUsers++
		usage.UpdatedAt = time.Now()
	}
	return nil
}

func (r *MemoryTenantUsageRepository) DecrementUserCount(ctx context.Context, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if usage, ok := r.usage[tenantID]; ok {
		if usage.CurrentUsers > 0 {
			usage.CurrentUsers--
		}
		usage.UpdatedAt = time.Now()
	}
	return nil
}

func (r *MemoryTenantUsageRepository) IncrementStoreCount(ctx context.Context, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if usage, ok := r.usage[tenantID]; ok {
		usage.CurrentStores++
		usage.UpdatedAt = time.Now()
	}
	return nil
}

func (r *MemoryTenantUsageRepository) DecrementStoreCount(ctx context.Context, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if usage, ok := r.usage[tenantID]; ok {
		if usage.CurrentStores > 0 {
			usage.CurrentStores--
		}
		usage.UpdatedAt = time.Now()
	}
	return nil
}

func (r *MemoryTenantUsageRepository) IncrementProductCount(ctx context.Context, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if usage, ok := r.usage[tenantID]; ok {
		usage.CurrentProducts++
		usage.UpdatedAt = time.Now()
	}
	return nil
}

func (r *MemoryTenantUsageRepository) DecrementProductCount(ctx context.Context, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if usage, ok := r.usage[tenantID]; ok {
		if usage.CurrentProducts > 0 {
			usage.CurrentProducts--
		}
		usage.UpdatedAt = time.Now()
	}
	return nil
}

func (r *MemoryTenantUsageRepository) IncrementTransactionCount(ctx context.Context, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if usage, ok := r.usage[tenantID]; ok {
		usage.TransactionsToday++
		usage.UpdatedAt = time.Now()
	}
	return nil
}

func (r *MemoryTenantUsageRepository) ResetDailyTransactions(ctx context.Context, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if usage, ok := r.usage[tenantID]; ok {
		usage.TransactionsToday = 0
		usage.LastResetDate = time.Now()
		usage.UpdatedAt = time.Now()
	}
	return nil
}

func (r *MemoryTenantUsageRepository) CheckLimits(ctx context.Context, tenantID string) (*repository.LimitStatus, error) {
	// Return default - within limits
	return &repository.LimitStatus{
		WithinUserLimit:        true,
		WithinStoreLimit:       true,
		WithinProductLimit:     true,
		WithinTransactionLimit: true,
		UsersRemaining:         100,
		StoresRemaining:        10,
		ProductsRemaining:      1000,
		TransactionsRemaining:  10000,
	}, nil
}
