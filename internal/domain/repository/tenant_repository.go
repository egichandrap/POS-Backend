package repository

import (
	"context"
	"time"

	"github.com/example/jwt-ddd-clean/internal/domain/model"
)

// TenantRepository defines the interface for tenant persistence operations
type TenantRepository interface {
	// Create creates a new tenant
	Create(ctx context.Context, tenant *model.Tenant) error

	// GetByID retrieves a tenant by ID
	GetByID(ctx context.Context, id string) (*model.Tenant, error)

	// GetByCompanySlug retrieves a tenant by company slug
	GetByCompanySlug(ctx context.Context, slug string) (*model.Tenant, error)

	// GetByDomain retrieves a tenant by custom domain
	GetByDomain(ctx context.Context, domain string) (*model.Tenant, error)

	// GetByUserID retrieves a tenant by user ID (assuming user belongs to a tenant)
	GetByUserID(ctx context.Context, userID string) (*model.Tenant, error)

	// List retrieves tenants with optional filtering
	List(ctx context.Context, filter TenantFilter) ([]*model.Tenant, error)

	// Update updates an existing tenant
	Update(ctx context.Context, tenant *model.Tenant) error

	// Delete soft deletes a tenant
	Delete(ctx context.Context, id string) error

	// ExistsByCompanySlug checks if a tenant with the given slug exists
	ExistsByCompanySlug(ctx context.Context, slug string) (bool, error)

	// ExistsByDomain checks if a tenant with the given domain exists
	ExistsByDomain(ctx context.Context, domain string) (bool, error)

	// Count returns the total number of tenants
	Count(ctx context.Context) (int64, error)
}

// SubscriptionPlanRepository defines the interface for subscription plan persistence operations
type SubscriptionPlanRepository interface {
	// GetByID retrieves a subscription plan by ID
	GetByID(ctx context.Context, id string) (*model.SubscriptionPlanDetail, error)

	// List retrieves all active subscription plans
	List(ctx context.Context) ([]*model.SubscriptionPlanDetail, error)

	// ListActive retrieves only active subscription plans
	ListActive(ctx context.Context) ([]*model.SubscriptionPlanDetail, error)
}

// RawMaterialRepository defines the interface for raw material persistence operations
type RawMaterialRepository interface {
	// Create creates a new raw material
	Create(ctx context.Context, material *model.RawMaterial) error

	// GetByID retrieves a raw material by ID
	GetByID(ctx context.Context, id string) (*model.RawMaterial, error)

	// GetBySKU retrieves a raw material by SKU within a tenant
	GetBySKU(ctx context.Context, tenantID, sku string) (*model.RawMaterial, error)

	// List retrieves raw materials with optional filtering
	List(ctx context.Context, filter RawMaterialFilter) ([]*model.RawMaterial, error)

	// Update updates an existing raw material
	Update(ctx context.Context, material *model.RawMaterial) error

	// Delete deletes a raw material
	Delete(ctx context.Context, id string) error

	// UpdateStock updates the stock quantity of a raw material
	UpdateStock(ctx context.Context, id string, quantity float64) error

	// AdjustStock adjusts stock by adding/subtracting quantity
	AdjustStock(ctx context.Context, id string, delta float64) error

	// GetLowStockMaterials retrieves materials below minimum stock
	GetLowStockMaterials(ctx context.Context, tenantID string) ([]*model.RawMaterial, error)

	// GetOutOfStockMaterials retrieves materials with zero stock
	GetOutOfStockMaterials(ctx context.Context, tenantID string) ([]*model.RawMaterial, error)

	// ExistsBySKU checks if a raw material with the given SKU exists for a tenant
	ExistsBySKU(ctx context.Context, tenantID, sku string) (bool, error)

	// Count returns the total number of raw materials for a tenant
	Count(ctx context.Context, tenantID string) (int64, error)
}

// ProductRecipeRepository defines the interface for product recipe persistence operations
type ProductRecipeRepository interface {
	// Create creates a new product recipe
	Create(ctx context.Context, recipe *model.ProductRecipe) error

	// GetByID retrieves a product recipe by ID
	GetByID(ctx context.Context, id string) (*model.ProductRecipe, error)

	// GetByProductID retrieves all recipes for a specific product
	GetByProductID(ctx context.Context, inventoryID string) ([]*model.ProductRecipe, error)

	// GetByProductAndMaterial retrieves a specific recipe for product-material combination
	GetByProductAndMaterial(ctx context.Context, inventoryID, rawMaterialID string) (*model.ProductRecipe, error)

	// List retrieves product recipes with optional filtering
	List(ctx context.Context, filter ProductRecipeFilter) ([]*model.ProductRecipe, error)

	// Update updates an existing product recipe
	Update(ctx context.Context, recipe *model.ProductRecipe) error

	// Delete deletes a product recipe
	Delete(ctx context.Context, id string) error

	// DeleteByProductID deletes all recipes for a specific product
	DeleteByProductID(ctx context.Context, inventoryID string) error

	// GetProductAvailability checks if a product can be produced based on material availability
	GetProductAvailability(ctx context.Context, inventoryID string) (*model.ProductAvailability, error)

	// GetBatchProductAvailability checks availability for multiple products
	GetBatchProductAvailability(ctx context.Context, inventoryIDs []string) ([]*model.ProductAvailability, error)

	// GetMaterialsNeededForProduction calculates materials needed to produce a quantity
	GetMaterialsNeededForProduction(ctx context.Context, inventoryID string, quantity int) ([]model.MaterialAvailabilityStatus, error)
}

// TenantFilter defines filter options for listing tenants
type TenantFilter struct {
	SubscriptionStatus   model.SubscriptionStatus
	SubscriptionPlanID  string
	IsActive            *bool
	Search              string
	Limit               int
	Offset              int
}

// RawMaterialFilter defines filter options for listing raw materials
type RawMaterialFilter struct {
	TenantID    string
	IsActive    *bool
	Search      string
	LowStock    bool
	OutOfStock  bool
	Supplier    string
	Limit       int
	Offset      int
}

// ProductRecipeFilter defines filter options for listing product recipes
type ProductRecipeFilter struct {
	TenantID      string
	InventoryID   string
	RawMaterialID string
	IsActive      *bool
	Limit         int
	Offset        int
}

// TenantUsageRepository defines the interface for tenant usage tracking
type TenantUsageRepository interface {
	// GetByTenantID retrieves usage stats for a tenant
	GetByTenantID(ctx context.Context, tenantID string) (*TenantUsage, error)

	// Create creates usage tracking for a tenant
	Create(ctx context.Context, usage *TenantUsage) error

	// Update updates usage stats
	Update(ctx context.Context, usage *TenantUsage) error

	// IncrementUserCount increments the user count for a tenant
	IncrementUserCount(ctx context.Context, tenantID string) error

	// DecrementUserCount decrements the user count for a tenant
	DecrementUserCount(ctx context.Context, tenantID string) error

	// IncrementStoreCount increments the store count for a tenant
	IncrementStoreCount(ctx context.Context, tenantID string) error

	// DecrementStoreCount decrements the store count for a tenant
	DecrementStoreCount(ctx context.Context, tenantID string) error

	// IncrementProductCount increments the product count for a tenant
	IncrementProductCount(ctx context.Context, tenantID string) error

	// DecrementProductCount decrements the product count for a tenant
	DecrementProductCount(ctx context.Context, tenantID string) error

	// IncrementTransactionCount increments the daily transaction count
	IncrementTransactionCount(ctx context.Context, tenantID string) error

	// ResetDailyTransactions resets daily transaction count (called at midnight)
	ResetDailyTransactions(ctx context.Context, tenantID string) error

	// CheckLimits checks if tenant is within subscription limits
	CheckLimits(ctx context.Context, tenantID string) (*LimitStatus, error)
}

// TenantUsage represents usage statistics for a tenant
type TenantUsage struct {
	ID                 string
	TenantID           string
	CurrentUsers       int
	CurrentStores      int
	CurrentProducts    int
	TransactionsToday  int
	LastResetDate      time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// LimitStatus represents the status of tenant limits
type LimitStatus struct {
	WithinUserLimit       bool
	WithinStoreLimit      bool
	WithinProductLimit    bool
	WithinTransactionLimit bool
	UsersRemaining        int
	StoresRemaining       int
	ProductsRemaining     int
	TransactionsRemaining int
}

// SubscriptionInvoiceRepository defines the interface for subscription invoice operations
type SubscriptionInvoiceRepository interface {
	// Create creates a new invoice
	Create(ctx context.Context, invoice *SubscriptionInvoice) error

	// GetByID retrieves an invoice by ID
	GetByID(ctx context.Context, id string) (*SubscriptionInvoice, error)

	// GetByInvoiceNumber retrieves an invoice by invoice number
	GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*SubscriptionInvoice, error)

	// GetByTenantID retrieves invoices for a tenant
	GetByTenantID(ctx context.Context, tenantID string, filter InvoiceFilter) ([]*SubscriptionInvoice, error)

	// Update updates an invoice
	Update(ctx context.Context, invoice *SubscriptionInvoice) error

	// ListPending retrieves all pending invoices
	ListPending(ctx context.Context) ([]*SubscriptionInvoice, error)

	// ListOverdue retrieves all overdue invoices
	ListOverdue(ctx context.Context) ([]*SubscriptionInvoice, error)
}

// SubscriptionInvoice represents a subscription invoice
type SubscriptionInvoice struct {
	ID                   string
	TenantID             string
	InvoiceNumber        string
	SubscriptionPlanID   string
	Amount               float64
	Currency             string
	Status               string
	DueDate              time.Time
	PaidAt               *time.Time
	PaymentMethod        string
	PaymentReference     string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// InvoiceFilter defines filter options for listing invoices
type InvoiceFilter struct {
	Status   string
	StartDate time.Time
	EndDate   time.Time
	Limit    int
	Offset   int
}
