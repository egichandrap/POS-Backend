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

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authUsecase usecase.AuthUsecase
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
	}
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		h.sendError(w, errors.NewValidationError("username dan password harus diisi"))
		return
	}

	response, err := h.authUsecase.Login(r.Context(), req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Login berhasil", response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	token := h.extractToken(r)
	if token == "" {
		h.sendError(w, errors.NewUnauthenticatedError("token tidak ditemukan"))
		return
	}

	err := h.authUsecase.Logout(r.Context(), token)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Logout berhasil", nil)
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" || req.FullName == "" {
		h.sendError(w, errors.NewValidationError("semua field harus diisi"))
		return
	}

	if req.Role == "" {
		req.Role = model.RoleCashier // Default role
	}

	response, err := h.authUsecase.Register(r.Context(), req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusCreated, true, "Registrasi berhasil", response)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	if req.RefreshToken == "" {
		h.sendError(w, errors.NewValidationError("refresh token harus diisi"))
		return
	}

	response, err := h.authUsecase.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Token berhasil di-refresh", response)
}

// GetMe returns current user information
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("user_id").(string)

	response, err := h.authUsecase.GetMe(r.Context(), userID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil data user", response)
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		h.sendError(w, errors.NewValidationError("password lama dan baru harus diisi"))
		return
	}

	err := h.authUsecase.ChangePassword(r.Context(), userID, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Password berhasil diubah", nil)
}

// CreateUser creates a new user (admin only)
func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" || req.FullName == "" {
		h.sendError(w, errors.NewValidationError("semua field harus diisi"))
		return
	}

	if req.Role == "" {
		req.Role = model.RoleCashier // Default role
	}

	response, err := h.authUsecase.Register(r.Context(), req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusCreated, true, "User berhasil dibuat", response)
}

// ListUsers returns paginated list of users (admin only)
func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Get query params
	limit := 20
	offset := 0
	role := r.URL.Query().Get("role")
	status := r.URL.Query().Get("status")
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

	filter := repository.UserFilter{
		Role:   model.UserRole(role),
		Status: model.UserStatus(status),
		Search: search,
		Limit:  limit,
		Offset: offset,
	}

	response, err := h.authUsecase.ListUsers(r.Context(), filter)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil daftar user", response)
}

// GetUserByID returns a specific user
func (h *AuthHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	// We'll use GetMe logic but for specific user
	// This requires admin permission
	response, err := h.authUsecase.GetMe(r.Context(), userID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "Berhasil mengambil data user", response)
}

// UpdateUser updates a user
func (h *AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.NewValidationError("request body tidak valid"))
		return
	}

	response, err := h.authUsecase.UpdateUser(r.Context(), userID, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "User berhasil diupdate", response)
}

// DeleteUser deletes a user
func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	err := h.authUsecase.DeleteUser(r.Context(), userID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, true, "User berhasil dihapus", nil)
}

// Helper methods

func (h *AuthHandler) sendJSON(w http.ResponseWriter, status int, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"message": message,
		"data":    data,
	})
}

func (h *AuthHandler) sendError(w http.ResponseWriter, err error) {
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

func (h *AuthHandler) extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) >= 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}
