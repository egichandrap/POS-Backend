package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
	"github.com/example/jwt-ddd-clean/internal/domain/valueobject"
)

// PostgresUserRepository implements repository.UserRepository for PostgreSQL
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository creates a new PostgresUserRepository
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func scanUser(row *sql.Row) (*model.User, error) {
	var id, username, passwordHash, emailStr, fullName string
	var roleStr, statusStr string
	var createdAt, updatedAt time.Time
	var lastLoginAt sql.NullTime

	err := row.Scan(
		&id, &username, &passwordHash, &emailStr, &fullName,
		&roleStr, &statusStr, &createdAt, &updatedAt, &lastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	email, _ := valueobject.NewEmail(emailStr)
	password := valueobject.Password(passwordHash)
	role := model.UserRole(roleStr)
	status := model.UserStatus(statusStr)

	user := model.ReconstructUser(id, username, email, password, fullName, role, status, createdAt, updatedAt, nil)

	if lastLoginAt.Valid {
		// Record login sets lastLoginAt - we need to set it manually
		// Since we can't access private field, we use a different approach
		// For now, just reconstruct with the lastLoginAt
		user = model.ReconstructUser(id, username, email, password, fullName, role, status, createdAt, updatedAt, &lastLoginAt.Time)
	}

	return user, nil
}

// FindByID retrieves a user by their ID
func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, email, full_name, role, status, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`

	return scanUser(r.db.QueryRowContext(ctx, query, id))
}

// FindByUsername retrieves a user by their username
func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, email, full_name, role, status, created_at, updated_at, last_login_at
		FROM users
		WHERE username = $1
	`

	return scanUser(r.db.QueryRowContext(ctx, query, username))
}

// FindByEmail retrieves a user by their email
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, email, full_name, role, status, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1
	`

	return scanUser(r.db.QueryRowContext(ctx, query, email))
}

// Create creates a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, username, password_hash, email, full_name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID(),
		user.Username(),
		user.PasswordHash().String(),
		user.Email().String(),
		user.FullName(),
		user.Role(),
		user.Status(),
		user.CreatedAt(),
		user.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Update updates an existing user
func (r *PostgresUserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET email = $2, full_name = $3, role = $4, status = $5, updated_at = $6
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID(),
		user.Email().String(),
		user.FullName(),
		user.Role(),
		user.Status(),
		user.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete removes a user
func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// List retrieves users with filtering
func (r *PostgresUserRepository) List(ctx context.Context, filter repository.UserFilter) ([]*model.User, error) {
	paginated, err := r.ListWithPagination(ctx, filter)
	if err != nil {
		return nil, err
	}
	return paginated.Users, nil
}

// Count returns total number of users with optional filter
func (r *PostgresUserRepository) Count(ctx context.Context, filter repository.UserFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if filter.Role != "" {
		query += fmt.Sprintf(" AND role = $%d", argCount)
		args = append(args, string(filter.Role))
		argCount++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, string(filter.Status))
		argCount++
	}

	if filter.Search != "" {
		query += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d OR full_name ILIKE $%d)", argCount, argCount, argCount)
		args = append(args, "%"+filter.Search+"%")
		argCount++
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// ListWithPagination retrieves users with pagination
func (r *PostgresUserRepository) ListWithPagination(ctx context.Context, filter repository.UserFilter) (*repository.PaginatedUsers, error) {
	// Get total count
	total, err := r.Count(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Build query
	query := `
		SELECT id, username, password_hash, email, full_name, role, status, created_at, updated_at, last_login_at
		FROM users
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if filter.Role != "" {
		query += fmt.Sprintf(" AND role = $%d", argCount)
		args = append(args, string(filter.Role))
		argCount++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, string(filter.Status))
		argCount++
	}

	if filter.Search != "" {
		query += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d OR full_name ILIKE $%d)", argCount, argCount, argCount)
		args = append(args, "%"+filter.Search+"%")
		argCount++
	}

	query += " ORDER BY created_at DESC"

	// Set defaults
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, filter.Limit, filter.Offset)

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users := make([]*model.User, 0)
	for rows.Next() {
		user, err := scanRows(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))

	return &repository.PaginatedUsers{
		Users:      users,
		Total:      total,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
		TotalPages: totalPages,
	}, nil
}

func scanRows(rows *sql.Rows) (*model.User, error) {
	var id, username, passwordHash, emailStr, fullName string
	var roleStr, statusStr string
	var createdAt, updatedAt time.Time
	var lastLoginAt sql.NullTime

	err := rows.Scan(
		&id, &username, &passwordHash, &emailStr, &fullName,
		&roleStr, &statusStr, &createdAt, &updatedAt, &lastLoginAt,
	)
	if err != nil {
		return nil, err
	}

	email, _ := valueobject.NewEmail(emailStr)
	password := valueobject.Password(passwordHash)
	role := model.UserRole(roleStr)
	status := model.UserStatus(statusStr)

	user := model.ReconstructUser(id, username, email, password, fullName, role, status, createdAt, updatedAt, nil)

	if lastLoginAt.Valid {
		user = model.ReconstructUser(id, username, email, password, fullName, role, status, createdAt, updatedAt, &lastLoginAt.Time)
	}

	return user, nil
}

// ExistsByUsername checks if a username already exists
func (r *PostgresUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	return exists, nil
}

// ExistsByEmail checks if an email already exists
func (r *PostgresUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return exists, nil
}

// UpdatePassword updates user password
func (r *PostgresUserRepository) UpdatePassword(ctx context.Context, id, hashedPassword string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, hashedPassword, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *PostgresUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	query := `UPDATE users SET last_login_at = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// Ensure interface implementation
var _ repository.UserRepository = (*PostgresUserRepository)(nil)
