package model

import (
	"fmt"
	"time"
)

// PaymentMethod represents the method of payment
type PaymentMethod string

const (
	PaymentCash     PaymentMethod = "CASH"
	PaymentCard     PaymentMethod = "CARD"
	PaymentQRIS     PaymentMethod = "QRIS"
	PaymentEWallet  PaymentMethod = "E_WALLET"
	PaymentTransfer PaymentMethod = "TRANSFER"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionPending   TransactionStatus = "PENDING"
	TransactionCompleted TransactionStatus = "COMPLETED"
	TransactionCancelled TransactionStatus = "CANCELLED"
	TransactionRefunded  TransactionStatus = "REFUNDED"
)

// TransactionItem represents an item in a transaction
type TransactionItem struct {
	productID   string
	productName string
	sku         string
	quantity    int
	unitPrice   float64
	subtotal    float64
}

// NewTransactionItem creates a transaction item
func NewTransactionItem(productID, productName, sku string, quantity int, unitPrice float64) (*TransactionItem, error) {
	if productID == "" {
		return nil, fmt.Errorf("product ID tidak boleh kosong")
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity harus lebih dari 0")
	}
	if unitPrice < 0 {
		return nil, fmt.Errorf("harga tidak boleh negatif")
	}

	return &TransactionItem{
		productID:   productID,
		productName: productName,
		sku:         sku,
		quantity:    quantity,
		unitPrice:   unitPrice,
		subtotal:    float64(quantity) * unitPrice,
	}, nil
}

// ReconstructTransactionItem recreates from database
func ReconstructTransactionItem(
	productID, productName, sku string,
	quantity int,
	unitPrice, subtotal float64,
) *TransactionItem {
	return &TransactionItem{
		productID:   productID,
		productName: productName,
		sku:         sku,
		quantity:    quantity,
		unitPrice:   unitPrice,
		subtotal:    subtotal,
	}
}

// Accessors
func (i *TransactionItem) ProductID() string   { return i.productID }
func (i *TransactionItem) ProductName() string { return i.productName }
func (i *TransactionItem) SKU() string         { return i.sku }
func (i *TransactionItem) Quantity() int       { return i.quantity }
func (i *TransactionItem) UnitPrice() float64  { return i.unitPrice }
func (i *TransactionItem) Subtotal() float64   { return i.subtotal }

// Transaction represents a sales transaction aggregate root
type Transaction struct {
	id              string
	transactionNo   string
	cashierID       string
	cashierName     string
	customerName    string
	items           []TransactionItem
	subtotal        float64
	discountAmount  float64
	discountPercent float64
	taxAmount       float64
	taxPercent      float64
	totalAmount     float64
	paymentMethod   PaymentMethod
	paymentAmount   float64
	changeAmount    float64
	status          TransactionStatus
	notes           string
	createdAt       time.Time
	updatedAt       time.Time
}

// NewTransaction creates a new transaction
func NewTransaction(transactionNo string, cashierID, cashierName string) (*Transaction, error) {
	if transactionNo == "" {
		return nil, fmt.Errorf("nomor transaksi tidak boleh kosong")
	}
	if cashierID == "" {
		return nil, fmt.Errorf("cashier ID tidak boleh kosong")
	}

	now := time.Now()
	return &Transaction{
		transactionNo: transactionNo,
		cashierID:     cashierID,
		cashierName:   cashierName,
		items:         make([]TransactionItem, 0),
		status:        TransactionPending,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

// ReconstructTransaction recreates a transaction from database
func ReconstructTransaction(
	id, transactionNo, cashierID, cashierName, customerName string,
	items []TransactionItem,
	subtotal, discountAmount, discountPercent, taxAmount, taxPercent,
	totalAmount, paymentAmount, changeAmount float64,
	paymentMethod PaymentMethod,
	status TransactionStatus,
	notes string,
	createdAt, updatedAt time.Time,
) *Transaction {
	return &Transaction{
		id:              id,
		transactionNo:   transactionNo,
		cashierID:       cashierID,
		cashierName:     cashierName,
		customerName:    customerName,
		items:           items,
		subtotal:        subtotal,
		discountAmount:  discountAmount,
		discountPercent: discountPercent,
		taxAmount:       taxAmount,
		taxPercent:      taxPercent,
		totalAmount:     totalAmount,
		paymentMethod:   paymentMethod,
		paymentAmount:   paymentAmount,
		changeAmount:    changeAmount,
		status:          status,
		notes:           notes,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

// Accessors
func (t *Transaction) ID() string                { return t.id }
func (t *Transaction) TransactionNo() string     { return t.transactionNo }
func (t *Transaction) CashierID() string         { return t.cashierID }
func (t *Transaction) CashierName() string       { return t.cashierName }
func (t *Transaction) CustomerName() string      { return t.customerName }
func (t *Transaction) Items() []TransactionItem  { return t.items }
func (t *Transaction) Subtotal() float64         { return t.subtotal }
func (t *Transaction) DiscountAmount() float64   { return t.discountAmount }
func (t *Transaction) DiscountPercent() float64  { return t.discountPercent }
func (t *Transaction) TaxAmount() float64        { return t.taxAmount }
func (t *Transaction) TaxPercent() float64       { return t.taxPercent }
func (t *Transaction) TotalAmount() float64      { return t.totalAmount }
func (t *Transaction) PaymentMethod() PaymentMethod { return t.paymentMethod }
func (t *Transaction) PaymentAmount() float64    { return t.paymentAmount }
func (t *Transaction) ChangeAmount() float64     { return t.changeAmount }
func (t *Transaction) Status() TransactionStatus { return t.status }
func (t *Transaction) Notes() string             { return t.notes }
func (t *Transaction) CreatedAt() time.Time      { return t.createdAt }
func (t *Transaction) UpdatedAt() time.Time      { return t.updatedAt }

// SetCustomerName sets the customer name
func (t *Transaction) SetCustomerName(name string) {
	t.customerName = name
	t.updatedAt = time.Now()
}

// SetNotes sets the transaction notes
func (t *Transaction) SetNotes(notes string) {
	t.notes = notes
	t.updatedAt = time.Now()
}

// AddItem adds an item to the transaction
func (t *Transaction) AddItem(productID, productName, sku string, quantity int, unitPrice float64) error {
	item, err := NewTransactionItem(productID, productName, sku, quantity, unitPrice)
	if err != nil {
		return err
	}
	t.items = append(t.items, *item)
	t.recalculateTotals()
	t.updatedAt = time.Now()
	return nil
}

// RecalculateTotals recalculates all totals
func (t *Transaction) RecalculateTotals() {
	t.recalculateTotals()
}

func (t *Transaction) recalculateTotals() {
	t.subtotal = 0
	for _, item := range t.items {
		t.subtotal += item.subtotal
	}

	t.totalAmount = t.subtotal - t.discountAmount + t.taxAmount
}

// ApplyDiscount applies a discount
func (t *Transaction) ApplyDiscount(amount float64, percent float64) {
	t.discountAmount = amount
	t.discountPercent = percent
	t.recalculateTotals()
}

// ApplyTax applies tax
func (t *Transaction) ApplyTax(percent float64) {
	t.taxPercent = percent
	t.taxAmount = t.subtotal * (percent / 100)
	t.recalculateTotals()
}

// Complete completes the transaction with payment
func (t *Transaction) Complete(method PaymentMethod, amountPaid float64) error {
	if t.status != TransactionPending {
		return fmt.Errorf("transaksi tidak dapat diproses, status: %s", t.status)
	}

	if len(t.items) == 0 {
		return fmt.Errorf("transaksi harus memiliki item")
	}

	t.paymentMethod = method
	t.paymentAmount = amountPaid
	t.changeAmount = amountPaid - t.totalAmount

	// Payment must be sufficient
	if t.changeAmount < 0 {
		return fmt.Errorf("pembayaran tidak mencukupi, kurang: %.2f", -t.changeAmount)
	}

	t.status = TransactionCompleted
	t.updatedAt = time.Now()
	return nil
}

// Cancel cancels the transaction
func (t *Transaction) Cancel() error {
	if t.status == TransactionCancelled {
		return fmt.Errorf("transaksi sudah dibatalkan")
	}
	t.status = TransactionCancelled
	t.updatedAt = time.Now()
	return nil
}

// IsCompleted checks if transaction is completed
func (t *Transaction) IsCompleted() bool {
	return t.status == TransactionCompleted
}

// IsCancellable checks if transaction can be cancelled
func (t *Transaction) IsCancellable() bool {
	return t.status == TransactionCompleted
}
