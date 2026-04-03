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
	authMiddleware *httpmiddleware.AuthMiddleware,
) {
	// Public routes (no authentication required)
	r.HandleFunc("/api/auth/login", authHandler.Login).Methods(http.MethodPost)
	r.HandleFunc("/api/auth/register", authHandler.Register).Methods(http.MethodPost)
	r.HandleFunc("/api/auth/refresh", authHandler.RefreshToken).Methods(http.MethodPost)
	r.HandleFunc("/api/token/generate", tokenHTTPHandler.GenerateToken).Methods(http.MethodPost)
	r.HandleFunc("/api/token/refresh", tokenHTTPHandler.RefreshToken).Methods(http.MethodPost)
	r.HandleFunc("/api/token/validate", tokenHTTPHandler.ValidateToken).Methods(http.MethodPost)
	r.HandleFunc("/api/token/revoke", tokenHTTPHandler.RevokeToken).Methods(http.MethodPost)
	r.HandleFunc("/api/health", tokenHTTPHandler.Health).Methods(http.MethodGet)

	// Protected routes (authentication required)
	protectedRouter := r.PathPrefix("/api").Subrouter()
	protectedRouter.Use(authMiddleware.Authenticate)

	// Auth routes (require authentication)
	protectedRouter.HandleFunc("/auth/logout", authHandler.Logout).Methods(http.MethodPost)
	protectedRouter.HandleFunc("/auth/me", authHandler.GetMe).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/auth/change-password", authHandler.ChangePassword).Methods(http.MethodPost)

	// Admin routes (require admin role)
	adminRouter := protectedRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(authMiddleware.RequireRole("SUPER_ADMIN", "ADMIN"))

	adminRouter.HandleFunc("/users", authHandler.ListUsers).Methods(http.MethodGet)
	adminRouter.HandleFunc("/users", authHandler.CreateUser).Methods(http.MethodPost)
	adminRouter.HandleFunc("/users/{id}", authHandler.GetUserByID).Methods(http.MethodGet)
	adminRouter.HandleFunc("/users/{id}", authHandler.UpdateUser).Methods(http.MethodPut)
	adminRouter.HandleFunc("/users/{id}", authHandler.DeleteUser).Methods(http.MethodDelete)

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

	// Sales summary
	posRouter.HandleFunc("/sales/today", posHandler.GetTodaySales).Methods(http.MethodGet)

	// Root endpoint - API info
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"service": "jwt-ddd-clean-pos",
			"version": "2.0.0",
			"endpoints": map[string][]string{
				"public": {
					"POST /api/auth/login",
					"POST /api/auth/register",
					"POST /api/auth/refresh",
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
				},
				"admin_only": {
					"GET    /api/admin/users",
					"GET    /api/admin/users/{id}",
					"PUT    /api/admin/users/{id}",
					"DELETE /api/admin/users/{id}",
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
	jwtProvider *jwt.Provider,
	accessTokenTTL, refreshTokenTTL time.Duration,
) (*TokenHTTPHandler, *inventoryhttp.InventoryHTTPHandler, *handler.AuthHandler, *handler.POSHandler, *httpmiddleware.AuthMiddleware) {
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

	// Application layer - Usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, tokenRepo, authService)
	inventoryUsecase := usecase.NewInventoryUsecase(inventoryRepo, inventoryService)
	posUsecase := usecase.NewPOSUsecase(cartRepo, transactionRepo, inventoryRepo, posService)
	tokenUsecase := usecase.NewTokenUsecase(tokenService)

	// Handler layer
	tokenHandler := handler.NewTokenHandler(tokenUsecase)
	tokenHTTPHandler := NewTokenHTTPHandler(tokenHandler)

	inventoryHTTPHandler := inventoryhttp.NewInventoryHTTPHandler(inventoryUsecase)

	authHandler := handler.NewAuthHandler(authUsecase)

	posHandler := handler.NewPOSHandler(posUsecase)

	// Middleware (still uses domain token service for low-level validation)
	authMiddleware := httpmiddleware.NewAuthMiddleware(tokenService)

	return tokenHTTPHandler, inventoryHTTPHandler, authHandler, posHandler, authMiddleware
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

	// Build application layers
	tokenHTTPHandler, inventoryHTTPHandler, authHandler, posHandler, authMiddleware := buildApp(
		tokenRepo, inventoryRepo, userRepo, cartRepo, transactionRepo,
		jwtProvider, config.AccessTokenTTL, config.RefreshTokenTTL,
	)

	// Create mux router
	r := mux.NewRouter()

	// Setup routes
	setupRoutes(r, tokenHTTPHandler, inventoryHTTPHandler, authHandler, posHandler, authMiddleware)

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
	var tokenRepo repository.TokenRepository = infrarepo.NewMemoryTokenRepository()
	var inventoryRepo repository.InventoryRepository = infrarepo.NewPostgresInventoryRepository(db)
	var userRepo repository.UserRepository = infrarepo.NewPostgresUserRepository(db)
	var cartRepo repository.CartRepository = infrarepo.NewMemoryCartRepository() // TODO: Implement PostgreSQL cart repository
	var transactionRepo repository.TransactionRepository = infrarepo.NewMemoryTransactionRepository() // TODO: Implement PostgreSQL transaction repository

	// Build application layers
	tokenHTTPHandler, inventoryHTTPHandler, authHandler, posHandler, authMiddleware := buildApp(
		tokenRepo, inventoryRepo, userRepo, cartRepo, transactionRepo,
		jwtProvider, config.AccessTokenTTL, config.RefreshTokenTTL,
	)

	// Create mux router
	r := mux.NewRouter()

	// Setup routes
	setupRoutes(r, tokenHTTPHandler, inventoryHTTPHandler, authHandler, posHandler, authMiddleware)

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
