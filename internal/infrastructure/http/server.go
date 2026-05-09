package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/example/jwt-ddd-clean/internal/application/usecase"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
	"github.com/example/jwt-ddd-clean/internal/domain/service"
	"github.com/example/jwt-ddd-clean/internal/handler"
	inventoryhttp "github.com/example/jwt-ddd-clean/internal/http/inventory"
	httpmiddleware "github.com/example/jwt-ddd-clean/internal/http/middleware"
	"github.com/example/jwt-ddd-clean/internal/infrastructure/jwt"
	infrarepo "github.com/example/jwt-ddd-clean/internal/infrastructure/repository"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	router     *mux.Router
	config     ServerConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host            string
	Port            string
	SecretKey       string
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	DatabasePath    string
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Host:            "localhost",
		Port:            "8080",
		SecretKey:       "your-super-secret-key-change-in-production",
		Issuer:          "jwt-ddd-clean",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		DatabasePath:    "./data/inventory.db",
	}
}

// setupRoutes configures all API routes using gorilla/mux
func setupRoutes(
	r *mux.Router,
	tokenHTTPHandler *TokenHTTPHandler,
	inventoryHTTPHandler *inventoryhttp.InventoryHTTPHandler,
	authHandler *handler.AuthHandler,
	posHandler *handler.POSHandler,
	tableHandler *handler.TableHandler,
	guestOrderHandler *handler.GuestOrderHandler,
	tenantHandler *handler.TenantHandler,
	healthHandler *handler.HealthHandler,
	authMiddleware *httpmiddleware.AuthMiddleware,
	tenantMiddleware *httpmiddleware.TenantMiddleware,
) {
	// Apply security middleware globally
	r.Use(httpmiddleware.CORSMiddleware(httpmiddleware.CORSConfig{
		AllowedOrigins:   []string{"*"}, // Allow all origins for development
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Request-ID", "X-CSRF-Token"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false, // Set to false when using "*" origin
		MaxAge:           3600,
	}))
	r.Use(httpmiddleware.SecurityHeadersMiddleware)
	r.Use(httpmiddleware.ValidationMiddleware)
	r.Use(httpmiddleware.MaxBodySizeMiddleware(httpmiddleware.DefaultMaxBodySize))

	// Handle OPTIONS for ALL /api/* routes (CORS preflight) - must be before specific routes
	r.HandleFunc("/api/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}).Methods(http.MethodOptions)

	// Public routes (no authentication required)
	publicRouter := r.PathPrefix("/api").Subrouter()
	publicRouter.Use(httpmiddleware.LoginMaxBodyMiddleware())

	publicRouter.HandleFunc("/auth/login", authHandler.Login).Methods(http.MethodPost)
	publicRouter.HandleFunc("/auth/register", authHandler.Register).Methods(http.MethodPost)
	publicRouter.HandleFunc("/auth/refresh", authHandler.RefreshToken).Methods(http.MethodPost)
	publicRouter.HandleFunc("/token/generate", tokenHTTPHandler.GenerateToken).Methods(http.MethodPost)
	publicRouter.HandleFunc("/token/refresh", tokenHTTPHandler.RefreshToken).Methods(http.MethodPost)
	publicRouter.HandleFunc("/token/validate", tokenHTTPHandler.ValidateToken).Methods(http.MethodPost)
	publicRouter.HandleFunc("/token/revoke", tokenHTTPHandler.RevokeToken).Methods(http.MethodPost)

	// Health check routes (no security headers needed)
	r.HandleFunc("/api/health", healthHandler.HealthCheck).Methods(http.MethodGet)
	r.HandleFunc("/api/ready", healthHandler.Ready).Methods(http.MethodGet)
	r.HandleFunc("/api/live", healthHandler.Live).Methods(http.MethodGet)

	// ============================================
	// TENANT PUBLIC ROUTES (no authentication required)
	// ============================================
	// Reuse the existing publicRouter from above

	// Subscription plans (public)
	publicRouter.HandleFunc("/subscriptions/plans", tenantHandler.GetSubscriptionPlans).Methods(http.MethodGet)

	// Tenant registration (public)
	publicRouter.HandleFunc("/tenants/register", tenantHandler.RegisterTenant).Methods(http.MethodPost)

	// Get tenant by slug (public)
	publicRouter.HandleFunc("/tenants/slug/{slug}", tenantHandler.GetTenantBySlug).Methods(http.MethodGet)

	// ============================================
	// PROTECTED ROUTES (authentication required)
	// ============================================
	protectedRouter := r.PathPrefix("/api").Subrouter()
	protectedRouter.Use(authMiddleware.Authenticate)
	protectedRouter.Use(tenantMiddleware.TenantFromUser) // Auto-extract tenant from authenticated user

	// Tenant management routes
	protectedRouter.HandleFunc("/tenant", tenantHandler.GetTenant).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/tenant", tenantHandler.UpdateTenant).Methods(http.MethodPut)
	protectedRouter.HandleFunc("/tenant/settings", tenantHandler.UpdateTenantSettings).Methods(http.MethodPut)
	protectedRouter.HandleFunc("/tenant/subscription", tenantHandler.GetSubscriptionStatus).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/tenant/subscription/upgrade", tenantHandler.UpgradeSubscription).Methods(http.MethodPost)
	protectedRouter.HandleFunc("/tenant/subscription/cancel", tenantHandler.CancelSubscription).Methods(http.MethodPost)

	// Raw Material routes
	protectedRouter.HandleFunc("/raw-materials", tenantHandler.ListRawMaterials).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/raw-materials", tenantHandler.CreateRawMaterial).Methods(http.MethodPost)
	protectedRouter.HandleFunc("/raw-materials/low-stock", tenantHandler.GetLowStockMaterials).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/raw-materials/out-of-stock", tenantHandler.GetOutOfStockMaterials).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/raw-materials/{id}", tenantHandler.GetRawMaterial).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/raw-materials/{id}", tenantHandler.UpdateRawMaterial).Methods(http.MethodPut)
	protectedRouter.HandleFunc("/raw-materials/{id}", tenantHandler.DeleteRawMaterial).Methods(http.MethodDelete)
	protectedRouter.HandleFunc("/raw-materials/{id}/stock", tenantHandler.AdjustRawMaterialStock).Methods(http.MethodPost)

	// Product Recipe routes (using /inventory path)
	protectedRouter.HandleFunc("/inventory/{inventoryId}/recipes", tenantHandler.GetProductRecipes).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/inventory/{inventoryId}/recipes", tenantHandler.CreateProductRecipe).Methods(http.MethodPost)
	protectedRouter.HandleFunc("/inventory/{inventoryId}/recipes/set", tenantHandler.SetProductRecipes).Methods(http.MethodPost)
	protectedRouter.HandleFunc("/inventory/{inventoryId}/recipes/delete", tenantHandler.DeleteProductRecipes).Methods(http.MethodDelete)
	protectedRouter.HandleFunc("/inventory/{inventoryId}/availability", tenantHandler.GetProductAvailability).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/inventory/{inventoryId}/can-produce", tenantHandler.CheckCanProduce).Methods(http.MethodGet)

	// Product Recipe routes (using /products path for backward compatibility)
	protectedRouter.HandleFunc("/products/{inventoryId}/recipes", tenantHandler.GetProductRecipes).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/products/{inventoryId}/recipes", tenantHandler.SetProductRecipes).Methods(http.MethodPost)
	protectedRouter.HandleFunc("/products/{inventoryId}/recipes/delete", tenantHandler.DeleteProductRecipes).Methods(http.MethodDelete)
	protectedRouter.HandleFunc("/products/{inventoryId}/availability", tenantHandler.GetProductAvailability).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/products/{inventoryId}/can-produce", tenantHandler.CheckCanProduce).Methods(http.MethodGet)

	// Individual recipe management
	protectedRouter.HandleFunc("/recipes/{id}", tenantHandler.UpdateProductRecipe).Methods(http.MethodPut)
	protectedRouter.HandleFunc("/recipes/{id}", tenantHandler.DeleteProductRecipe).Methods(http.MethodDelete)

	// Admin routes (require admin role)
	adminRouter := protectedRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(authMiddleware.RequireRole("SUPER_ADMIN", "ADMIN"))

	adminRouter.HandleFunc("/users", authHandler.ListUsers).Methods(http.MethodGet)
	adminRouter.HandleFunc("/users", authHandler.CreateUser).Methods(http.MethodPost)
	adminRouter.HandleFunc("/users/{id}", authHandler.GetUserByID).Methods(http.MethodGet)
	adminRouter.HandleFunc("/users/{id}", authHandler.UpdateUser).Methods(http.MethodPut)
	adminRouter.HandleFunc("/users/{id}", authHandler.DeleteUser).Methods(http.MethodDelete)

	// Tenant admin routes (SUPER_ADMIN only)
	superAdminRouter := protectedRouter.PathPrefix("/admin/tenants").Subrouter()
	superAdminRouter.Use(authMiddleware.RequireRole("SUPER_ADMIN"))

	superAdminRouter.HandleFunc("", tenantHandler.ListTenants).Methods(http.MethodGet)
	superAdminRouter.HandleFunc("/{id}", tenantHandler.GetTenantByID).Methods(http.MethodGet)
	superAdminRouter.HandleFunc("/{id}", tenantHandler.UpdateTenantByID).Methods(http.MethodPut)
	superAdminRouter.HandleFunc("/{id}/settings", tenantHandler.UpdateTenantSettingsByID).Methods(http.MethodPut)
	superAdminRouter.HandleFunc("/{id}/activate", tenantHandler.ActivateTenant).Methods(http.MethodPost)
	superAdminRouter.HandleFunc("/{id}/deactivate", tenantHandler.DeactivateTenant).Methods(http.MethodPost)

	// Inventory routes (require authentication, admin for write operations)
	inventoryRouter := protectedRouter.PathPrefix("/inventory").Subrouter()

	// Read operations - any authenticated user can read
	inventoryRouter.HandleFunc("", inventoryHTTPHandler.ListInventory).Methods(http.MethodGet)
	inventoryRouter.HandleFunc("/{id}", inventoryHTTPHandler.GetInventory).Methods(http.MethodGet)

	// Write operations - only admin/superadmin (handled in handler with context check)
	inventoryRouter.HandleFunc("", inventoryHTTPHandler.CreateInventory).Methods(http.MethodPost)
	inventoryRouter.HandleFunc("/{id}", inventoryHTTPHandler.UpdateInventory).Methods(http.MethodPut)
	inventoryRouter.HandleFunc("/{id}", inventoryHTTPHandler.DeleteInventory).Methods(http.MethodDelete)
	inventoryRouter.HandleFunc("/{id}/stock", inventoryHTTPHandler.UpdateStock).Methods(http.MethodPut)
	inventoryRouter.HandleFunc("/{id}/stock/adjust", inventoryHTTPHandler.AdjustStock).Methods(http.MethodPost)

	// POS routes (require authentication)
	posRouter := protectedRouter.PathPrefix("/pos").Subrouter()

	// Cart routes
	posRouter.HandleFunc("/cart", posHandler.CreateCart).Methods(http.MethodPost)
	posRouter.HandleFunc("/cart/my", posHandler.GetOrCreateCart).Methods(http.MethodGet)
	posRouter.HandleFunc("/cart/{id}", posHandler.GetCart).Methods(http.MethodGet)
	posRouter.HandleFunc("/cart/{id}/items", posHandler.AddToCart).Methods(http.MethodPost)
	posRouter.HandleFunc("/cart/{id}/items", posHandler.UpdateCartItemQuantity).Methods(http.MethodPut)
	posRouter.HandleFunc("/cart/{id}/items", posHandler.RemoveFromCart).Methods(http.MethodDelete)
	posRouter.HandleFunc("/cart/{id}/clear", posHandler.ClearCart).Methods(http.MethodPost)
	posRouter.HandleFunc("/cart/{id}", posHandler.DeleteCart).Methods(http.MethodDelete)

	// Checkout & Transaction routes
	posRouter.HandleFunc("/checkout/{id}", posHandler.Checkout).Methods(http.MethodPost)
	posRouter.HandleFunc("/transactions", posHandler.ListTransactions).Methods(http.MethodGet)
	posRouter.HandleFunc("/transactions/{id}", posHandler.GetTransaction).Methods(http.MethodGet)
	posRouter.HandleFunc("/transactions/{id}/cancel", posHandler.CancelTransaction).Methods(http.MethodPost)
	posRouter.HandleFunc("/transactions/{id}/refund", posHandler.RefundTransaction).Methods(http.MethodPost)

	// Sales summary
	posRouter.HandleFunc("/sales/today", posHandler.GetTodaySales).Methods(http.MethodGet)

	// Table Management routes (require authentication, admin only)
	tableRouter := protectedRouter.PathPrefix("/tables").Subrouter()
	tableRouter.Use(authMiddleware.RequireRole("SUPER_ADMIN", "ADMIN"))

	tableRouter.HandleFunc("", tableHandler.ListTables).Methods(http.MethodGet)
	tableRouter.HandleFunc("", tableHandler.CreateTable).Methods(http.MethodPost)
	tableRouter.HandleFunc("/available", tableHandler.GetAvailableTables).Methods(http.MethodGet)
	tableRouter.HandleFunc("/{id}", tableHandler.GetTable).Methods(http.MethodGet)
	tableRouter.HandleFunc("/{id}", tableHandler.UpdateTable).Methods(http.MethodPut)
	tableRouter.HandleFunc("/{id}", tableHandler.DeleteTable).Methods(http.MethodDelete)
	tableRouter.HandleFunc("/{id}/status", tableHandler.UpdateTableStatus).Methods(http.MethodPost)
	tableRouter.HandleFunc("/{id}/qr", tableHandler.GenerateQRCode).Methods(http.MethodPost)

	// Guest Order routes
	// Public routes (no auth required for customer ordering)
	guestOrderRouter := r.PathPrefix("/api").Subrouter()
	guestOrderRouter.HandleFunc("/guest/orders", guestOrderHandler.CreateOrder).Methods(http.MethodPost)
	guestOrderRouter.HandleFunc("/guest/orders/{id}", guestOrderHandler.GetOrder).Methods(http.MethodGet)
	guestOrderRouter.HandleFunc("/guest/orders/{id}/items", guestOrderHandler.AddItem).Methods(http.MethodPost)
	guestOrderRouter.HandleFunc("/guest/orders/{id}/items/{productID}", guestOrderHandler.UpdateItemQuantity).Methods(http.MethodPut)
	guestOrderRouter.HandleFunc("/guest/orders/{id}/items/{productID}", guestOrderHandler.RemoveItem).Methods(http.MethodDelete)
	guestOrderRouter.HandleFunc("/guest/orders/{id}/checkout", guestOrderHandler.Checkout).Methods(http.MethodPost)
	guestOrderRouter.HandleFunc("/guest/orders/{id}/cancel", guestOrderHandler.CancelOrder).Methods(http.MethodPost)

	// Protected routes for order management (staff only)
	orderRouter := protectedRouter.PathPrefix("/orders").Subrouter()
	orderRouter.HandleFunc("", guestOrderHandler.ListOrders).Methods(http.MethodGet)
	orderRouter.HandleFunc("/pending", guestOrderHandler.GetPendingOrders).Methods(http.MethodGet)
	orderRouter.HandleFunc("/active", guestOrderHandler.GetActiveOrders).Methods(http.MethodGet)
	orderRouter.HandleFunc("/{id}", guestOrderHandler.GetOrder).Methods(http.MethodGet)
	orderRouter.HandleFunc("/{id}/status", guestOrderHandler.UpdateOrderStatus).Methods(http.MethodPost)
	orderRouter.HandleFunc("/{id}/cancel", guestOrderHandler.CancelOrder).Methods(http.MethodPost)
	orderRouter.HandleFunc("/table/{tableID}", guestOrderHandler.GetOrdersByTable).Methods(http.MethodGet)

	// Reports routes (admin only)
	reportRouter := protectedRouter.PathPrefix("/reports").Subrouter()
	reportRouter.Use(authMiddleware.RequireRole("SUPER_ADMIN", "ADMIN"))

	reportRouter.HandleFunc("/sales/today", guestOrderHandler.GetTodaySales).Methods(http.MethodGet)

	// Root endpoint - API info
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"service": "jwt-ddd-clean-pos",
			"version": "3.0.0",
			"endpoints": map[string][]string{
				"public": {
					"POST /api/auth/login",
					"POST /api/auth/register",
					"POST /api/auth/refresh",
					"POST /api/guest/orders",
					"GET  /api/guest/orders/{id}",
					"POST /api/guest/orders/{id}/items",
					"PUT  /api/guest/orders/{id}/items/{productID}",
					"DELETE /api/guest/orders/{id}/items/{productID}",
					"POST /api/guest/orders/{id}/checkout",
					"POST /api/guest/orders/{id}/cancel",
					"GET  /api/health",
				},
				"protected": {
					"POST   /api/auth/logout",
					"GET    /api/auth/me",
					"POST   /api/auth/change-password",
					"GET    /api/inventory",
					"POST   /api/inventory",
					"GET    /api/inventory/{id}",
					"PUT    /api/inventory/{id}",
					"DELETE /api/inventory/{id}",
					"POST   /api/pos/cart",
					"GET    /api/pos/cart/my",
					"GET    /api/pos/cart/{id}",
					"POST   /api/pos/cart/{id}/items",
					"PUT    /api/pos/cart/{id}/items",
					"DELETE /api/pos/cart/{id}/items",
					"POST   /api/pos/checkout/{id}",
					"GET    /api/pos/transactions",
					"GET    /api/pos/transactions/{id}",
					"POST   /api/pos/transactions/{id}/cancel",
					"GET    /api/pos/sales/today",
					"GET    /api/orders",
					"GET    /api/orders/pending",
					"GET    /api/orders/active",
					"GET    /api/orders/{id}",
					"POST   /api/orders/{id}/status",
					"POST   /api/orders/{id}/cancel",
					"GET    /api/orders/table/{tableID}",
				},
				"admin_only": {
					"GET    /api/admin/users",
					"GET    /api/admin/users/{id}",
					"PUT    /api/admin/users/{id}",
					"DELETE /api/admin/users/{id}",
					"GET    /api/tables",
					"POST   /api/tables",
					"GET    /api/tables/available",
					"GET    /api/tables/{id}",
					"PUT    /api/tables/{id}",
					"DELETE /api/tables/{id}",
					"POST   /api/tables/{id}/status",
					"POST   /api/tables/{id}/qr",
					"GET    /api/reports/sales/today",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}).Methods(http.MethodGet)
}

// buildApp wires all layers: infrastructure -> domain -> application -> handlers
func buildApp(
	tokenRepo repository.TokenRepository,
	inventoryRepo repository.InventoryRepository,
	userRepo repository.UserRepository,
	cartRepo repository.CartRepository,
	transactionRepo repository.TransactionRepository,
	tableRepo repository.TableRepository,
	guestOrderRepo repository.GuestOrderRepository,
	tenantRepo repository.TenantRepository,
	planRepo repository.SubscriptionPlanRepository,
	rawMaterialRepo repository.RawMaterialRepository,
	recipeRepo repository.ProductRecipeRepository,
	usageRepo repository.TenantUsageRepository,
	jwtProvider *jwt.Provider,
	accessTokenTTL, refreshTokenTTL time.Duration,
) (*TokenHTTPHandler, *inventoryhttp.InventoryHTTPHandler, *handler.AuthHandler, *handler.POSHandler, *handler.TableHandler, *handler.GuestOrderHandler, *handler.TenantHandler, *handler.HealthHandler, *httpmiddleware.AuthMiddleware, *httpmiddleware.TenantMiddleware) {
	// Domain layer - Services
	tokenService := service.NewTokenService(
		tokenRepo,
		jwtProvider,
		accessTokenTTL,
		refreshTokenTTL,
	)

	inventoryService := service.NewInventoryService(inventoryRepo)
	authService := service.NewAuthService(userRepo, tokenRepo, jwtProvider)
	posService := service.NewPOSService(cartRepo, transactionRepo, inventoryRepo)

	// QR Code Service (TODO: Configure with actual logo path and base URL)
	qrConfig := service.QRCodeConfig{
		BaseURL:         "http://localhost:8080",
		MerchantName:    "POS Restaurant",
		Size:            256,
		ErrorCorrection: 1, // Medium
	}
	qrService := service.NewQRCodeService(qrConfig)

	// Table & Guest Order Services
	tableService := service.NewTableService(tableRepo, qrService)
	guestOrderService := service.NewGuestOrderService(guestOrderRepo, tableRepo, inventoryRepo, 11.0) // 11% default tax

	// Application layer - Usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, tokenRepo, authService)
	inventoryUsecase := usecase.NewInventoryUsecase(inventoryRepo, inventoryService)
	posUsecase := usecase.NewPOSUsecase(cartRepo, transactionRepo, inventoryRepo, posService)
	tokenUsecase := usecase.NewTokenUsecase(tokenService)
	tableUsecase := usecase.NewTableUsecase(tableService)
	guestOrderUsecase := usecase.NewGuestOrderUsecase(guestOrderService)

	// Handler layer
	tokenHandler := handler.NewTokenHandler(tokenUsecase)
	tokenHTTPHandler := NewTokenHTTPHandler(tokenHandler)

	inventoryHTTPHandler := inventoryhttp.NewInventoryHTTPHandler(inventoryUsecase)

	authHandler := handler.NewAuthHandler(authUsecase)

	posHandler := handler.NewPOSHandler(posUsecase)

	tableHandler := handler.NewTableHandler(tableUsecase)

	guestOrderHandler := handler.NewGuestOrderHandler(guestOrderUsecase)

	// Tenant usecase and handler
	tenantUsecase := usecase.NewTenantUsecase(tenantRepo, planRepo, rawMaterialRepo, recipeRepo, usageRepo, userRepo)
	tenantHandler := handler.NewTenantHandler(tenantUsecase)

	healthHandler := handler.NewHealthHandler("3.0.0")

	// Middleware (still uses domain token service for low-level validation)
	authMiddleware := httpmiddleware.NewAuthMiddleware(tokenService)
	tenantMiddleware := httpmiddleware.NewTenantMiddleware(tenantRepo)

	return tokenHTTPHandler, inventoryHTTPHandler, authHandler, posHandler, tableHandler, guestOrderHandler, tenantHandler, healthHandler, authMiddleware, tenantMiddleware
}

// NewServer creates a new HTTP server with gorilla/mux
func NewServer(config ServerConfig) *Server {
	// Infrastructure layer - JWT
	jwtProvider := jwt.NewProvider(jwt.Config{
		SecretKey: config.SecretKey,
		Issuer:    config.Issuer,
		Algorithm: "HS256",
	})

	// Infrastructure layer - Repositories (In-Memory)
	var tokenRepo repository.TokenRepository = infrarepo.NewMemoryTokenRepository()
	var inventoryRepo repository.InventoryRepository = infrarepo.NewMemoryInventoryRepository()
	var userRepo repository.UserRepository = infrarepo.NewMemoryUserRepository()
	var cartRepo repository.CartRepository = infrarepo.NewMemoryCartRepository()
	var transactionRepo repository.TransactionRepository = infrarepo.NewMemoryTransactionRepository()
	var tableRepo repository.TableRepository = infrarepo.NewMemoryTableRepository()
	var guestOrderRepo repository.GuestOrderRepository = infrarepo.NewMemoryGuestOrderRepository()
	var tenantRepo repository.TenantRepository = infrarepo.NewMemoryTenantRepository()
	var planRepo repository.SubscriptionPlanRepository = infrarepo.NewMemorySubscriptionPlanRepository()
	var rawMaterialRepo repository.RawMaterialRepository = infrarepo.NewMemoryRawMaterialRepository()
	var recipeRepo repository.ProductRecipeRepository = infrarepo.NewMemoryProductRecipeRepository()
	var usageRepo repository.TenantUsageRepository = infrarepo.NewMemoryTenantUsageRepository()

	// Build application layers
	tokenHTTPHandler, inventoryHTTPHandler, authHandler, posHandler, tableHandler, guestOrderHandler, tenantHandler, healthHandler, authMiddleware, tenantMiddleware := buildApp(
		tokenRepo, inventoryRepo, userRepo, cartRepo, transactionRepo, tableRepo, guestOrderRepo,
		tenantRepo, planRepo, rawMaterialRepo, recipeRepo, usageRepo,
		jwtProvider, config.AccessTokenTTL, config.RefreshTokenTTL,
	)

	// Create mux router
	r := mux.NewRouter()

	// Setup routes
	setupRoutes(r, tokenHTTPHandler, inventoryHTTPHandler, authHandler, posHandler, tableHandler, guestOrderHandler, tenantHandler, healthHandler, authMiddleware, tenantMiddleware)

	server := &http.Server{
		Addr:         config.Host + ":" + config.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: server,
		router:     r,
		config:     config,
	}
}

// NewServerWithDatabase creates a new HTTP server with PostgreSQL database connection using gorilla/mux
func NewServerWithDatabase(config ServerConfig, db *sql.DB) *Server {
	// Infrastructure layer - JWT
	jwtProvider := jwt.NewProvider(jwt.Config{
		SecretKey: config.SecretKey,
		Issuer:    config.Issuer,
		Algorithm: "HS256",
	})

	// Infrastructure layer - Repositories (PostgreSQL)
	var tokenRepo repository.TokenRepository = infrarepo.NewMemoryTokenRepository() // TODO: Implement PostgreSQL token repository
	var inventoryRepo repository.InventoryRepository = infrarepo.NewPostgresInventoryRepository(db)
	var userRepo repository.UserRepository = infrarepo.NewPostgresUserRepository(db)
	var cartRepo repository.CartRepository = infrarepo.NewPostgresCartRepository(db)
	var transactionRepo repository.TransactionRepository = infrarepo.NewPostgresTransactionRepository(db)
	var tableRepo repository.TableRepository = infrarepo.NewPostgresTableRepository(db)
	var guestOrderRepo repository.GuestOrderRepository = infrarepo.NewPostgresGuestOrderRepository(db)
	var tenantRepo repository.TenantRepository = infrarepo.NewPostgresTenantRepository(db)
	var planRepo repository.SubscriptionPlanRepository = infrarepo.NewPostgresSubscriptionPlanRepository(db)
	var rawMaterialRepo repository.RawMaterialRepository = infrarepo.NewPostgresRawMaterialRepository(db)
	var recipeRepo repository.ProductRecipeRepository = infrarepo.NewPostgresProductRecipeRepository(db)
	var usageRepo repository.TenantUsageRepository = infrarepo.NewMemoryTenantUsageRepository() // TODO: Implement PostgreSQL usage repository

	// Build application layers
	tokenHTTPHandler, inventoryHTTPHandler, authHandler, posHandler, tableHandler, guestOrderHandler, tenantHandler, healthHandler, authMiddleware, tenantMiddleware := buildApp(
		tokenRepo, inventoryRepo, userRepo, cartRepo, transactionRepo, tableRepo, guestOrderRepo,
		tenantRepo, planRepo, rawMaterialRepo, recipeRepo, usageRepo,
		jwtProvider, config.AccessTokenTTL, config.RefreshTokenTTL,
	)

	// Create mux router
	r := mux.NewRouter()

	// Setup routes
	setupRoutes(r, tokenHTTPHandler, inventoryHTTPHandler, authHandler, posHandler, tableHandler, guestOrderHandler, tenantHandler, healthHandler, authMiddleware, tenantMiddleware)

	server := &http.Server{
		Addr:         config.Host + ":" + config.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: server,
		router:     r,
		config:     config,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	fmt.Printf("🚀 Starting JWT API Server on http://%s:%s\n", s.config.Host, s.config.Port)
	fmt.Printf("📋 Available endpoints:\n")
	fmt.Printf("   Public Endpoints:\n")
	fmt.Printf("     POST /api/token/generate - Generate new tokens\n")
	fmt.Printf("     POST /api/token/refresh  - Refresh access token\n")
	fmt.Printf("     POST /api/token/validate - Validate token\n")
	fmt.Printf("     POST /api/token/revoke   - Revoke token\n")
	fmt.Printf("     GET  /api/health         - Health check\n")
	fmt.Printf("   Protected Endpoints (require authentication):\n")
	fmt.Printf("     GET    /api/inventory           - List inventory items\n")
	fmt.Printf("     POST   /api/inventory           - Create inventory item\n")
	fmt.Printf("     GET    /api/inventory/{id}      - Get inventory item\n")
	fmt.Printf("     PUT    /api/inventory/{id}      - Update inventory item\n")
	fmt.Printf("     DELETE /api/inventory/{id}      - Delete inventory item\n")
	fmt.Printf("     PUT    /api/inventory/{id}/stock - Update stock quantity\n")
	fmt.Printf("     POST   /api/inventory/{id}/stock/adjust - Adjust stock\n")
	fmt.Println()

	return s.httpServer.ListenAndServe()
}

// StartWithGracefulShutdown starts the server with graceful shutdown
func (s *Server) StartWithGracefulShutdown() error {
	// Start server in goroutine
	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("❌ Server failed to start: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		fmt.Printf("❌ Server forced to shutdown: %v\n", err)
		return err
	}

	fmt.Println("✅ Server stopped gracefully")
	return nil
}

// Stop stops the HTTP server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}
