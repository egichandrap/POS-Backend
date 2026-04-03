package usecase

import (
	"context"
	"time"

	"github.com/example/jwt-ddd-clean/internal/application/dto"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
	"github.com/example/jwt-ddd-clean/internal/domain/service"
)

// POSUsecase defines the POS usecase interface
type POSUsecase interface {
	CreateCart(ctx context.Context, userID, customerName string) (*dto.CartResponse, error)
	GetCart(ctx context.Context, cartID string) (*dto.CartResponse, error)
	GetOrCreateCart(ctx context.Context, userID string) (*dto.CartResponse, error)
	AddToCart(ctx context.Context, cartID string, req dto.AddToCartRequest) (*dto.CartResponse, error)
	UpdateCartItemQuantity(ctx context.Context, cartID string, req dto.UpdateCartItemRequest) (*dto.CartResponse, error)
	RemoveFromCart(ctx context.Context, cartID, productID string) (*dto.CartResponse, error)
	ClearCart(ctx context.Context, cartID string) error
	DeleteCart(ctx context.Context, cartID string) error
	Checkout(ctx context.Context, cartID string, req dto.CheckoutRequest, cashierName string) (*dto.TransactionResponse, error)
	GetTransaction(ctx context.Context, transactionID string) (*dto.TransactionResponse, error)
	ListTransactions(ctx context.Context, filter repository.TransactionFilter) (*dto.TransactionListResponse, error)
	GetTodaySales(ctx context.Context) (*dto.SalesSummaryResponse, error)
	CancelTransaction(ctx context.Context, transactionID string) (*dto.TransactionResponse, error)
}

type posUsecase struct {
	cartRepo        repository.CartRepository
	transactionRepo repository.TransactionRepository
	inventoryRepo   repository.InventoryRepository
	posService      *service.POSService
}

// NewPOSUsecase creates a new POSUsecase
func NewPOSUsecase(
	cartRepo repository.CartRepository,
	transactionRepo repository.TransactionRepository,
	inventoryRepo repository.InventoryRepository,
	posService *service.POSService,
) POSUsecase {
	return &posUsecase{
		cartRepo:        cartRepo,
		transactionRepo: transactionRepo,
		inventoryRepo:   inventoryRepo,
		posService:      posService,
	}
}

func (u *posUsecase) CreateCart(ctx context.Context, userID, customerName string) (*dto.CartResponse, error) {
	cart, err := u.posService.CreateCart(ctx, userID, customerName)
	if err != nil {
		return nil, err
	}

	return dto.ToCartResponse(cart), nil
}

func (u *posUsecase) GetCart(ctx context.Context, cartID string) (*dto.CartResponse, error) {
	cart, err := u.posService.GetCart(ctx, cartID)
	if err != nil {
		return nil, err
	}

	return dto.ToCartResponse(cart), nil
}

func (u *posUsecase) GetOrCreateCart(ctx context.Context, userID string) (*dto.CartResponse, error) {
	cart, err := u.posService.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	return dto.ToCartResponse(cart), nil
}

func (u *posUsecase) AddToCart(ctx context.Context, cartID string, req dto.AddToCartRequest) (*dto.CartResponse, error) {
	cart, err := u.posService.AddToCart(ctx, cartID, req.ProductID, req.Quantity)
	if err != nil {
		return nil, err
	}

	return dto.ToCartResponse(cart), nil
}

func (u *posUsecase) UpdateCartItemQuantity(ctx context.Context, cartID string, req dto.UpdateCartItemRequest) (*dto.CartResponse, error) {
	cart, err := u.posService.UpdateCartItemQuantity(ctx, cartID, req.ProductID, req.Quantity)
	if err != nil {
		return nil, err
	}

	return dto.ToCartResponse(cart), nil
}

func (u *posUsecase) RemoveFromCart(ctx context.Context, cartID, productID string) (*dto.CartResponse, error) {
	cart, err := u.posService.RemoveFromCart(ctx, cartID, productID)
	if err != nil {
		return nil, err
	}

	return dto.ToCartResponse(cart), nil
}

func (u *posUsecase) ClearCart(ctx context.Context, cartID string) error {
	return u.posService.ClearCart(ctx, cartID)
}

func (u *posUsecase) DeleteCart(ctx context.Context, cartID string) error {
	return u.posService.DeleteCart(ctx, cartID)
}

func (u *posUsecase) Checkout(ctx context.Context, cartID string, req dto.CheckoutRequest, cashierName string) (*dto.TransactionResponse, error) {
	// Delegate to domain service for complex checkout logic
	transaction, err := u.posService.Checkout(
		ctx,
		cartID,
		req.PaymentMethod,
		req.PaymentAmount,
		req.CustomerName,
		req.Notes,
	)
	if err != nil {
		return nil, err
	}

	return dto.ToTransactionResponse(transaction), nil
}

func (u *posUsecase) GetTransaction(ctx context.Context, transactionID string) (*dto.TransactionResponse, error) {
	transaction, err := u.posService.GetTransaction(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	return dto.ToTransactionResponse(transaction), nil
}

func (u *posUsecase) ListTransactions(ctx context.Context, filter repository.TransactionFilter) (*dto.TransactionListResponse, error) {
	paginated, err := u.posService.ListTransactions(ctx, filter)
	if err != nil {
		return nil, err
	}

	transactions := make([]dto.TransactionResponse, len(paginated.Transactions))
	for i, txn := range paginated.Transactions {
		transactions[i] = *dto.ToTransactionResponse(txn)
	}

	return &dto.TransactionListResponse{
		Transactions: transactions,
		Total:        paginated.Total,
		Limit:        paginated.Limit,
		Offset:       paginated.Offset,
		TotalPages:   paginated.TotalPages,
	}, nil
}

func (u *posUsecase) GetTodaySales(ctx context.Context) (*dto.SalesSummaryResponse, error) {
	sales, err := u.posService.GetTodaySales(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.SalesSummaryResponse{
		TotalSales:        sales["total_sales"].(float64),
		TotalTransactions: sales["total_transactions"].(int),
		TotalItems:        sales["total_items"].(int),
		Date:              sales["date"].(string),
	}, nil
}

func (u *posUsecase) CancelTransaction(ctx context.Context, transactionID string) (*dto.TransactionResponse, error) {
	transaction, err := u.posService.CancelTransaction(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	return dto.ToTransactionResponse(transaction), nil
}

// Helper function to generate simple IDs
func generateID(prefix string) string {
	return prefix + "-" + time.Now().Format("20060102150405")
}
