package usecase

import (
	"context"
	"time"

	"github.com/example/jwt-ddd-clean/internal/application/dto"
	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
	"github.com/example/jwt-ddd-clean/internal/domain/valueobject"
	"github.com/example/jwt-ddd-clean/internal/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// TenantUsecase defines the interface for tenant use cases
type TenantUsecase interface {
	// Tenant registration and management
	RegisterTenant(ctx context.Context, req dto.RegisterTenantRequest) (*dto.TenantResponse, error)
	GetTenantByID(ctx context.Context, id string) (*dto.TenantResponse, error)
	GetTenantBySlug(ctx context.Context, slug string) (*dto.TenantResponse, error)
	GetTenantByDomain(ctx context.Context, domain string) (*dto.TenantResponse, error)
	UpdateTenant(ctx context.Context, id string, req dto.UpdateTenantRequest) (*dto.TenantResponse, error)
	UpdateTenantSettings(ctx context.Context, id string, req dto.UpdateTenantSettingsRequest) error
	ListTenants(ctx context.Context, filter repository.TenantFilter) (*dto.TenantListResponse, error)
	DeactivateTenant(ctx context.Context, id string) error
	ActivateTenant(ctx context.Context, id string) error

	// Subscription management
	GetSubscriptionPlans(ctx context.Context) (*dto.SubscriptionPlanListResponse, error)
	GetSubscriptionPlan(ctx context.Context, planID string) (*dto.SubscriptionPlanResponse, error)
	UpgradeSubscription(ctx context.Context, tenantID string, req dto.UpgradeSubscriptionRequest) (*dto.SubscriptionResponse, error)
	GetSubscriptionStatus(ctx context.Context, tenantID string) (*dto.SubscriptionResponse, error)
	CancelSubscription(ctx context.Context, tenantID string) error

	// Raw material management
	CreateRawMaterial(ctx context.Context, tenantID string, req dto.CreateRawMaterialRequest) (*dto.RawMaterialResponse, error)
	GetRawMaterial(ctx context.Context, id string) (*dto.RawMaterialResponse, error)
	ListRawMaterials(ctx context.Context, filter repository.RawMaterialFilter) (*dto.RawMaterialListResponse, error)
	UpdateRawMaterial(ctx context.Context, id string, req dto.UpdateRawMaterialRequest) (*dto.RawMaterialResponse, error)
	DeleteRawMaterial(ctx context.Context, id string) error
	AdjustRawMaterialStock(ctx context.Context, id string, req dto.AdjustRawMaterialStockRequest) error
	GetLowStockMaterials(ctx context.Context, tenantID string) (*dto.RawMaterialListResponse, error)
	GetOutOfStockMaterials(ctx context.Context, tenantID string) (*dto.RawMaterialListResponse, error)

	// Product recipe management
	CreateProductRecipe(ctx context.Context, tenantID string, req dto.CreateProductRecipeRequest) (*dto.ProductRecipeResponse, error)
	GetProductRecipes(ctx context.Context, inventoryID string) (*dto.ProductRecipeListResponse, error)
	SetProductRecipes(ctx context.Context, tenantID, inventoryID string, req dto.SetProductRecipeRequest) error
	UpdateProductRecipe(ctx context.Context, id string, req dto.UpdateProductRecipeRequest) (*dto.ProductRecipeResponse, error)
	DeleteProductRecipe(ctx context.Context, id string) error
	DeleteProductRecipes(ctx context.Context, inventoryID string) error

	// Product availability
	GetProductAvailability(ctx context.Context, inventoryID string) (*dto.ProductAvailabilityResponse, error)
	GetBatchProductAvailability(ctx context.Context, inventoryIDs []string) ([]*dto.ProductAvailabilityResponse, error)
	CheckCanProduce(ctx context.Context, inventoryID string, quantity int) (bool, []*dto.MaterialAvailabilityStatusResponse, error)
}

type tenantUsecase struct {
	tenantRepo       repository.TenantRepository
	planRepo         repository.SubscriptionPlanRepository
	rawMaterialRepo  repository.RawMaterialRepository
	recipeRepo       repository.ProductRecipeRepository
	usageRepo        repository.TenantUsageRepository
	userRepo         repository.UserRepository
}

// NewTenantUsecase creates a new TenantUsecase
func NewTenantUsecase(
	tenantRepo repository.TenantRepository,
	planRepo repository.SubscriptionPlanRepository,
	rawMaterialRepo repository.RawMaterialRepository,
	recipeRepo repository.ProductRecipeRepository,
	usageRepo repository.TenantUsageRepository,
	userRepo repository.UserRepository,
) TenantUsecase {
	return &tenantUsecase{
		tenantRepo:      tenantRepo,
		planRepo:        planRepo,
		rawMaterialRepo: rawMaterialRepo,
		recipeRepo:      recipeRepo,
		usageRepo:       usageRepo,
		userRepo:        userRepo,
	}
}

// Tenant Registration and Management

func (u *tenantUsecase) RegisterTenant(ctx context.Context, req dto.RegisterTenantRequest) (*dto.TenantResponse, error) {
	// Validate company slug
	exists, err := u.tenantRepo.ExistsByCompanySlug(ctx, req.CompanySlug)
	if err != nil {
		return nil, errors.NewInternalError("gagal memeriksa slug perusahaan")
	}
	if exists {
		return nil, errors.NewValidationError("slug perusahaan telah digunakan")
	}

	// Validate domain if provided
	if req.Domain != "" {
		exists, err = u.tenantRepo.ExistsByDomain(ctx, req.Domain)
		if err != nil {
			return nil, errors.NewInternalError("gagal memeriksa domain")
		}
		if exists {
			return nil, errors.NewValidationError("domain telah digunakan")
		}
	}

	// Set default plan if not specified
	planID := req.SubscriptionPlanID
	if planID == "" {
		planID = string(model.PlanPlus)
	}

	// Validate plan exists
	plan, err := u.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, errors.NewValidationError("paket langganan tidak valid")
	}
	if !plan.IsActive() {
		return nil, errors.NewValidationError("paket langganan tidak aktif")
	}

	// Create tenant
	tenant, err := model.NewTenant(
		req.CompanyName,
		req.CompanySlug,
		req.Email,
		planID,
		"", // createdBy will be set after admin user creation
	)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// Set optional fields
	if req.Domain != "" {
		tenant.UpdateDomain(req.Domain)
	}
	if req.Phone != "" {
		tenant.UpdateProfile(req.CompanyName, req.Email, req.Phone, req.Address, "")
	}

	// Create tenant
	err = u.tenantRepo.Create(ctx, tenant)
	if err != nil {
		return nil, errors.NewInternalError("gagal membuat tenant: %v", err)
	}

	// Create admin user
	email, err := valueobject.NewEmail(req.AdminUser.Email)
	if err != nil {
		return nil, errors.NewValidationError("email admin tidak valid")
	}

	password, err := valueobject.NewPassword(req.AdminUser.Password)
	if err != nil {
		return nil, errors.NewValidationError("password admin tidak valid")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password.String()), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.NewInternalError("gagal memproses password")
	}

	adminUser, err := model.NewUser(
		req.AdminUser.Username,
		email,
		valueobject.Password(hashedPassword),
		req.AdminUser.FullName,
		model.RoleAdmin,
	)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// Set tenant ID for user
	adminUser.SetTenantID(tenant.ID())

	// Create admin user
	err = u.userRepo.Create(ctx, adminUser)
	if err != nil {
		return nil, errors.NewInternalError("gagal membuat user admin: %v", err)
	}

	response := dto.ToTenantResponse(tenant)
	return &response, nil
}

func (u *tenantUsecase) GetTenantByID(ctx context.Context, id string) (*dto.TenantResponse, error) {
	tenant, err := u.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("tenant", "id", id)
	}

	response := dto.ToTenantResponse(tenant)
	return &response, nil
}

func (u *tenantUsecase) GetTenantBySlug(ctx context.Context, slug string) (*dto.TenantResponse, error) {
	tenant, err := u.tenantRepo.GetByCompanySlug(ctx, slug)
	if err != nil {
		return nil, errors.NewNotFoundError("tenant", "slug", slug)
	}

	response := dto.ToTenantResponse(tenant)
	return &response, nil
}

func (u *tenantUsecase) GetTenantByDomain(ctx context.Context, domain string) (*dto.TenantResponse, error) {
	tenant, err := u.tenantRepo.GetByDomain(ctx, domain)
	if err != nil {
		return nil, errors.NewNotFoundError("tenant", "domain", domain)
	}

	response := dto.ToTenantResponse(tenant)
	return &response, nil
}

func (u *tenantUsecase) UpdateTenant(ctx context.Context, id string, req dto.UpdateTenantRequest) (*dto.TenantResponse, error) {
	tenant, err := u.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("tenant", "id", id)
	}

	// Update profile
	err = tenant.UpdateProfile(
		req.CompanyName,
		req.Email,
		req.Phone,
		req.Address,
		req.LogoURL,
	)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// Update domain if provided
	if req.Domain != "" {
		// Check if domain is already taken by another tenant
		exists, err := u.tenantRepo.ExistsByDomain(ctx, req.Domain)
		if err != nil {
			return nil, errors.NewInternalError("gagal memeriksa domain")
		}
		if exists && tenant.Domain() != req.Domain {
			return nil, errors.NewValidationError("domain telah digunakan")
		}
		tenant.UpdateDomain(req.Domain)
	}

	err = u.tenantRepo.Update(ctx, tenant)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengupdate tenant: %v", err)
	}

	response := dto.ToTenantResponse(tenant)
	return &response, nil
}

func (u *tenantUsecase) UpdateTenantSettings(ctx context.Context, id string, req dto.UpdateTenantSettingsRequest) error {
	tenant, err := u.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("tenant", "id", id)
	}

	err = tenant.UpdateSettings(req.Settings)
	if err != nil {
		return errors.NewValidationError(err.Error())
	}

	err = u.tenantRepo.Update(ctx, tenant)
	if err != nil {
		return errors.NewInternalError("gagal mengupdate settings: %v", err)
	}

	return nil
}

func (u *tenantUsecase) ListTenants(ctx context.Context, filter repository.TenantFilter) (*dto.TenantListResponse, error) {
	tenants, err := u.tenantRepo.List(ctx, filter)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengambil daftar tenant: %v", err)
	}

	total, err := u.tenantRepo.Count(ctx)
	if err != nil {
		return nil, errors.NewInternalError("gagal menghitung tenant: %v", err)
	}

	responses := make([]dto.TenantResponse, len(tenants))
	for i, tenant := range tenants {
		responses[i] = dto.ToTenantResponse(tenant)
	}

	return &dto.TenantListResponse{
		Tenants: responses,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
	}, nil
}

func (u *tenantUsecase) DeactivateTenant(ctx context.Context, id string) error {
	tenant, err := u.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("tenant", "id", id)
	}

	tenant.Deactivate()
	err = u.tenantRepo.Update(ctx, tenant)
	if err != nil {
		return errors.NewInternalError("gagal menonaktifkan tenant: %v", err)
	}

	return nil
}

func (u *tenantUsecase) ActivateTenant(ctx context.Context, id string) error {
	tenant, err := u.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("tenant", "id", id)
	}

	tenant.Activate()
	err = u.tenantRepo.Update(ctx, tenant)
	if err != nil {
		return errors.NewInternalError("gagal mengaktifkan tenant: %v", err)
	}

	return nil
}

// Subscription Management

func (u *tenantUsecase) GetSubscriptionPlans(ctx context.Context) (*dto.SubscriptionPlanListResponse, error) {
	plans, err := u.planRepo.ListActive(ctx)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengambil paket langganan: %v", err)
	}

	responses := make([]dto.SubscriptionPlanResponse, len(plans))
	for i, plan := range plans {
		responses[i] = dto.ToSubscriptionPlanResponse(plan)
	}

	return &dto.SubscriptionPlanListResponse{
		Plans: responses,
	}, nil
}

func (u *tenantUsecase) GetSubscriptionPlan(ctx context.Context, planID string) (*dto.SubscriptionPlanResponse, error) {
	plan, err := u.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, errors.NewNotFoundError("subscription_plan", "id", planID)
	}

	response := dto.ToSubscriptionPlanResponse(plan)
	return &response, nil
}

func (u *tenantUsecase) UpgradeSubscription(ctx context.Context, tenantID string, req dto.UpgradeSubscriptionRequest) (*dto.SubscriptionResponse, error) {
	tenant, err := u.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, errors.NewNotFoundError("tenant", "id", tenantID)
	}

	plan, err := u.planRepo.GetByID(ctx, req.PlanID)
	if err != nil {
		return nil, errors.NewNotFoundError("subscription_plan", "id", req.PlanID)
	}
	if !plan.IsActive() {
		return nil, errors.NewValidationError("paket langganan tidak aktif")
	}

	now := time.Now()
	var endsAt time.Time

	if req.Duration == "yearly" {
		endsAt = now.AddDate(1, 0, 0)
	} else {
		endsAt = now.AddDate(0, 1, 0)
	}

	err = tenant.ActivateSubscription(req.PlanID, now, endsAt)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengaktifkan langganan: %v", err)
	}

	err = u.tenantRepo.Update(ctx, tenant)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengupdate tenant: %v", err)
	}

	return u.GetSubscriptionStatus(ctx, tenantID)
}

func (u *tenantUsecase) GetSubscriptionStatus(ctx context.Context, tenantID string) (*dto.SubscriptionResponse, error) {
	tenant, err := u.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, errors.NewNotFoundError("tenant", "id", tenantID)
	}

	plan, err := u.planRepo.GetByID(ctx, tenant.SubscriptionPlanID())
	if err != nil {
		return nil, errors.NewInternalError("gagal mengambil paket langganan")
	}

	usage, err := u.usageRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengambil penggunaan")
	}

	limits, err := u.usageRepo.CheckLimits(ctx, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengecek batas langganan")
	}

	return &dto.SubscriptionResponse{
		PlanID:             plan.ID(),
		PlanName:           plan.Name(),
		Status:             tenant.SubscriptionStatus(),
		TrialEndsAt:        tenant.TrialEndsAt(),
		CurrentPeriodStart: tenant.SubscriptionStartsAt(),
		CurrentPeriodEnd:   tenant.SubscriptionEndsAt(),
		WillRenew:          tenant.SubscriptionStatus() == model.SubscriptionStatusActive,
		Usage: dto.TenantUsageResponse{
			CurrentUsers:      usage.CurrentUsers,
			CurrentStores:     usage.CurrentStores,
			CurrentProducts:   usage.CurrentProducts,
			TransactionsToday: usage.TransactionsToday,
			LastResetDate:     usage.LastResetDate,
		},
		Limits: dto.LimitStatusResponse{
			WithinUserLimit:        limits.WithinUserLimit,
			WithinStoreLimit:       limits.WithinStoreLimit,
			WithinProductLimit:     limits.WithinProductLimit,
			WithinTransactionLimit: limits.WithinTransactionLimit,
			UsersRemaining:         limits.UsersRemaining,
			StoresRemaining:        limits.StoresRemaining,
			ProductsRemaining:      limits.ProductsRemaining,
			TransactionsRemaining:  limits.TransactionsRemaining,
		},
	}, nil
}

func (u *tenantUsecase) CancelSubscription(ctx context.Context, tenantID string) error {
	tenant, err := u.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return errors.NewNotFoundError("tenant", "id", tenantID)
	}

	tenant.CancelSubscription()
	err = u.tenantRepo.Update(ctx, tenant)
	if err != nil {
		return errors.NewInternalError("gagal membatalkan langganan: %v", err)
	}

	return nil
}

// Raw Material Management

func (u *tenantUsecase) CreateRawMaterial(ctx context.Context, tenantID string, req dto.CreateRawMaterialRequest) (*dto.RawMaterialResponse, error) {
	// Check if SKU already exists
	exists, err := u.rawMaterialRepo.ExistsBySKU(ctx, tenantID, req.SKU)
	if err != nil {
		return nil, errors.NewInternalError("gagal memeriksa SKU")
	}
	if exists {
		return nil, errors.NewValidationError("SKU telah digunakan")
	}

	material, err := model.NewRawMaterial(
		tenantID,
		req.SKU,
		req.Name,
		req.Unit,
		req.Quantity,
		req.CostPerUnit,
	)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// Set optional fields
	material.UpdateDetails(req.Name, req.Description, req.Unit, req.Supplier, req.Location, req.MinStock)

	err = u.rawMaterialRepo.Create(ctx, material)
	if err != nil {
		return nil, errors.NewInternalError("gagal membuat bahan baku: %v", err)
	}

	response := dto.ToRawMaterialResponse(material)
	return &response, nil
}

func (u *tenantUsecase) GetRawMaterial(ctx context.Context, id string) (*dto.RawMaterialResponse, error) {
	material, err := u.rawMaterialRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("raw_material", "id", id)
	}

	response := dto.ToRawMaterialResponse(material)
	return &response, nil
}

func (u *tenantUsecase) ListRawMaterials(ctx context.Context, filter repository.RawMaterialFilter) (*dto.RawMaterialListResponse, error) {
	materials, err := u.rawMaterialRepo.List(ctx, filter)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengambil daftar bahan baku: %v", err)
	}

	total, err := u.rawMaterialRepo.Count(ctx, filter.TenantID)
	if err != nil {
		return nil, errors.NewInternalError("gagal menghitung bahan baku: %v", err)
	}

	responses := make([]dto.RawMaterialResponse, len(materials))
	for i, material := range materials {
		responses[i] = dto.ToRawMaterialResponse(material)
	}

	return &dto.RawMaterialListResponse{
		Materials: responses,
		Total:     total,
		Limit:     filter.Limit,
		Offset:    filter.Offset,
	}, nil
}

func (u *tenantUsecase) UpdateRawMaterial(ctx context.Context, id string, req dto.UpdateRawMaterialRequest) (*dto.RawMaterialResponse, error) {
	material, err := u.rawMaterialRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("raw_material", "id", id)
	}

	name := material.Name()
	unit := material.Unit()
	supplier := material.Supplier()
	location := material.Location()
	minStock := material.MinStock()

	if req.Name != "" {
		name = req.Name
	}
	if req.Unit != "" {
		unit = req.Unit
	}
	if req.Supplier != "" {
		supplier = req.Supplier
	}
	if req.Location != "" {
		location = req.Location
	}
	if req.MinStock > 0 {
		minStock = req.MinStock
	}

	err = material.UpdateDetails(name, req.Description, unit, supplier, location, minStock)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	if req.CostPerUnit > 0 {
		err = material.UpdateCost(req.CostPerUnit)
		if err != nil {
			return nil, errors.NewValidationError(err.Error())
		}
	}

	err = u.rawMaterialRepo.Update(ctx, material)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengupdate bahan baku: %v", err)
	}

	response := dto.ToRawMaterialResponse(material)
	return &response, nil
}

func (u *tenantUsecase) DeleteRawMaterial(ctx context.Context, id string) error {
	err := u.rawMaterialRepo.Delete(ctx, id)
	if err != nil {
		return errors.NewInternalError("gagal menghapus bahan baku: %v", err)
	}
	return nil
}

func (u *tenantUsecase) AdjustRawMaterialStock(ctx context.Context, id string, req dto.AdjustRawMaterialStockRequest) error {
	material, err := u.rawMaterialRepo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("raw_material", "id", id)
	}

	if req.Quantity > 0 {
		err = material.AddStock(req.Quantity)
	} else if req.Quantity < 0 {
		err = material.ReduceStock(-req.Quantity)
	}
	if err != nil {
		return errors.NewValidationError(err.Error())
	}

	err = u.rawMaterialRepo.Update(ctx, material)
	if err != nil {
		return errors.NewInternalError("gagal mengupdate stok: %v", err)
	}

	return nil
}

func (u *tenantUsecase) GetLowStockMaterials(ctx context.Context, tenantID string) (*dto.RawMaterialListResponse, error) {
	materials, err := u.rawMaterialRepo.GetLowStockMaterials(ctx, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengambil bahan baku stok rendah: %v", err)
	}

	responses := make([]dto.RawMaterialResponse, len(materials))
	for i, material := range materials {
		responses[i] = dto.ToRawMaterialResponse(material)
	}

	return &dto.RawMaterialListResponse{
		Materials: responses,
		Total:     int64(len(materials)),
	}, nil
}

func (u *tenantUsecase) GetOutOfStockMaterials(ctx context.Context, tenantID string) (*dto.RawMaterialListResponse, error) {
	materials, err := u.rawMaterialRepo.GetOutOfStockMaterials(ctx, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengambil bahan baku habis: %v", err)
	}

	responses := make([]dto.RawMaterialResponse, len(materials))
	for i, material := range materials {
		responses[i] = dto.ToRawMaterialResponse(material)
	}

	return &dto.RawMaterialListResponse{
		Materials: responses,
		Total:     int64(len(materials)),
	}, nil
}

// Product Recipe Management

func (u *tenantUsecase) CreateProductRecipe(ctx context.Context, tenantID string, req dto.CreateProductRecipeRequest) (*dto.ProductRecipeResponse, error) {
	recipe, err := model.NewProductRecipe(
		tenantID,
		req.InventoryID,
		req.RawMaterialID,
		req.QuantityRequired,
	)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	err = u.recipeRepo.Create(ctx, recipe)
	if err != nil {
		return nil, errors.NewInternalError("gagal membuat resep produk: %v", err)
	}

	response := dto.ToProductRecipeResponse(recipe)
	return &response, nil
}

func (u *tenantUsecase) GetProductRecipes(ctx context.Context, inventoryID string) (*dto.ProductRecipeListResponse, error) {
	recipes, err := u.recipeRepo.GetByProductID(ctx, inventoryID)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengambil resep produk: %v", err)
	}

	responses := make([]dto.ProductRecipeResponse, len(recipes))
	for i, recipe := range recipes {
		responses[i] = dto.ToProductRecipeResponse(recipe)
	}

	return &dto.ProductRecipeListResponse{
		Recipes: responses,
		Total:   int64(len(recipes)),
	}, nil
}

func (u *tenantUsecase) SetProductRecipes(ctx context.Context, tenantID, inventoryID string, req dto.SetProductRecipeRequest) error {
	// Delete existing recipes
	err := u.recipeRepo.DeleteByProductID(ctx, inventoryID)
	if err != nil {
		return errors.NewInternalError("gagal menghapus resep lama: %v", err)
	}

	// Create new recipes
	for _, recipeReq := range req.Recipes {
		recipe, err := model.NewProductRecipe(
			tenantID,
			inventoryID,
			recipeReq.RawMaterialID,
			recipeReq.QuantityRequired,
		)
		if err != nil {
			return errors.NewValidationError(err.Error())
		}

		err = u.recipeRepo.Create(ctx, recipe)
		if err != nil {
			return errors.NewInternalError("gagal membuat resep: %v", err)
		}
	}

	return nil
}

func (u *tenantUsecase) UpdateProductRecipe(ctx context.Context, id string, req dto.UpdateProductRecipeRequest) (*dto.ProductRecipeResponse, error) {
	recipe, err := u.recipeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("product_recipe", "id", id)
	}

	err = recipe.UpdateQuantity(req.QuantityRequired)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	err = u.recipeRepo.Update(ctx, recipe)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengupdate resep: %v", err)
	}

	response := dto.ToProductRecipeResponse(recipe)
	return &response, nil
}

func (u *tenantUsecase) DeleteProductRecipe(ctx context.Context, id string) error {
	err := u.recipeRepo.Delete(ctx, id)
	if err != nil {
		return errors.NewInternalError("gagal menghapus resep: %v", err)
	}
	return nil
}

func (u *tenantUsecase) DeleteProductRecipes(ctx context.Context, inventoryID string) error {
	err := u.recipeRepo.DeleteByProductID(ctx, inventoryID)
	if err != nil {
		return errors.NewInternalError("gagal menghapus resep: %v", err)
	}
	return nil
}

// Product Availability

func (u *tenantUsecase) GetProductAvailability(ctx context.Context, inventoryID string) (*dto.ProductAvailabilityResponse, error) {
	availability, err := u.recipeRepo.GetProductAvailability(ctx, inventoryID)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengecek ketersediaan produk: %v", err)
	}

	response := dto.ToProductAvailabilityResponse(availability)
	return &response, nil
}

func (u *tenantUsecase) GetBatchProductAvailability(ctx context.Context, inventoryIDs []string) ([]*dto.ProductAvailabilityResponse, error) {
	availabilities, err := u.recipeRepo.GetBatchProductAvailability(ctx, inventoryIDs)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengecek ketersediaan produk: %v", err)
	}

	responses := make([]*dto.ProductAvailabilityResponse, len(availabilities))
	for i, availability := range availabilities {
		resp := dto.ToProductAvailabilityResponse(availability)
		responses[i] = &resp
	}

	return responses, nil
}

func (u *tenantUsecase) CheckCanProduce(ctx context.Context, inventoryID string, quantity int) (bool, []*dto.MaterialAvailabilityStatusResponse, error) {
	materials, err := u.recipeRepo.GetMaterialsNeededForProduction(ctx, inventoryID, quantity)
	if err != nil {
		return false, nil, errors.NewInternalError("gagal mengecek bahan yang dibutuhkan: %v", err)
	}

	canProduce := true
	responses := make([]*dto.MaterialAvailabilityStatusResponse, len(materials))
	for i, m := range materials {
		if !m.IsSufficient {
			canProduce = false
		}
		responses[i] = &dto.MaterialAvailabilityStatusResponse{
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

	return canProduce, responses, nil
}
