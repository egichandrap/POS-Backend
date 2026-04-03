package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
)

// MemoryTransactionRepository implements repository.TransactionRepository for testing
type MemoryTransactionRepository struct {
	transactions   map[string]*model.Transaction
	transactionNum int
	mu             sync.RWMutex
}

// NewMemoryTransactionRepository creates a new MemoryTransactionRepository
func NewMemoryTransactionRepository() *MemoryTransactionRepository {
	return &MemoryTransactionRepository{
		transactions:   make(map[string]*model.Transaction),
		transactionNum: 0,
	}
}

// Create creates a new transaction
func (r *MemoryTransactionRepository) Create(ctx context.Context, transaction *model.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.transactions[transaction.ID()] = transaction
	return nil
}

// GetByID retrieves a transaction by ID
func (r *MemoryTransactionRepository) GetByID(ctx context.Context, id string) (*model.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transaction, exists := r.transactions[id]
	if !exists {
		return nil, fmt.Errorf("transaction not found")
	}

	return transaction, nil
}

// GetByTransactionNo retrieves a transaction by transaction number
func (r *MemoryTransactionRepository) GetByTransactionNo(ctx context.Context, transactionNo string) (*model.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, t := range r.transactions {
		if t.TransactionNo() == transactionNo {
			return t, nil
		}
	}

	return nil, fmt.Errorf("transaction not found")
}

// Update updates an existing transaction
func (r *MemoryTransactionRepository) Update(ctx context.Context, transaction *model.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.transactions[transaction.ID()] = transaction
	return nil
}

// Delete removes a transaction
func (r *MemoryTransactionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.transactions, id)
	return nil
}

// List retrieves transactions with filtering
func (r *MemoryTransactionRepository) List(ctx context.Context, filter repository.TransactionFilter) ([]*model.Transaction, error) {
	paginated, err := r.ListWithPagination(ctx, filter)
	if err != nil {
		return nil, err
	}
	return paginated.Transactions, nil
}

// Count returns total number of transactions
func (r *MemoryTransactionRepository) Count(ctx context.Context, filter repository.TransactionFilter) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := int64(0)
	for _, t := range r.transactions {
		if matchesFilter(t, filter) {
			count++
		}
	}

	return count, nil
}

// ListWithPagination retrieves transactions with pagination
func (r *MemoryTransactionRepository) ListWithPagination(ctx context.Context, filter repository.TransactionFilter) (*repository.PaginatedTransactions, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Filter transactions
	var filtered []*model.Transaction
	for _, t := range r.transactions {
		if matchesFilter(t, filter) {
			filtered = append(filtered, t)
		}
	}

	total := int64(len(filtered))

	// Set defaults
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// Apply pagination
	start := filter.Offset
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + filter.Limit
	if end > len(filtered) {
		end = len(filtered)
	}

	paginated := filtered[start:end]
	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit != 0 {
		totalPages++
	}

	return &repository.PaginatedTransactions{
		Transactions: paginated,
		Total:        total,
		Limit:        filter.Limit,
		Offset:       filter.Offset,
		TotalPages:   totalPages,
	}, nil
}

// GetByDateRange retrieves transactions within a date range
func (r *MemoryTransactionRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*model.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*model.Transaction
	for _, t := range r.transactions {
		if t.CreatedAt().After(startDate) && t.CreatedAt().Before(endDate) {
			result = append(result, t)
		}
	}

	return result, nil
}

// GetByCashierID retrieves transactions by cashier
func (r *MemoryTransactionRepository) GetByCashierID(ctx context.Context, cashierID string, limit int) ([]*model.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*model.Transaction
	for _, t := range r.transactions {
		if t.CashierID() == cashierID {
			result = append(result, t)
			if len(result) >= limit {
				break
			}
		}
	}

	return result, nil
}

// GenerateTransactionNo generates a unique transaction number
func (r *MemoryTransactionRepository) GenerateTransactionNo(ctx context.Context) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.transactionNum++
	now := time.Now()
	transactionNo := fmt.Sprintf("TRX-%s-%04d", now.Format("20060102"), r.transactionNum)

	return transactionNo, nil
}

// matchesFilter checks if a transaction matches the filter
func matchesFilter(t *model.Transaction, filter repository.TransactionFilter) bool {
	if filter.Status != "" && t.Status() != filter.Status {
		return false
	}
	if filter.PaymentMethod != "" && t.PaymentMethod() != filter.PaymentMethod {
		return false
	}
	if filter.CashierID != "" && t.CashierID() != filter.CashierID {
		return false
	}
	if !filter.StartDate.IsZero() && t.CreatedAt().Before(filter.StartDate) {
		return false
	}
	if !filter.EndDate.IsZero() && t.CreatedAt().After(filter.EndDate) {
		return false
	}
	if filter.Search != "" {
		if t.TransactionNo() != filter.Search && t.CustomerName() != filter.Search {
			return false
		}
	}
	return true
}

// Ensure interface implementation
var _ repository.TransactionRepository = (*MemoryTransactionRepository)(nil)
