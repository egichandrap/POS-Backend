package dto

import (
	"github.com/example/jwt-ddd-clean/internal/domain/model"
)

// AddToCartRequest represents the request to add item to cart
type AddToCartRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,min=1"`
}

// UpdateCartItemRequest represents the request to update cart item quantity
type UpdateCartItemRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,min=0"`
}

// CheckoutRequest represents the checkout request
type CheckoutRequest struct {
	PaymentMethod model.PaymentMethod `json:"payment_method" validate:"required,oneof=CASH CARD QRIS E_WALLET TRANSFER"`
	PaymentAmount float64             `json:"payment_amount" validate:"required,min=0"`
	CustomerName  string              `json:"customer_name,omitempty"`
	Notes         string              `json:"notes,omitempty"`
}

// CartItemResponse represents a cart item in response
type CartItemResponse struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Subtotal    float64 `json:"subtotal"`
}

// CartResponse represents the cart response
type CartResponse struct {
	ID           string           `json:"id"`
	UserID       string           `json:"user_id"`
	CustomerName string           `json:"customer_name,omitempty"`
	Items        []CartItemResponse `json:"items"`
	TotalAmount  float64          `json:"total_amount"`
	CreatedAt    string           `json:"created_at"`
	UpdatedAt    string           `json:"updated_at"`
}

// TransactionItemResponse represents a transaction item in response
type TransactionItemResponse struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Subtotal    float64 `json:"subtotal"`
}

// TransactionResponse represents the transaction response
type TransactionResponse struct {
	ID              string                    `json:"id"`
	TransactionNo   string                    `json:"transaction_no"`
	CashierID       string                    `json:"cashier_id"`
	CashierName     string                    `json:"cashier_name"`
	CustomerName    string                    `json:"customer_name,omitempty"`
	Items           []TransactionItemResponse `json:"items"`
	Subtotal        float64                   `json:"subtotal"`
	DiscountAmount  float64                   `json:"discount_amount"`
	DiscountPercent float64                   `json:"discount_percent"`
	TaxAmount       float64                   `json:"tax_amount"`
	TaxPercent      float64                   `json:"tax_percent"`
	TotalAmount     float64                   `json:"total_amount"`
	PaymentMethod   string                    `json:"payment_method"`
	PaymentAmount   float64                   `json:"payment_amount"`
	ChangeAmount    float64                   `json:"change_amount"`
	Status          string                    `json:"status"`
	Notes           string                    `json:"notes,omitempty"`
	CreatedAt       string                    `json:"created_at"`
}

// TransactionListResponse represents paginated transaction list
type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int64                 `json:"total"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
	TotalPages   int                   `json:"total_pages"`
}

// SalesSummaryResponse represents daily sales summary
type SalesSummaryResponse struct {
	TotalSales        float64 `json:"total_sales"`
	TotalTransactions int     `json:"total_transactions"`
	TotalItems        int     `json:"total_items"`
	Date              string  `json:"date"`
}

// ToCartResponse converts Cart model to DTO
func ToCartResponse(cart *model.Cart) *CartResponse {
	items := make([]CartItemResponse, len(cart.Items()))
	for i, item := range cart.Items() {
		items[i] = CartItemResponse{
			ProductID:   item.ProductID(),
			ProductName: item.ProductName(),
			SKU:         item.SKU(),
			Quantity:    item.Quantity(),
			UnitPrice:   item.UnitPrice(),
			Subtotal:    item.Subtotal(),
		}
	}

	return &CartResponse{
		ID:           cart.ID(),
		UserID:       cart.UserID(),
		CustomerName: cart.CustomerName(),
		Items:        items,
		TotalAmount:  cart.TotalAmount(),
		CreatedAt:    cart.CreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    cart.UpdatedAt().Format("2006-01-02T15:04:05Z"),
	}
}

// ToTransactionResponse converts Transaction model to DTO
func ToTransactionResponse(transaction *model.Transaction) *TransactionResponse {
	items := make([]TransactionItemResponse, len(transaction.Items()))
	for i, item := range transaction.Items() {
		items[i] = TransactionItemResponse{
			ProductID:   item.ProductID(),
			ProductName: item.ProductName(),
			SKU:         item.SKU(),
			Quantity:    item.Quantity(),
			UnitPrice:   item.UnitPrice(),
			Subtotal:    item.Subtotal(),
		}
	}

	return &TransactionResponse{
		ID:              transaction.ID(),
		TransactionNo:   transaction.TransactionNo(),
		CashierID:       transaction.CashierID(),
		CashierName:     transaction.CashierName(),
		CustomerName:    transaction.CustomerName(),
		Items:           items,
		Subtotal:        transaction.Subtotal(),
		DiscountAmount:  transaction.DiscountAmount(),
		DiscountPercent: transaction.DiscountPercent(),
		TaxAmount:       transaction.TaxAmount(),
		TaxPercent:      transaction.TaxPercent(),
		TotalAmount:     transaction.TotalAmount(),
		PaymentMethod:   string(transaction.PaymentMethod()),
		PaymentAmount:   transaction.PaymentAmount(),
		ChangeAmount:    transaction.ChangeAmount(),
		Status:          string(transaction.Status()),
		Notes:           transaction.Notes(),
		CreatedAt:       transaction.CreatedAt().Format("2006-01-02T15:04:05Z"),
	}
}
