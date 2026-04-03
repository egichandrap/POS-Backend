package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/example/jwt-ddd-clean/internal/application/dto"
	"github.com/example/jwt-ddd-clean/internal/application/usecase"
	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
	"github.com/example/jwt-ddd-clean/internal/pkg/errors"
	"github.com/gorilla/mux"
)

// POSHandler handles Point of Sale HTTP requests
type POSHandler struct {
	posUsecase usecase.POSUsecase
}

// NewPOSHandler creates a new POSHandler
func NewPOSHandler(posUsecase usecase.POSUsecase) *POSHandler {
	return &POSHandler{
		posUsecase: posUsecase,
	}
}

// CreateCart creates a new cart
func (h *POSHandler) CreateCart(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req struct {
		CustomerName string `json:"customer_name,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	cart, err := h.posUsecase.CreateCart(r.Context(), userID, req.CustomerName)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusCreated, true, "Cart berhasil dibuat", cart)
}

// GetCart retrieves a cart
func (h *POSHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartID := vars["id"]

	cart, err := h.posUsecase.GetCart(r.Context(), cartID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil cart", cart)
}

// GetOrCreateCart gets existing cart or creates new one
func (h *POSHandler) GetOrCreateCart(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	cart, err := h.posUsecase.GetOrCreateCart(r.Context(), userID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil cart", cart)
}

// AddToCart adds an item to cart
func (h *POSHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartID := vars["id"]

	var req dto.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	if req.ProductID == "" || req.Quantity < 1 {
		h.sendError(w, errors.NewValidationError("product_id dan quantity harus diisi"))
		return
	}

	cart, err := h.posUsecase.AddToCart(r.Context(), cartID, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Item berhasil ditambahkan", cart)
}

// RemoveFromCart removes an item from cart
func (h *POSHandler) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartID := vars["id"]

	var req struct {
		ProductID string `json:"product_id" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	cart, err := h.posUsecase.RemoveFromCart(r.Context(), cartID, req.ProductID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Item berhasil dihapus", cart)
}

// UpdateCartItemQuantity updates cart item quantity
func (h *POSHandler) UpdateCartItemQuantity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartID := vars["id"]

	var req dto.UpdateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	cart, err := h.posUsecase.UpdateCartItemQuantity(r.Context(), cartID, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Quantity berhasil diupdate", cart)
}

// ClearCart clears all items from cart
func (h *POSHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartID := vars["id"]

	err := h.posUsecase.ClearCart(r.Context(), cartID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Cart berhasil dikosongkan", nil)
}

// DeleteCart deletes the cart
func (h *POSHandler) DeleteCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartID := vars["id"]

	err := h.posUsecase.DeleteCart(r.Context(), cartID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Cart berhasil dihapus", nil)
}

// Checkout processes a checkout
func (h *POSHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartID := vars["id"]

	var req dto.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	// Get cashier name from context
	cashierName, _ := r.Context().Value("username").(string)

	transaction, err := h.posUsecase.Checkout(r.Context(), cartID, req, cashierName)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusCreated, true, "Checkout berhasil", transaction)
}

// GetTransaction retrieves a transaction
func (h *POSHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	transactionID := vars["id"]

	transaction, err := h.posUsecase.GetTransaction(r.Context(), transactionID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil transaksi", transaction)
}

// ListTransactions lists transactions with pagination
func (h *POSHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	limit := 20
	offset := 0
	status := r.URL.Query().Get("status")
	paymentMethod := r.URL.Query().Get("payment_method")
	search := r.URL.Query().Get("search")

	if l := r.URL.Query().Get("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			limit = 20
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if _, err := fmt.Sscanf(o, "%d", &offset); err != nil {
			offset = 0
		}
	}

	filter := repository.TransactionFilter{
		Status:        model.TransactionStatus(status),
		PaymentMethod: model.PaymentMethod(paymentMethod),
		Search:        search,
		Limit:         limit,
		Offset:        offset,
	}

	response, err := h.posUsecase.ListTransactions(r.Context(), filter)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil daftar transaksi", response)
}

// GetTodaySales retrieves today's sales summary
func (h *POSHandler) GetTodaySales(w http.ResponseWriter, r *http.Request) {
	sales, err := h.posUsecase.GetTodaySales(r.Context())
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil sales summary", sales)
}

// CancelTransaction cancels a transaction
func (h *POSHandler) CancelTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	transactionID := vars["id"]

	transaction, err := h.posUsecase.CancelTransaction(r.Context(), transactionID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Transaksi berhasil dibatalkan", transaction)
}

// Helper methods

func (h *POSHandler) sendJSON(w http.ResponseWriter, status int, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"message": message,
		"data":    data,
	})
}

func (h *POSHandler) sendError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.GetHTTPStatus())
		json.NewEncoder(w).Encode(appErr.ToResponse())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(errors.ErrInternalErr.ToResponse())
}
