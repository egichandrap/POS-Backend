package service

import (
	"context"
	"time"

	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
	"github.com/example/jwt-ddd-clean/internal/pkg/errors"
)

// POSService handles Point of Sale business logic
type POSService struct {
	cartRepo        repository.CartRepository
	transactionRepo repository.TransactionRepository
	inventoryRepo   repository.InventoryRepository
}

// NewPOSService creates a new POSService
func NewPOSService(
	cartRepo repository.CartRepository,
	transactionRepo repository.TransactionRepository,
	inventoryRepo repository.InventoryRepository,
) *POSService {
	return &POSService{
		cartRepo:        cartRepo,
		transactionRepo: transactionRepo,
		inventoryRepo:   inventoryRepo,
	}
}

// CreateCart creates a new shopping cart
func (s *POSService) CreateCart(ctx context.Context, userID, customerName string) (*model.Cart, error) {
	cart, err := model.NewCart(userID, customerName)
	if err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	if err := s.cartRepo.Create(ctx, cart); err != nil {
		return nil, errors.NewInternalError("gagal membuat cart")
	}

	return cart, nil
}

// GetCart retrieves a cart by ID
func (s *POSService) GetCart(ctx context.Context, cartID string) (*model.Cart, error) {
	cart, err := s.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return nil, errors.NewNotFoundError("cart", "id", cartID)
	}
	return cart, nil
}

// GetOrCreateCart gets existing cart or creates new one
func (s *POSService) GetOrCreateCart(ctx context.Context, userID string) (*model.Cart, error) {
	cart, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		// Create new cart if not found
		return s.CreateCart(ctx, userID, "")
	}
	return cart, nil
}

// AddToCart adds an item to the cart
func (s *POSService) AddToCart(ctx context.Context, cartID, productID string, quantity int) (*model.Cart, error) {
	// Get cart
	cart, err := s.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return nil, errors.NewNotFoundError("cart", "id", cartID)
	}

	// Get product
	product, err := s.inventoryRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, errors.NewNotFoundError("produk", "id", productID)
	}

	// Check stock
	if product.Quantity() < quantity {
		return nil, errors.NewValidationError("stok tidak mencukupi untuk produk %s", product.Name())
	}

	// Add item to cart
	if err := cart.AddItem(product.ID(), product.Name(), product.SKU(), quantity, product.Price()); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	if err := s.cartRepo.Update(ctx, cart); err != nil {
		return nil, errors.NewInternalError("gagal menambahkan item ke cart")
	}

	return cart, nil
}

// RemoveFromCart removes an item from the cart
func (s *POSService) RemoveFromCart(ctx context.Context, cartID, productID string) (*model.Cart, error) {
	cart, err := s.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return nil, errors.NewNotFoundError("cart", "id", cartID)
	}

	cart.RemoveItem(productID)

	if err := s.cartRepo.Update(ctx, cart); err != nil {
		return nil, errors.NewInternalError("gagal menghapus item dari cart")
	}

	return cart, nil
}

// UpdateCartItemQuantity updates the quantity of an item in the cart
func (s *POSService) UpdateCartItemQuantity(ctx context.Context, cartID, productID string, quantity int) (*model.Cart, error) {
	cart, err := s.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return nil, errors.NewNotFoundError("cart", "id", cartID)
	}

	// Validate quantity
	if quantity < 0 {
		return nil, errors.NewValidationError("quantity tidak boleh negatif")
	}

	// If quantity is 0, remove the item
	if quantity == 0 {
		return s.RemoveFromCart(ctx, cartID, productID)
	}

	// Check stock availability
	if quantity > 0 {
		product, err := s.inventoryRepo.GetByID(ctx, productID)
		if err != nil {
			return nil, errors.NewNotFoundError("produk", "id", productID)
		}

		if product.Quantity() < quantity {
			return nil, errors.NewValidationError("stok tidak mencukupi untuk produk %s", product.Name())
		}
	}

	cart.UpdateItemQuantity(productID, quantity)

	if err := s.cartRepo.Update(ctx, cart); err != nil {
		return nil, errors.NewInternalError("gagal mengupdate quantity")
	}

	return cart, nil
}

// ClearCart removes all items from the cart
func (s *POSService) ClearCart(ctx context.Context, cartID string) error {
	cart, err := s.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return errors.NewNotFoundError("cart", "id", cartID)
	}

	cart.Clear()

	if err := s.cartRepo.Update(ctx, cart); err != nil {
		return errors.NewInternalError("gagal mengosongkan cart")
	}

	return nil
}

// DeleteCart deletes the entire cart
func (s *POSService) DeleteCart(ctx context.Context, cartID string) error {
	_, err := s.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return errors.NewNotFoundError("cart", "id", cartID)
	}

	if err := s.cartRepo.Delete(ctx, cartID); err != nil {
		return errors.NewInternalError("gagal menghapus cart")
	}

	return nil
}

// Checkout processes a transaction from cart
func (s *POSService) Checkout(ctx context.Context, cartID string, paymentMethod model.PaymentMethod, paymentAmount float64, customerName string, notes string) (*model.Transaction, error) {
	// Get cart
	cart, err := s.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return nil, errors.NewNotFoundError("cart", "id", cartID)
	}

	// Validate cart is not empty
	if len(cart.Items()) == 0 {
		return nil, errors.NewValidationError("cart kosong")
	}

	// Generate transaction number
	transactionNo, err := s.transactionRepo.GenerateTransactionNo(ctx)
	if err != nil {
		return nil, errors.NewInternalError("gagal membuat nomor transaksi")
	}

	// Get cashier info (from context - we'll pass this as parameter)
	cashierID := cart.UserID()
	cashierName := "" // Will be set from user service

	// Create transaction
	transaction, err := model.NewTransaction(transactionNo, cashierID, cashierName)
	if err != nil {
		return nil, errors.NewInternalError("gagal membuat transaksi")
	}
	transaction.SetCustomerName(customerName)
	transaction.SetNotes(notes)

	// Add items to transaction and update inventory
	for _, item := range cart.Items() {
		// Validate stock again
		product, err := s.inventoryRepo.GetByID(ctx, item.ProductID())
		if err != nil {
			return nil, errors.NewNotFoundError("produk", "id", item.ProductID())
		}

		if product.Quantity() < item.Quantity() {
			return nil, errors.NewValidationError("stok tidak mencukupi untuk produk %s", product.Name())
		}

		transaction.AddItem(item.ProductID(), item.ProductName(), item.SKU(), item.Quantity(), item.UnitPrice())

		// Update inventory
		newQty := product.Quantity() - item.Quantity()
		if err := s.inventoryRepo.UpdateQuantity(ctx, product.ID(), newQty); err != nil {
			return nil, errors.NewInternalError("gagal mengupdate stok produk %s", product.Name())
		}
	}

	// Apply default tax (11% PPN - can be configured)
	transaction.ApplyTax(11)

	// Process payment
	if err := transaction.Complete(paymentMethod, paymentAmount); err != nil {
		// Rollback inventory
		s.rollbackInventory(ctx, cart.Items())
		return nil, errors.NewValidationError("pembayaran tidak mencukupi")
	}

	// Save transaction
	if err := s.transactionRepo.Create(ctx, transaction); err != nil {
		// Rollback inventory
		s.rollbackInventory(ctx, cart.Items())
		return nil, errors.NewInternalError("gagal menyimpan transaksi")
	}

	// Clear and delete cart
	cart.Clear()
	s.cartRepo.Update(ctx, cart)
	cartIDVal := cart.ID()
	s.cartRepo.Delete(ctx, cartIDVal)

	return transaction, nil
}

// rollbackInventory rolls back inventory updates (in case of failure)
func (s *POSService) rollbackInventory(ctx context.Context, items []model.CartItem) {
	for _, item := range items {
		product, err := s.inventoryRepo.GetByID(ctx, item.ProductID())
		if err == nil {
			newQty := product.Quantity() + item.Quantity()
			s.inventoryRepo.UpdateQuantity(ctx, product.ID(), newQty)
		}
	}
}

// GetTransaction retrieves a transaction by ID
func (s *POSService) GetTransaction(ctx context.Context, transactionID string) (*model.Transaction, error) {
	transaction, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return nil, errors.NewNotFoundError("transaksi", "id", transactionID)
	}
	return transaction, nil
}

// ListTransactions retrieves paginated list of transactions
func (s *POSService) ListTransactions(ctx context.Context, filter repository.TransactionFilter) (*repository.PaginatedTransactions, error) {
	return s.transactionRepo.ListWithPagination(ctx, filter)
}

// GetTodaySales retrieves today's sales summary
func (s *POSService) GetTodaySales(ctx context.Context) (map[string]interface{}, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	transactions, err := s.transactionRepo.GetByDateRange(ctx, startOfDay, endOfDay)
	if err != nil {
		return nil, errors.NewInternalError("gagal mengambil data penjualan")
	}

	totalSales := 0.0
	totalTransactions := len(transactions)
	totalItems := 0

	for _, t := range transactions {
		if t.Status() == model.TransactionCompleted {
			totalSales += t.TotalAmount()
			totalItems += len(t.Items())
		}
	}

	return map[string]interface{}{
		"total_sales":        totalSales,
		"total_transactions": totalTransactions,
		"total_items":        totalItems,
		"date":               startOfDay.Format("2006-01-02"),
	}, nil
}

// CancelTransaction cancels a transaction and restores inventory
func (s *POSService) CancelTransaction(ctx context.Context, transactionID string) (*model.Transaction, error) {
	transaction, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return nil, errors.NewNotFoundError("transaksi", "id", transactionID)
	}

	if transaction.Status() != model.TransactionCompleted {
		return nil, errors.NewValidationError("hanya transaksi yang sudah selesai yang bisa dibatalkan")
	}

	// Restore inventory
	for _, item := range transaction.Items() {
		product, err := s.inventoryRepo.GetByID(ctx, item.ProductID())
		if err == nil {
			newQty := product.Quantity() + item.Quantity()
			s.inventoryRepo.UpdateQuantity(ctx, product.ID(), newQty)
		}
	}

	// Update transaction status
	transaction.Cancel()
	if err := s.transactionRepo.Update(ctx, transaction); err != nil {
		return nil, errors.NewInternalError("gagal membatalkan transaksi")
	}

	return transaction, nil
}
