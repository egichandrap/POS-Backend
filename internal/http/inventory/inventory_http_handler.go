package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/example/jwt-ddd-clean/internal/application/dto"
	"github.com/example/jwt-ddd-clean/internal/application/usecase"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
	apperrors "github.com/example/jwt-ddd-clean/internal/pkg/errors"
)

// StockUpdateResponse represents a stock update response with previous quantity
type StockUpdateResponse struct {
	ID          string  `json:"id"`
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Quantity    int     `json:"quantity"`
	PreviousQty int     `json:"previous_quantity"`
	UpdatedAt   string  `json:"updated_at"`
}

// InventoryHTTPHandler handles HTTP requests for inventory operations
type InventoryHTTPHandler struct {
	inventoryUsecase usecase.InventoryUsecase
}

// NewInventoryHTTPHandler creates a new InventoryHTTPHandler
func NewInventoryHTTPHandler(inventoryUsecase usecase.InventoryUsecase) *InventoryHTTPHandler {
	return &InventoryHTTPHandler{
		inventoryUsecase: inventoryUsecase,
	}
}

// CreateInventory handles POST /api/inventory
func (h *InventoryHTTPHandler) CreateInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CreateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, apperrors.ErrValidationErr.WithDetails("Invalid request body"))
		return
	}

	result, err := h.inventoryUsecase.CreateInventory(r.Context(), req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendSuccess(w, "Inventory item created successfully", result, http.StatusCreated)
}

// GetInventory handles GET /api/inventory/{id}
func (h *InventoryHTTPHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	id := extractIDFromPath(r.URL.Path)
	if id == "" {
		h.sendError(w, apperrors.NewValidationError("id", "is required"))
		return
	}

	result, err := h.inventoryUsecase.GetInventory(r.Context(), id)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendSuccess(w, "Inventory item retrieved successfully", result, http.StatusOK)
}

// UpdateInventory handles PUT /api/inventory/{id}
func (h *InventoryHTTPHandler) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	id := extractIDFromPath(r.URL.Path)

	var req dto.UpdateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, apperrors.ErrValidationErr.WithDetails("Invalid request body"))
		return
	}

	result, err := h.inventoryUsecase.UpdateInventory(r.Context(), id, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendSuccess(w, "Inventory item updated successfully", result, http.StatusOK)
}

// DeleteInventory handles DELETE /api/inventory/{id}
func (h *InventoryHTTPHandler) DeleteInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	id := extractIDFromPath(r.URL.Path)
	if id == "" {
		h.sendError(w, apperrors.NewValidationError("id", "is required"))
		return
	}

	if err := h.inventoryUsecase.DeleteInventory(r.Context(), id); err != nil {
		h.sendError(w, err)
		return
	}

	h.sendSuccess(w, "Inventory item deleted successfully", nil, http.StatusOK)
}

// ListInventory handles GET /api/inventory
func (h *InventoryHTTPHandler) ListInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	filter := repository.InventoryFilter{
		Limit:  20,
		Offset: 0,
	}

	if sku := r.URL.Query().Get("sku"); sku != "" {
		filter.SKU = &sku
	}
	if name := r.URL.Query().Get("name"); name != "" {
		filter.Name = &name
	}
	if location := r.URL.Query().Get("location"); location != "" {
		filter.Location = &location
	}
	if minQty := r.URL.Query().Get("min_qty"); minQty != "" {
		if val, err := strconv.Atoi(minQty); err == nil {
			filter.MinQty = &val
		}
	}
	if maxQty := r.URL.Query().Get("max_qty"); maxQty != "" {
		if val, err := strconv.Atoi(maxQty); err == nil {
			filter.MaxQty = &val
		}
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			filter.Limit = val
		}
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		if val, err := strconv.Atoi(offset); err == nil {
			filter.Offset = val
		}
	}

	result, err := h.inventoryUsecase.ListInventory(r.Context(), filter)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendSuccess(w, "Inventory items retrieved successfully", result, http.StatusOK)
}

// UpdateStock handles PUT /api/inventory/{id}/stock
func (h *InventoryHTTPHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	id := extractIDFromPath(r.URL.Path)
	if id == "" {
		h.sendError(w, apperrors.NewValidationError("id", "is required"))
		return
	}

	var req dto.UpdateStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, apperrors.ErrValidationErr.WithDetails("Invalid request body"))
		return
	}

	// Get current inventory for previous quantity
	current, _ := h.inventoryUsecase.GetInventory(r.Context(), id)
	previousQty := 0
	if current != nil {
		previousQty = current.Quantity
	}

	result, err := h.inventoryUsecase.UpdateStock(r.Context(), id, req.Quantity)
	if err != nil {
		h.sendError(w, err)
		return
	}

	response := StockUpdateResponse{
		ID:          result.ID,
		SKU:         result.SKU,
		Name:        result.Name,
		Quantity:    result.Quantity,
		PreviousQty: previousQty,
		UpdatedAt:   result.UpdatedAt,
	}

	h.sendSuccess(w, "Stock quantity updated successfully", response, http.StatusOK)
}

// AdjustStock handles POST /api/inventory/{id}/stock/adjust
func (h *InventoryHTTPHandler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	id := extractIDFromPath(r.URL.Path)
	if id == "" {
		h.sendError(w, apperrors.NewValidationError("id", "is required"))
		return
	}

	var req dto.AdjustStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, apperrors.ErrValidationErr.WithDetails("Invalid request body"))
		return
	}

	// Get current inventory for previous quantity
	current, _ := h.inventoryUsecase.GetInventory(r.Context(), id)
	previousQty := 0
	if current != nil {
		previousQty = current.Quantity
	}

	result, err := h.inventoryUsecase.AdjustStock(r.Context(), id, req.Quantity)
	if err != nil {
		h.sendError(w, err)
		return
	}

	response := StockUpdateResponse{
		ID:          result.ID,
		SKU:         result.SKU,
		Name:        result.Name,
		Quantity:    result.Quantity,
		PreviousQty: previousQty,
		UpdatedAt:   result.UpdatedAt,
	}

	h.sendSuccess(w, "Stock quantity adjusted successfully", response, http.StatusOK)
}

func (h *InventoryHTTPHandler) sendSuccess(w http.ResponseWriter, message string, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := apperrors.SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *InventoryHTTPHandler) sendError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		w.WriteHeader(appErr.GetHTTPStatus())
		json.NewEncoder(w).Encode(appErr.ToResponse())
		return
	}

	// Fallback for non-AppError
	w.WriteHeader(http.StatusInternalServerError)
	response := apperrors.ErrorResponse{
		Success: false,
		Error: apperrors.ErrorDetail{
			Code:    string(apperrors.ErrInternal),
			Message: "An unexpected error occurred",
			Details: err.Error(),
		},
	}
	json.NewEncoder(w).Encode(response)
}

// extractIDFromPath extracts the ID from a URL path like /api/inventory/{id}
func extractIDFromPath(path string) string {
	// Remove trailing slash
	path = strings.TrimSuffix(path, "/")
	// Split by /
	parts := strings.Split(path, "/")
	// Return the last part
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
