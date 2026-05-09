package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/example/jwt-ddd-clean/internal/application/dto"
	"github.com/example/jwt-ddd-clean/internal/application/usecase"
	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
	"github.com/example/jwt-ddd-clean/internal/pkg/errors"
	"github.com/gorilla/mux"
	middlewarehttp "github.com/example/jwt-ddd-clean/internal/http/middleware"
)

// TenantHandler handles tenant-related HTTP requests
type TenantHandler struct {
	tenantUsecase usecase.TenantUsecase
}

// NewTenantHandler creates a new TenantHandler
func NewTenantHandler(tenantUsecase usecase.TenantUsecase) *TenantHandler {
	return &TenantHandler{
		tenantUsecase: tenantUsecase,
	}
}

// Public endpoints (no authentication required)

// RegisterTenant handles tenant registration
func (h *TenantHandler) RegisterTenant(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	// Validate required fields
	if req.CompanyName == "" {
		h.sendError(w, errors.NewValidationError("nama perusahaan harus diisi"))
		return
	}
	if req.CompanySlug == "" {
		h.sendError(w, errors.NewValidationError("slug perusahaan harus diisi"))
		return
	}
	if req.Email == "" {
		h.sendError(w, errors.NewValidationError("email harus diisi"))
		return
	}
	if req.AdminUser.Username == "" || req.AdminUser.Email == "" || req.AdminUser.Password == "" || req.AdminUser.FullName == "" {
		h.sendError(w, errors.NewValidationError("data user admin tidak lengkap"))
		return
	}

	response, err := h.tenantUsecase.RegisterTenant(r.Context(), req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusCreated, true, "Registrasi perusahaan berhasil", response)
}

// GetSubscriptionPlans handles subscription plans listing
func (h *TenantHandler) GetSubscriptionPlans(w http.ResponseWriter, r *http.Request) {
	response, err := h.tenantUsecase.GetSubscriptionPlans(r.Context())
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil daftar paket langganan", response)
}

// Protected endpoints (require authentication)

// GetTenant handles getting current user's tenant
func (h *TenantHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context (set by tenant middleware)
	tenantID, ok := r.Context().Value(middlewarehttp.TenantIDKey).(string)
	if !ok || tenantID == "" {
		// Fallback: get tenant from user
		userID, ok := r.Context().Value(middlewarehttp.UserIDKey).(string)
		if !ok || userID == "" {
			h.sendError(w, errors.NewUnauthenticatedError("tenant not found in context"))
			return
		}
		// This would need a method to get tenant by user ID
		h.sendError(w, errors.NewValidationError("tenant information not available"))
		return
	}

	response, err := h.tenantUsecase.GetTenantByID(r.Context(), tenantID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil data perusahaan", response)
}

// GetTenantBySlug handles getting tenant by slug
func (h *TenantHandler) GetTenantBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	response, err := h.tenantUsecase.GetTenantBySlug(r.Context(), slug)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil data perusahaan", response)
}

// UpdateTenant handles tenant update
func (h *TenantHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middlewarehttp.TenantIDKey).(string)
	if !ok || tenantID == "" {
		h.sendError(w, errors.NewUnauthenticatedError("tenant not found in context"))
		return
	}

	var req dto.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	response, err := h.tenantUsecase.UpdateTenant(r.Context(), tenantID, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Data perusahaan berhasil diupdate", response)
}

// UpdateTenantSettings handles tenant settings update
func (h *TenantHandler) UpdateTenantSettings(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middlewarehttp.TenantIDKey).(string)
	if !ok || tenantID == "" {
		h.sendError(w, errors.NewUnauthenticatedError("tenant not found in context"))
		return
	}

	var req dto.UpdateTenantSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	err := h.tenantUsecase.UpdateTenantSettings(r.Context(), tenantID, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Pengaturan berhasil diupdate", nil)
}

// GetSubscriptionStatus handles getting subscription status
func (h *TenantHandler) GetSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middlewarehttp.TenantIDKey).(string)
	if !ok || tenantID == "" {
		h.sendError(w, errors.NewUnauthenticatedError("tenant not found in context"))
		return
	}

	response, err := h.tenantUsecase.GetSubscriptionStatus(r.Context(), tenantID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil status langganan", response)
}

// UpgradeSubscription handles subscription upgrade/downgrade
func (h *TenantHandler) UpgradeSubscription(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middlewarehttp.TenantIDKey).(string)
	if !ok || tenantID == "" {
		h.sendError(w, errors.NewUnauthenticatedError("tenant not found in context"))
		return
	}

	var req dto.UpgradeSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	if req.PlanID == "" {
		h.sendError(w, errors.NewValidationError("paket langganan harus dipilih"))
		return
	}

	response, err := h.tenantUsecase.UpgradeSubscription(r.Context(), tenantID, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Langganan berhasil diupgrade", response)
}

// CancelSubscription handles subscription cancellation
func (h *TenantHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middlewarehttp.TenantIDKey).(string)
	if !ok || tenantID == "" {
		h.sendError(w, errors.NewUnauthenticatedError("tenant not found in context"))
		return
	}

	err := h.tenantUsecase.CancelSubscription(r.Context(), tenantID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Langganan berhasil dibatalkan", nil)
}

// Raw Material endpoints

// CreateRawMaterial handles raw material creation
func (h *TenantHandler) CreateRawMaterial(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middlewarehttp.TenantIDKey).(string)
	if !ok || tenantID == "" {
		h.sendError(w, errors.NewUnauthenticatedError("tenant not found in context"))
		return
	}

	var req dto.CreateRawMaterialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	if req.SKU == "" || req.Name == "" || req.Unit == "" {
		h.sendError(w, errors.NewValidationError("SKU, nama, dan satuan harus diisi"))
		return
	}

	response, err := h.tenantUsecase.CreateRawMaterial(r.Context(), tenantID, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusCreated, true, "Bahan baku berhasil dibuat", response)
}

// GetRawMaterial handles getting a raw material
func (h *TenantHandler) GetRawMaterial(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	response, err := h.tenantUsecase.GetRawMaterial(r.Context(), id)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil data bahan baku", response)
}

// ListRawMaterials handles listing raw materials
func (h *TenantHandler) ListRawMaterials(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middlewarehttp.TenantIDKey).(string)
	if !ok || tenantID == "" {
		h.sendError(w, errors.NewUnauthenticatedError("tenant not found in context"))
		return
	}

	filter := repository.RawMaterialFilter{
		TenantID: tenantID,
		Limit:    20,
		Offset:   0,
	}

	// Parse query parameters
	if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = search
	}
	if supplier := r.URL.Query().Get("supplier"); supplier != "" {
		filter.Supplier = supplier
	}
	if r.URL.Query().Get("low_stock") == "true" {
		filter.LowStock = true
	}
	if r.URL.Query().Get("out_of_stock") == "true" {
		filter.OutOfStock = true
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if limit, err := strconv.Atoi(l); err == nil {
			filter.Limit = limit
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if offset, err := strconv.Atoi(o); err == nil {
			filter.Offset = offset
		}
	}

	response, err := h.tenantUsecase.ListRawMaterials(r.Context(), filter)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil daftar bahan baku", response)
}

// UpdateRawMaterial handles raw material update
func (h *TenantHandler) UpdateRawMaterial(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req dto.UpdateRawMaterialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	response, err := h.tenantUsecase.UpdateRawMaterial(r.Context(), id, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Bahan baku berhasil diupdate", response)
}

// DeleteRawMaterial handles raw material deletion
func (h *TenantHandler) DeleteRawMaterial(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.tenantUsecase.DeleteRawMaterial(r.Context(), id)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Bahan baku berhasil dihapus", nil)
}

// AdjustRawMaterialStock handles raw material stock adjustment
func (h *TenantHandler) AdjustRawMaterialStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req dto.AdjustRawMaterialStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	err := h.tenantUsecase.AdjustRawMaterialStock(r.Context(), id, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Stok bahan baku berhasil diupdate", nil)
}

// GetLowStockMaterials handles listing low stock materials
func (h *TenantHandler) GetLowStockMaterials(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middlewarehttp.TenantIDKey).(string)
	if !ok || tenantID == "" {
		h.sendError(w, errors.NewUnauthenticatedError("tenant not found in context"))
		return
	}

	response, err := h.tenantUsecase.GetLowStockMaterials(r.Context(), tenantID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil daftar bahan baku stok rendah", response)
}

// GetOutOfStockMaterials handles listing out of stock materials
func (h *TenantHandler) GetOutOfStockMaterials(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middlewarehttp.TenantIDKey).(string)
	if !ok || tenantID == "" {
		h.sendError(w, errors.NewUnauthenticatedError("tenant not found in context"))
		return
	}

	response, err := h.tenantUsecase.GetOutOfStockMaterials(r.Context(), tenantID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil daftar bahan baku habis", response)
}

// Product Recipe endpoints

// CreateProductRecipe handles product recipe creation
func (h *TenantHandler) CreateProductRecipe(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middlewarehttp.TenantIDKey).(string)
	if !ok || tenantID == "" {
		h.sendError(w, errors.NewUnauthenticatedError("tenant not found in context"))
		return
	}

	var req dto.CreateProductRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	if req.InventoryID == "" || req.RawMaterialID == "" || req.QuantityRequired <= 0 {
		h.sendError(w, errors.NewValidationError("data resep tidak lengkap"))
		return
	}

	response, err := h.tenantUsecase.CreateProductRecipe(r.Context(), tenantID, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusCreated, true, "Resep produk berhasil dibuat", response)
}

// GetProductRecipes handles getting product recipes
func (h *TenantHandler) GetProductRecipes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	inventoryID := vars["inventoryId"]

	response, err := h.tenantUsecase.GetProductRecipes(r.Context(), inventoryID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil resep produk", response)
}

// SetProductRecipes handles setting all recipes for a product
func (h *TenantHandler) SetProductRecipes(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middlewarehttp.TenantIDKey).(string)
	if !ok || tenantID == "" {
		h.sendError(w, errors.NewUnauthenticatedError("tenant not found in context"))
		return
	}

	vars := mux.Vars(r)
	inventoryID := vars["inventoryId"]

	var req dto.SetProductRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	err := h.tenantUsecase.SetProductRecipes(r.Context(), tenantID, inventoryID, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Resep produk berhasil diset", nil)
}

// UpdateProductRecipe handles product recipe update
func (h *TenantHandler) UpdateProductRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req dto.UpdateProductRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	if req.QuantityRequired <= 0 {
		h.sendError(w, errors.NewValidationError("jumlah yang dibutuhkan harus lebih dari 0"))
		return
	}

	response, err := h.tenantUsecase.UpdateProductRecipe(r.Context(), id, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Resep produk berhasil diupdate", response)
}

// DeleteProductRecipe handles product recipe deletion
func (h *TenantHandler) DeleteProductRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.tenantUsecase.DeleteProductRecipe(r.Context(), id)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Resep produk berhasil dihapus", nil)
}

// DeleteProductRecipes handles deleting all recipes for a product
func (h *TenantHandler) DeleteProductRecipes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	inventoryID := vars["inventoryId"]

	err := h.tenantUsecase.DeleteProductRecipes(r.Context(), inventoryID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Semua resep produk berhasil dihapus", nil)
}

// Product Availability endpoints

// GetProductAvailability handles checking product availability
func (h *TenantHandler) GetProductAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	inventoryID := vars["inventoryId"]

	response, err := h.tenantUsecase.GetProductAvailability(r.Context(), inventoryID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengecek ketersediaan produk", response)
}

// CheckCanProduce handles checking if product can be produced
func (h *TenantHandler) CheckCanProduce(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	inventoryID := vars["inventoryId"]

	quantityStr := r.URL.Query().Get("quantity")
	quantity := 1
	if quantityStr != "" {
		if q, err := strconv.Atoi(quantityStr); err == nil && q > 0 {
			quantity = q
		}
	}

	canProduce, materials, err := h.tenantUsecase.CheckCanProduce(r.Context(), inventoryID, quantity)
	if err != nil {
		h.sendError(w, err)
		return
	}

	message := "Produk dapat dibuat"
	if !canProduce {
		message = "Produk tidak dapat dibuat - bahan baku tidak mencukupi"
	}

	h.sendJSON(w, http.StatusOK, true, message, map[string]interface{}{
		"can_produce": canProduce,
		"quantity":    quantity,
		"materials":   materials,
	})
}

// Admin endpoints (require SUPER_ADMIN role)

// ListTenants handles listing all tenants (admin only)
func (h *TenantHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	filter := repository.TenantFilter{
		Limit:  20,
		Offset: 0,
	}

	// Parse query parameters
	if status := r.URL.Query().Get("status"); status != "" {
		filter.SubscriptionStatus = model.SubscriptionStatus(status)
	}
	if planID := r.URL.Query().Get("plan"); planID != "" {
		filter.SubscriptionPlanID = planID
	}
	if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = search
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if limit, err := strconv.Atoi(l); err == nil {
			filter.Limit = limit
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if offset, err := strconv.Atoi(o); err == nil {
			filter.Offset = offset
		}
	}

	response, err := h.tenantUsecase.ListTenants(r.Context(), filter)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil daftar perusahaan", response)
}

// ActivateTenant handles tenant activation (admin only)
func (h *TenantHandler) ActivateTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.tenantUsecase.ActivateTenant(r.Context(), id)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Perusahaan berhasil diaktifkan", nil)
}

// DeactivateTenant handles tenant deactivation (admin only)
func (h *TenantHandler) DeactivateTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.tenantUsecase.DeactivateTenant(r.Context(), id)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Perusahaan berhasil dinonaktifkan", nil)
}

// GetTenantByID handles getting a tenant by ID (admin only)
func (h *TenantHandler) GetTenantByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	response, err := h.tenantUsecase.GetTenantByID(r.Context(), id)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil data perusahaan", response)
}

// UpdateTenantByID handles updating a tenant by ID (admin only)
func (h *TenantHandler) UpdateTenantByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req dto.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	response, err := h.tenantUsecase.UpdateTenant(r.Context(), id, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Data perusahaan berhasil diupdate", response)
}

// UpdateTenantSettingsByID handles updating tenant settings by ID (admin only)
func (h *TenantHandler) UpdateTenantSettingsByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req dto.UpdateTenantSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	err := h.tenantUsecase.UpdateTenantSettings(r.Context(), id, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Pengaturan berhasil diupdate", nil)
}

// Helper methods

func (h *TenantHandler) sendJSON(w http.ResponseWriter, status int, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"message": message,
		"data":    data,
	})
}

func (h *TenantHandler) sendError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.GetHTTPStatus())
		json.NewEncoder(w).Encode(appErr.ToResponse())
		return
	}

	// Unknown error
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(errors.ErrInternalErr.ToResponse())
}
