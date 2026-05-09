package dto

import (
	"time"

	"github.com/example/jwt-ddd-clean/internal/domain/model"
)

// Tenant Registration and Management DTOs

// RegisterTenantRequest represents a request to register a new tenant/company
type RegisterTenantRequest struct {
	CompanyName        string `json:"company_name"`
	CompanySlug        string `json:"company_slug"`
	Domain             string `json:"domain,omitempty"`
	Email              string `json:"email"`
	Phone              string `json:"phone,omitempty"`
	Address            string `json:"address,omitempty"`
	SubscriptionPlanID string `json:"subscription_plan_id,omitempty"`
	AdminUser          AdminUserCreate `json:"admin_user"`
}

// AdminUserCreate represents admin user creation during tenant registration
type AdminUserCreate struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

// TenantResponse represents a tenant response
type TenantResponse struct {
	ID                    string                     `json:"id"`
	CompanyName           string                     `json:"company_name"`
	CompanySlug           string                     `json:"company_slug"`
	Domain                string                     `json:"domain,omitempty"`
	Email                 string                     `json:"email"`
	Phone                 string                     `json:"phone,omitempty"`
	Address               string                     `json:"address,omitempty"`
	LogoURL               string                     `json:"logo_url,omitempty"`
	SubscriptionPlanID    string                     `json:"subscription_plan_id"`
	SubscriptionStatus    model.SubscriptionStatus   `json:"subscription_status"`
	TrialEndsAt           *time.Time                 `json:"trial_ends_at,omitempty"`
	SubscriptionStartsAt  *time.Time                 `json:"subscription_starts_at,omitempty"`
	SubscriptionEndsAt    *time.Time                 `json:"subscription_ends_at,omitempty"`
	IsActive              bool                       `json:"is_active"`
	Settings              map[string]interface{}     `json:"settings,omitempty"`
	CreatedAt             time.Time                  `json:"created_at"`
	UpdatedAt             time.Time                  `json:"updated_at"`
	Usage                 *TenantUsageResponse       `json:"usage,omitempty"`
	Subscription          *SubscriptionPlanResponse  `json:"subscription,omitempty"`
}

// TenantListResponse represents a paginated list of tenants
type TenantListResponse struct {
	Tenants []TenantResponse `json:"tenants"`
	Total   int64            `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

// UpdateTenantRequest represents a request to update tenant information
type UpdateTenantRequest struct {
	CompanyName string `json:"company_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Address     string `json:"address,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
	Domain      string `json:"domain,omitempty"`
}

// UpdateTenantSettingsRequest represents a request to update tenant settings
type UpdateTenantSettingsRequest struct {
	Settings map[string]interface{} `json:"settings"`
}

// SubscriptionPlanResponse represents a subscription plan response
type SubscriptionPlanResponse struct {
	ID                    string                   `json:"id"`
	Name                  string                   `json:"name"`
	Description           string                   `json:"description"`
	PriceMonthly          float64                  `json:"price_monthly"`
	PriceYearly           float64                  `json:"price_yearly"`
	MaxUsers              int                      `json:"max_users"`
	MaxStores             int                      `json:"max_stores"`
	MaxProducts           int                      `json:"max_products"`
	MaxTransactionsPerDay int                      `json:"max_transactions_per_day"`
	Features              model.SubscriptionFeatures `json:"features"`
	IsActive              bool                     `json:"is_active"`
}

// SubscriptionPlanListResponse represents a list of subscription plans
type SubscriptionPlanListResponse struct {
	Plans []SubscriptionPlanResponse `json:"plans"`
}

// UpgradeSubscriptionRequest represents a request to upgrade/downgrade subscription
type UpgradeSubscriptionRequest struct {
	PlanID          string `json:"plan_id"`
	PaymentMethod   string `json:"payment_method,omitempty"`
	Duration        string `json:"duration"` // "monthly" or "yearly"
}

// SubscriptionResponse represents subscription details
type SubscriptionResponse struct {
	PlanID            string                    `json:"plan_id"`
	PlanName          string                    `json:"plan_name"`
	Status            model.SubscriptionStatus  `json:"status"`
	TrialEndsAt       *time.Time                `json:"trial_ends_at,omitempty"`
	CurrentPeriodStart *time.Time               `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time               `json:"current_period_end,omitempty"`
	WillRenew         bool                      `json:"will_renew"`
	Usage             TenantUsageResponse       `json:"usage"`
	Limits            LimitStatusResponse       `json:"limits"`
}

// TenantUsageResponse represents tenant usage statistics
type TenantUsageResponse struct {
	CurrentUsers          int       `json:"current_users"`
	CurrentStores         int       `json:"current_stores"`
	CurrentProducts       int       `json:"current_products"`
	TransactionsToday     int       `json:"transactions_today"`
	LastResetDate         time.Time `json:"last_reset_date"`
}

// LimitStatusResponse represents subscription limit status
type LimitStatusResponse struct {
	WithinUserLimit        bool `json:"within_user_limit"`
	WithinStoreLimit       bool `json:"within_store_limit"`
	WithinProductLimit     bool `json:"within_product_limit"`
	WithinTransactionLimit bool `json:"within_transaction_limit"`
	UsersRemaining         int  `json:"users_remaining,omitempty"`
	StoresRemaining        int  `json:"stores_remaining,omitempty"`
	ProductsRemaining      int  `json:"products_remaining,omitempty"`
	TransactionsRemaining  int  `json:"transactions_remaining,omitempty"`
}

// Raw Material DTOs

// CreateRawMaterialRequest represents a request to create a raw material
type CreateRawMaterialRequest struct {
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Unit        string  `json:"unit"`
	Quantity    float64 `json:"quantity"`
	MinStock    float64 `json:"min_stock,omitempty"`
	CostPerUnit float64 `json:"cost_per_unit,omitempty"`
	Supplier    string  `json:"supplier,omitempty"`
	Location    string  `json:"location,omitempty"`
}

// UpdateRawMaterialRequest represents a request to update a raw material
type UpdateRawMaterialRequest struct {
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Unit        string  `json:"unit,omitempty"`
	MinStock    float64 `json:"min_stock,omitempty"`
	CostPerUnit float64 `json:"cost_per_unit,omitempty"`
	Supplier    string  `json:"supplier,omitempty"`
	Location    string  `json:"location,omitempty"`
}

// RawMaterialResponse represents a raw material response
type RawMaterialResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Unit        string    `json:"unit"`
	Quantity    float64   `json:"quantity"`
	MinStock    float64   `json:"min_stock"`
	CostPerUnit float64   `json:"cost_per_unit"`
	Supplier    string    `json:"supplier,omitempty"`
	Location    string    `json:"location,omitempty"`
	IsLowStock  bool      `json:"is_low_stock"`
	IsOutOfStock bool     `json:"is_out_of_stock"`
	StockStatus string    `json:"stock_status"`
	TotalCost   float64   `json:"total_cost"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RawMaterialListResponse represents a paginated list of raw materials
type RawMaterialListResponse struct {
	Materials []RawMaterialResponse `json:"materials"`
	Total     int64                 `json:"total"`
	Limit     int                   `json:"limit"`
	Offset    int                   `json:"offset"`
}

// AdjustRawMaterialStockRequest represents a request to adjust raw material stock
type AdjustRawMaterialStockRequest struct {
	Quantity float64 `json:"quantity"` // Positive to add, negative to reduce
	Reason   string  `json:"reason,omitempty"`
}

// Product Recipe DTOs

// CreateProductRecipeRequest represents a request to create a product recipe
type CreateProductRecipeRequest struct {
	InventoryID      string  `json:"inventory_id"`
	RawMaterialID    string  `json:"raw_material_id"`
	QuantityRequired float64 `json:"quantity_required"`
}

// UpdateProductRecipeRequest represents a request to update a product recipe
type UpdateProductRecipeRequest struct {
	QuantityRequired float64 `json:"quantity_required"`
}

// ProductRecipeResponse represents a product recipe response
type ProductRecipeResponse struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	InventoryID      string    `json:"inventory_id"`
	RawMaterialID    string    `json:"raw_material_id"`
	RawMaterialName  string    `json:"raw_material_name,omitempty"`
	RawMaterialSKU   string    `json:"raw_material_sku,omitempty"`
	QuantityRequired float64   `json:"quantity_required"`
	Unit             string    `json:"unit,omitempty"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ProductRecipeListResponse represents a list of product recipes
type ProductRecipeListResponse struct {
	Recipes []ProductRecipeResponse `json:"recipes"`
	Total   int64                   `json:"total"`
}

// SetProductRecipeRequest represents a request to set all recipes for a product
type SetProductRecipeRequest struct {
	Recipes []CreateProductRecipeRequest `json:"recipes"`
}

// Product Availability DTOs

// MaterialAvailabilityStatusResponse represents material availability status
type MaterialAvailabilityStatusResponse struct {
	RawMaterialID     string  `json:"raw_material_id"`
	RawMaterialName   string  `json:"raw_material_name"`
	RawMaterialSKU    string  `json:"raw_material_sku"`
	RequiredQuantity  float64 `json:"required"`
	AvailableQuantity float64 `json:"available"`
	Unit              string  `json:"unit"`
	IsSufficient      bool    `json:"is_sufficient"`
	Shortage          float64 `json:"shortage,omitempty"`
}

// ProductAvailabilityResponse represents product availability status
type ProductAvailabilityResponse struct {
	InventoryID          string                                   `json:"inventory_id"`
	TenantID             string                                   `json:"tenant_id"`
	ProductName          string                                   `json:"product_name"`
	ProductSKU           string                                   `json:"product_sku"`
	ProductQuantity      int                                      `json:"product_quantity"`
	TotalIngredients     int                                      `json:"total_ingredients"`
	AvailableIngredients int                                      `json:"available_ingredients"`
	IsAvailable          bool                                     `json:"is_available"`
	MaterialsStatus      []MaterialAvailabilityStatusResponse      `json:"materials_status"`
	MissingMaterials     []MaterialAvailabilityStatusResponse      `json:"missing_materials,omitempty"`
	CanProduceQuantity   int                                      `json:"can_produce_quantity"`
}

// Helper functions to convert domain models to DTOs

func ToTenantResponse(tenant *model.Tenant) TenantResponse {
	return TenantResponse{
		ID:                 tenant.ID(),
		CompanyName:        tenant.CompanyName(),
		CompanySlug:        tenant.CompanySlug(),
		Domain:             tenant.Domain(),
		Email:              tenant.Email(),
		Phone:              tenant.Phone(),
		Address:            tenant.Address(),
		LogoURL:            tenant.LogoURL(),
		SubscriptionPlanID: tenant.SubscriptionPlanID(),
		SubscriptionStatus: tenant.SubscriptionStatus(),
		TrialEndsAt:        tenant.TrialEndsAt(),
		SubscriptionStartsAt: tenant.SubscriptionStartsAt(),
		SubscriptionEndsAt:   tenant.SubscriptionEndsAt(),
		IsActive:          tenant.IsActive(),
		Settings:          tenant.Settings(),
		CreatedAt:         tenant.CreatedAt(),
		UpdatedAt:         tenant.UpdatedAt(),
	}
}

func ToSubscriptionPlanResponse(plan *model.SubscriptionPlanDetail) SubscriptionPlanResponse {
	return SubscriptionPlanResponse{
		ID:                    plan.ID(),
		Name:                  plan.Name(),
		Description:           plan.Description(),
		PriceMonthly:          plan.PriceMonthly(),
		PriceYearly:           plan.PriceYearly(),
		MaxUsers:              plan.MaxUsers(),
		MaxStores:             plan.MaxStores(),
		MaxProducts:           plan.MaxProducts(),
		MaxTransactionsPerDay: plan.MaxTransactionsPerDay(),
		Features:              plan.Features(),
		IsActive:              plan.IsActive(),
	}
}

func ToRawMaterialResponse(material *model.RawMaterial) RawMaterialResponse {
	return RawMaterialResponse{
		ID:           material.ID(),
		TenantID:     material.TenantID(),
		SKU:          material.SKU(),
		Name:         material.Name(),
		Description:  material.Description(),
		Unit:         material.Unit(),
		Quantity:     material.Quantity(),
		MinStock:     material.MinStock(),
		CostPerUnit:  material.CostPerUnit(),
		Supplier:     material.Supplier(),
		Location:     material.Location(),
		IsLowStock:   material.IsLowStock(),
		IsOutOfStock: material.IsOutOfStock(),
		StockStatus:  material.StockStatus(),
		TotalCost:    material.GetTotalCost(),
		IsActive:     material.IsActive(),
		CreatedAt:    material.CreatedAt(),
		UpdatedAt:    material.UpdatedAt(),
	}
}

func ToProductRecipeResponse(recipe *model.ProductRecipe) ProductRecipeResponse {
	return ProductRecipeResponse{
		ID:               recipe.ID(),
		TenantID:         recipe.TenantID(),
		InventoryID:      recipe.InventoryID(),
		RawMaterialID:    recipe.RawMaterialID(),
		QuantityRequired: recipe.QuantityRequired(),
		IsActive:         recipe.IsActive(),
		CreatedAt:        recipe.CreatedAt(),
		UpdatedAt:        recipe.UpdatedAt(),
	}
}

func ToProductAvailabilityResponse(availability *model.ProductAvailability) ProductAvailabilityResponse {
	materialsStatus := make([]MaterialAvailabilityStatusResponse, len(availability.MaterialsStatus))
	for i, m := range availability.MaterialsStatus {
		materialsStatus[i] = MaterialAvailabilityStatusResponse{
			RawMaterialID:     m.RawMaterialID,
			RawMaterialName:   m.RawMaterialName,
			RawMaterialSKU:    m.RawMaterialSKU,
			RequiredQuantity:  m.RequiredQuantity,
			AvailableQuantity: m.AvailableQuantity,
			Unit:              m.Unit,
			IsSufficient:      m.IsSufficient,
			Shortage:          m.Shortage,
		}
	}

	missingMaterials := make([]MaterialAvailabilityStatusResponse, 0)
	for _, m := range availability.MaterialsStatus {
		if !m.IsSufficient {
			missingMaterials = append(missingMaterials, MaterialAvailabilityStatusResponse{
				RawMaterialID:     m.RawMaterialID,
				RawMaterialName:   m.RawMaterialName,
				RawMaterialSKU:    m.RawMaterialSKU,
				RequiredQuantity:  m.RequiredQuantity,
				AvailableQuantity: m.AvailableQuantity,
				Unit:              m.Unit,
				IsSufficient:      m.IsSufficient,
				Shortage:          m.Shortage,
			})
		}
	}

	// Calculate how many can be produced
	canProduceQuantity := -1 // Unlimited if no ingredients
	if availability.TotalIngredients > 0 {
		canProduceQuantity = availability.ProductQuantity // Default to product quantity
		for _, m := range availability.MaterialsStatus {
			if m.RequiredQuantity > 0 {
				possibleFromMaterial := int(m.AvailableQuantity / m.RequiredQuantity)
				if possibleFromMaterial < canProduceQuantity {
					canProduceQuantity = possibleFromMaterial
				}
			}
		}
	}

	return ProductAvailabilityResponse{
		InventoryID:          availability.InventoryID,
		TenantID:             availability.TenantID,
		ProductName:          availability.ProductName,
		ProductSKU:           availability.ProductSKU,
		ProductQuantity:      availability.ProductQuantity,
		TotalIngredients:     availability.TotalIngredients,
		AvailableIngredients: availability.AvailableIngredients,
		IsAvailable:          availability.IsAvailable,
		MaterialsStatus:      materialsStatus,
		MissingMaterials:     missingMaterials,
		CanProduceQuantity:   canProduceQuantity,
	}
}
