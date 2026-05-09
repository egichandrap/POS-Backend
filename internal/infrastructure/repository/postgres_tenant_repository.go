package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/example/jwt-ddd-clean/internal/domain/model"
	"github.com/example/jwt-ddd-clean/internal/domain/repository"
	"github.com/google/uuid"
)

type PostgresTenantRepository struct {
	db *sql.DB
}

func NewPostgresTenantRepository(db *sql.DB) repository.TenantRepository {
	return &PostgresTenantRepository{db: db}
}

func (r *PostgresTenantRepository) Create(ctx context.Context, tenant *model.Tenant) error {
	settingsJSON, err := json.Marshal(tenant.Settings())
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Handle empty domain as NULL (unique constraint allows multiple NULLs)
	var domain interface{} = nil
	if tenant.Domain() != "" {
		domain = tenant.Domain()
	}

	// Handle empty createdBy as NULL for UUID column
	var createdBy interface{} = nil
	if tenant.CreatedBy() != "" {
		createdBy = tenant.CreatedBy()
	}

	query := `
		INSERT INTO tenants (
			id, company_name, company_slug, domain, email, phone, address,
			logo_url, subscription_plan_id, subscription_status,
			trial_ends_at, subscription_starts_at, subscription_ends_at,
			is_active, settings, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err = r.db.ExecContext(ctx, query,
		tenant.ID(), tenant.CompanyName(), tenant.CompanySlug(), domain,
		tenant.Email(), tenant.Phone(), tenant.Address(), tenant.LogoURL(),
		tenant.SubscriptionPlanID(), tenant.SubscriptionStatus(),
		tenant.TrialEndsAt(), tenant.SubscriptionStartsAt(), tenant.SubscriptionEndsAt(),
		tenant.IsActive(), settingsJSON, createdBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create initial usage record
	usageQuery := `
		INSERT INTO tenant_usage (id, tenant_id, current_users, current_stores, current_products, transactions_today, last_reset_date)
		VALUES ($1, $2, 0, 0, 0, 0, CURRENT_DATE)
	`
	_, err = r.db.ExecContext(ctx, usageQuery, uuid.New().String(), tenant.ID())
	if err != nil {
		return fmt.Errorf("failed to create tenant usage: %w", err)
	}

	return nil
}

func (r *PostgresTenantRepository) GetByID(ctx context.Context, id string) (*model.Tenant, error) {
	query := `
		SELECT id, company_name, company_slug, domain, email, phone, address,
			logo_url, subscription_plan_id, subscription_status,
			trial_ends_at, subscription_starts_at, subscription_ends_at,
			is_active, settings, created_at, updated_at, created_by
		FROM tenants
		WHERE id = $1
	`

	var tenantID, companyName, companySlug, email, phone, address string
	var domain, logoURL sql.NullString
	var subscriptionPlanID string
	var subscriptionStatus string
	var trialEndsAt, subscriptionStartsAt, subscriptionEndsAt sql.NullTime
	var isActive bool
	var settingsJSON string
	var createdAt, updatedAt time.Time
	var createdBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tenantID, &companyName, &companySlug, &domain, &email, &phone,
		&address, &logoURL, &subscriptionPlanID, &subscriptionStatus,
		&trialEndsAt, &subscriptionStartsAt, &subscriptionEndsAt,
		&isActive, &settingsJSON, &createdAt, &updatedAt, &createdBy,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Convert sql.NullTime to *time.Time
	var trialEndsPtr, subscriptionStartsPtr, subscriptionEndsPtr *time.Time
	if trialEndsAt.Valid {
		trialEndsPtr = &trialEndsAt.Time
	}
	if subscriptionStartsAt.Valid {
		subscriptionStartsPtr = &subscriptionStartsAt.Time
	}
	if subscriptionEndsAt.Valid {
		subscriptionEndsPtr = &subscriptionEndsAt.Time
	}

	return model.ReconstructTenant(
		tenantID, companyName, companySlug, domain.String, email, phone,
		address, logoURL.String, subscriptionPlanID, model.SubscriptionStatus(subscriptionStatus),
		trialEndsPtr, subscriptionStartsPtr, subscriptionEndsPtr,
		isActive, settingsJSON, createdAt, updatedAt, createdBy.String,
	), nil
}

func (r *PostgresTenantRepository) GetByCompanySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	query := `
		SELECT id, company_name, company_slug, domain, email, phone, address,
			logo_url, subscription_plan_id, subscription_status,
			trial_ends_at, subscription_starts_at, subscription_ends_at,
			is_active, settings, created_at, updated_at, created_by
		FROM tenants
		WHERE company_slug = $1
	`

	var tenantID, companyName, companySlug, email, phone, address string
	var domain, logoURL sql.NullString
	var subscriptionPlanID string
	var subscriptionStatus string
	var trialEndsAt, subscriptionStartsAt, subscriptionEndsAt sql.NullTime
	var isActive bool
	var settingsJSON string
	var createdAt, updatedAt time.Time
	var createdBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&tenantID, &companyName, &companySlug, &domain, &email, &phone,
		&address, &logoURL, &subscriptionPlanID, &subscriptionStatus,
		&trialEndsAt, &subscriptionStartsAt, &subscriptionEndsAt,
		&isActive, &settingsJSON, &createdAt, &updatedAt, &createdBy,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Convert sql.NullTime to *time.Time
	var trialEndsPtr, subscriptionStartsPtr, subscriptionEndsPtr *time.Time
	if trialEndsAt.Valid {
		trialEndsPtr = &trialEndsAt.Time
	}
	if subscriptionStartsAt.Valid {
		subscriptionStartsPtr = &subscriptionStartsAt.Time
	}
	if subscriptionEndsAt.Valid {
		subscriptionEndsPtr = &subscriptionEndsAt.Time
	}

	return model.ReconstructTenant(
		tenantID, companyName, companySlug, domain.String, email, phone,
		address, logoURL.String, subscriptionPlanID, model.SubscriptionStatus(subscriptionStatus),
		trialEndsPtr, subscriptionStartsPtr, subscriptionEndsPtr,
		isActive, settingsJSON, createdAt, updatedAt, createdBy.String,
	), nil
}

func (r *PostgresTenantRepository) GetByDomain(ctx context.Context, domainParam string) (*model.Tenant, error) {
	query := `
		SELECT id, company_name, company_slug, domain, email, phone, address,
			logo_url, subscription_plan_id, subscription_status,
			trial_ends_at, subscription_starts_at, subscription_ends_at,
			is_active, settings, created_at, updated_at, created_by
		FROM tenants
		WHERE domain = $1
	`

	var tenantID, companyName, companySlug, email, phone, address string
	var domainVal, logoURL sql.NullString
	var subscriptionPlanID string
	var subscriptionStatus string
	var trialEndsAt, subscriptionStartsAt, subscriptionEndsAt sql.NullTime
	var isActive bool
	var settingsJSON string
	var createdAt, updatedAt time.Time
	var createdBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, domainParam).Scan(
		&tenantID, &companyName, &companySlug, &domainVal, &email, &phone,
		&address, &logoURL, &subscriptionPlanID, &subscriptionStatus,
		&trialEndsAt, &subscriptionStartsAt, &subscriptionEndsAt,
		&isActive, &settingsJSON, &createdAt, &updatedAt, &createdBy,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Convert sql.NullTime to *time.Time
	var trialEndsPtr, subscriptionStartsPtr, subscriptionEndsPtr *time.Time
	if trialEndsAt.Valid {
		trialEndsPtr = &trialEndsAt.Time
	}
	if subscriptionStartsAt.Valid {
		subscriptionStartsPtr = &subscriptionStartsAt.Time
	}
	if subscriptionEndsAt.Valid {
		subscriptionEndsPtr = &subscriptionEndsAt.Time
	}

	return model.ReconstructTenant(
		tenantID, companyName, companySlug, domainVal.String, email, phone,
		address, logoURL.String, subscriptionPlanID, model.SubscriptionStatus(subscriptionStatus),
		trialEndsPtr, subscriptionStartsPtr, subscriptionEndsPtr,
		isActive, settingsJSON, createdAt, updatedAt, createdBy.String,
	), nil
}

func (r *PostgresTenantRepository) GetByUserID(ctx context.Context, userID string) (*model.Tenant, error) {
	query := `
		SELECT t.id, t.company_name, t.company_slug, t.domain, t.email, t.phone, t.address,
			t.logo_url, t.subscription_plan_id, t.subscription_status,
			t.trial_ends_at, t.subscription_starts_at, t.subscription_ends_at,
			t.is_active, t.settings, t.created_at, t.updated_at, t.created_by
		FROM tenants t
		INNER JOIN users u ON u.tenant_id = t.id
		WHERE u.id = $1
	`

	var tenantID, companyName, companySlug, email, phone, address string
	var domain, logoURL sql.NullString
	var subscriptionPlanID string
	var subscriptionStatus string
	var trialEndsAt, subscriptionStartsAt, subscriptionEndsAt sql.NullTime
	var isActive bool
	var settingsJSON string
	var createdAt, updatedAt time.Time
	var createdBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&tenantID, &companyName, &companySlug, &domain, &email, &phone,
		&address, &logoURL, &subscriptionPlanID, &subscriptionStatus,
		&trialEndsAt, &subscriptionStartsAt, &subscriptionEndsAt,
		&isActive, &settingsJSON, &createdAt, &updatedAt, &createdBy,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found for user")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by user: %w", err)
	}

	// Convert sql.NullTime to *time.Time
	var trialEndsPtr, subscriptionStartsPtr, subscriptionEndsPtr *time.Time
	if trialEndsAt.Valid {
		trialEndsPtr = &trialEndsAt.Time
	}
	if subscriptionStartsAt.Valid {
		subscriptionStartsPtr = &subscriptionStartsAt.Time
	}
	if subscriptionEndsAt.Valid {
		subscriptionEndsPtr = &subscriptionEndsAt.Time
	}

	return model.ReconstructTenant(
		tenantID, companyName, companySlug, domain.String, email, phone,
		address, logoURL.String, subscriptionPlanID, model.SubscriptionStatus(subscriptionStatus),
		trialEndsPtr, subscriptionStartsPtr, subscriptionEndsPtr,
		isActive, settingsJSON, createdAt, updatedAt, createdBy.String,
	), nil
}

func (r *PostgresTenantRepository) List(ctx context.Context, filter repository.TenantFilter) ([]*model.Tenant, error) {
	query := `
		SELECT id, company_name, company_slug, domain, email, phone, address,
			logo_url, subscription_plan_id, subscription_status,
			trial_ends_at, subscription_starts_at, subscription_ends_at,
			is_active, settings, created_at, updated_at, created_by
		FROM tenants
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if filter.SubscriptionStatus != "" {
		query += fmt.Sprintf(" AND subscription_status = $%d", argCount)
		args = append(args, filter.SubscriptionStatus)
		argCount++
	}

	if filter.SubscriptionPlanID != "" {
		query += fmt.Sprintf(" AND subscription_plan_id = $%d", argCount)
		args = append(args, filter.SubscriptionPlanID)
		argCount++
	}

	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argCount)
		args = append(args, *filter.IsActive)
		argCount++
	}

	if filter.Search != "" {
		query += fmt.Sprintf(" AND (company_name ILIKE $%d OR email ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+filter.Search+"%")
		argCount++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*model.Tenant
	for rows.Next() {
		var tenantID, companyName, companySlug, email, phone, address string
		var domain, logoURL sql.NullString
		var subscriptionPlanID string
		var subscriptionStatus string
		var trialEndsAt, subscriptionStartsAt, subscriptionEndsAt sql.NullTime
		var isActive bool
		var settingsJSON string
		var createdAt, updatedAt time.Time
		var createdBy sql.NullString

		err := rows.Scan(
			&tenantID, &companyName, &companySlug, &domain, &email, &phone,
			&address, &logoURL, &subscriptionPlanID, &subscriptionStatus,
			&trialEndsAt, &subscriptionStartsAt, &subscriptionEndsAt,
			&isActive, &settingsJSON, &createdAt, &updatedAt, &createdBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}

		// Convert sql.NullTime to *time.Time
		var trialEndsPtr, subscriptionStartsPtr, subscriptionEndsPtr *time.Time
		if trialEndsAt.Valid {
			trialEndsPtr = &trialEndsAt.Time
		}
		if subscriptionStartsAt.Valid {
			subscriptionStartsPtr = &subscriptionStartsAt.Time
		}
		if subscriptionEndsAt.Valid {
			subscriptionEndsPtr = &subscriptionEndsAt.Time
		}

		tenants = append(tenants, model.ReconstructTenant(
			tenantID, companyName, companySlug, domain.String, email, phone,
			address, logoURL.String, subscriptionPlanID, model.SubscriptionStatus(subscriptionStatus),
			trialEndsPtr, subscriptionStartsPtr, subscriptionEndsPtr,
			isActive, settingsJSON, createdAt, updatedAt, createdBy.String,
		))
	}

	return tenants, nil
}

func (r *PostgresTenantRepository) Update(ctx context.Context, tenant *model.Tenant) error {
	settingsJSON, err := json.Marshal(tenant.Settings())
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		UPDATE tenants SET
			company_name = $2,
			company_slug = $3,
			domain = $4,
			email = $5,
			phone = $6,
			address = $7,
			logo_url = $8,
			subscription_plan_id = $9,
			subscription_status = $10,
			trial_ends_at = $11,
			subscription_starts_at = $12,
			subscription_ends_at = $13,
			is_active = $14,
			settings = $15
		WHERE id = $1
	`

	_, err = r.db.ExecContext(ctx, query,
		tenant.ID(), tenant.CompanyName(), tenant.CompanySlug(), tenant.Domain(),
		tenant.Email(), tenant.Phone(), tenant.Address(), tenant.LogoURL(),
		tenant.SubscriptionPlanID(), tenant.SubscriptionStatus(),
		tenant.TrialEndsAt(), tenant.SubscriptionStartsAt(), tenant.SubscriptionEndsAt(),
		tenant.IsActive(), settingsJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	return nil
}

func (r *PostgresTenantRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tenants WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	return nil
}

func (r *PostgresTenantRepository) ExistsByCompanySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tenants WHERE company_slug = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check slug existence: %w", err)
	}
	return exists, nil
}

func (r *PostgresTenantRepository) ExistsByDomain(ctx context.Context, domain string) (bool, error) {
	if domain == "" {
		return false, nil
	}
	query := `SELECT EXISTS(SELECT 1 FROM tenants WHERE domain = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, domain).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check domain existence: %w", err)
	}
	return exists, nil
}

func (r *PostgresTenantRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM tenants`
	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tenants: %w", err)
	}
	return count, nil
}

// PostgresSubscriptionPlanRepository implements subscription plan repository
type PostgresSubscriptionPlanRepository struct {
	db *sql.DB
}

func NewPostgresSubscriptionPlanRepository(db *sql.DB) repository.SubscriptionPlanRepository {
	return &PostgresSubscriptionPlanRepository{db: db}
}

func (r *PostgresSubscriptionPlanRepository) GetByID(ctx context.Context, id string) (*model.SubscriptionPlanDetail, error) {
	query := `
		SELECT id, name, description, price_monthly, price_yearly,
			max_users, max_stores, max_products, max_transactions_per_day,
			features, is_active, created_at, updated_at
		FROM subscription_plans
		WHERE id = $1
	`

	var planID, name, description string
	var priceMonthly, priceYearly float64
	var maxUsers, maxStores, maxProducts, maxTransactionsPerDay int
	var featuresJSON string
	var isActive bool
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&planID, &name, &description, &priceMonthly, &priceYearly,
		&maxUsers, &maxStores, &maxProducts, &maxTransactionsPerDay,
		&featuresJSON, &isActive, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("subscription plan not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	return model.ReconstructSubscriptionPlanDetail(
		planID, name, description, priceMonthly, priceYearly,
		maxUsers, maxStores, maxProducts, maxTransactionsPerDay,
		featuresJSON, isActive, createdAt, updatedAt,
	), nil
}

func (r *PostgresSubscriptionPlanRepository) List(ctx context.Context) ([]*model.SubscriptionPlanDetail, error) {
	query := `
		SELECT id, name, description, price_monthly, price_yearly,
			max_users, max_stores, max_products, max_transactions_per_day,
			features, is_active, created_at, updated_at
		FROM subscription_plans
		ORDER BY price_monthly ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscription plans: %w", err)
	}
	defer rows.Close()

	var plans []*model.SubscriptionPlanDetail
	for rows.Next() {
		var planID, name, description string
		var priceMonthly, priceYearly float64
		var maxUsers, maxStores, maxProducts, maxTransactionsPerDay int
		var featuresJSON string
		var isActive bool
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&planID, &name, &description, &priceMonthly, &priceYearly,
			&maxUsers, &maxStores, &maxProducts, &maxTransactionsPerDay,
			&featuresJSON, &isActive, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription plan: %w", err)
		}

		plans = append(plans, model.ReconstructSubscriptionPlanDetail(
			planID, name, description, priceMonthly, priceYearly,
			maxUsers, maxStores, maxProducts, maxTransactionsPerDay,
			featuresJSON, isActive, createdAt, updatedAt,
		))
	}

	return plans, nil
}

func (r *PostgresSubscriptionPlanRepository) ListActive(ctx context.Context) ([]*model.SubscriptionPlanDetail, error) {
	query := `
		SELECT id, name, description, price_monthly, price_yearly,
			max_users, max_stores, max_products, max_transactions_per_day,
			features, is_active, created_at, updated_at
		FROM subscription_plans
		WHERE is_active = true
		ORDER BY price_monthly ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active subscription plans: %w", err)
	}
	defer rows.Close()

	var plans []*model.SubscriptionPlanDetail
	for rows.Next() {
		var planID, name, description string
		var priceMonthly, priceYearly float64
		var maxUsers, maxStores, maxProducts, maxTransactionsPerDay int
		var featuresJSON string
		var isActive bool
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&planID, &name, &description, &priceMonthly, &priceYearly,
			&maxUsers, &maxStores, &maxProducts, &maxTransactionsPerDay,
			&featuresJSON, &isActive, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription plan: %w", err)
		}

		plans = append(plans, model.ReconstructSubscriptionPlanDetail(
			planID, name, description, priceMonthly, priceYearly,
			maxUsers, maxStores, maxProducts, maxTransactionsPerDay,
			featuresJSON, isActive, createdAt, updatedAt,
		))
	}

	return plans, nil
}
