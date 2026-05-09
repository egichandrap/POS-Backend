package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusTrial     SubscriptionStatus = "TRIAL"
	SubscriptionStatusActive    SubscriptionStatus = "ACTIVE"
	SubscriptionStatusSuspended SubscriptionStatus = "SUSPENDED"
	SubscriptionStatusCancelled SubscriptionStatus = "CANCELLED"
	SubscriptionStatusExpired   SubscriptionStatus = "EXPIRED"
)

// SubscriptionPlan represents subscription plan types
type SubscriptionPlan string

const (
	PlanPlus       SubscriptionPlan = "plus"
	PlanPro        SubscriptionPlan = "pro"
	PlanEnterprise SubscriptionPlan = "enterprise"
)

// SubscriptionFeatures represents available features based on subscription plan
type SubscriptionFeatures struct {
	POS                     bool `json:"pos"`
	InventoryManagement     bool `json:"inventory_management"`
	BasicReports            bool `json:"basic_reports"`
	MultiUser               bool `json:"multi_user"`
	QROrdering              bool `json:"qr_ordering"`
	AdvancedReports         bool `json:"advanced_reports"`
	APIAccess               bool `json:"api_access"`
	CustomBranding          bool `json:"custom_branding"`
	MultiStore              bool `json:"multi_store"`
	RawMaterialManagement   bool `json:"raw_material_management"`
	PrioritySupport         bool `json:"priority_support"`
	CustomIntegrations      bool `json:"custom_integrations"`
	DedicatedServer         bool `json:"dedicated_server"`
}

// Tenant represents a company/organization entity
type Tenant struct {
	id                    string
	companyName           string
	companySlug           string
	domain                string
	email                 string
	phone                 string
	address               string
	logoURL               string
	subscriptionPlanID    string
	subscriptionStatus    SubscriptionStatus
	trialEndsAt           *time.Time
	subscriptionStartsAt  *time.Time
	subscriptionEndsAt    *time.Time
	isActive              bool
	settings              map[string]interface{}
	createdAt             time.Time
	updatedAt             time.Time
	createdBy             string
}

// NewTenant creates a new tenant entity with validation
func NewTenant(
	companyName string,
	companySlug string,
	email string,
	subscriptionPlanID string,
	createdBy string,
) (*Tenant, error) {
	if companyName == "" {
		return nil, fmt.Errorf("nama perusahaan tidak boleh kosong")
	}
	if companySlug == "" {
		return nil, fmt.Errorf("slug perusahaan tidak boleh kosong")
	}
	if email == "" {
		return nil, fmt.Errorf("email tidak boleh kosong")
	}
	if subscriptionPlanID == "" {
		subscriptionPlanID = string(PlanPlus) // Default plan
	}

	now := time.Now()
	trialEndsAt := now.AddDate(0, 0, 14) // 14 days trial

	return &Tenant{
		id:                 uuid.New().String(),
		companyName:        companyName,
		companySlug:        companySlug,
		email:              email,
		subscriptionPlanID: subscriptionPlanID,
		subscriptionStatus: SubscriptionStatusTrial,
		trialEndsAt:        &trialEndsAt,
		isActive:           true,
		settings:           make(map[string]interface{}),
		createdAt:          now,
		updatedAt:          now,
		createdBy:          createdBy,
	}, nil
}

// ReconstructTenant recreates a tenant entity from database (trusted data)
func ReconstructTenant(
	id, companyName, companySlug, domain, email, phone, address, logoURL string,
	subscriptionPlanID string,
	subscriptionStatus SubscriptionStatus,
	trialEndsAt, subscriptionStartsAt, subscriptionEndsAt *time.Time,
	isActive bool,
	settingsJSON string,
	createdAt, updatedAt time.Time,
	createdBy string,
) *Tenant {
	settings := make(map[string]interface{})
	if settingsJSON != "" {
		json.Unmarshal([]byte(settingsJSON), &settings)
	}

	return &Tenant{
		id:                    id,
		companyName:           companyName,
		companySlug:           companySlug,
		domain:                domain,
		email:                 email,
		phone:                 phone,
		address:               address,
		logoURL:               logoURL,
		subscriptionPlanID:    subscriptionPlanID,
		subscriptionStatus:    subscriptionStatus,
		trialEndsAt:           trialEndsAt,
		subscriptionStartsAt:  subscriptionStartsAt,
		subscriptionEndsAt:    subscriptionEndsAt,
		isActive:              isActive,
		settings:              settings,
		createdAt:             createdAt,
		updatedAt:             updatedAt,
		createdBy:             createdBy,
	}
}

// Accessor methods (read-only)

func (t *Tenant) ID() string                       { return t.id }
func (t *Tenant) CompanyName() string              { return t.companyName }
func (t *Tenant) CompanySlug() string              { return t.companySlug }
func (t *Tenant) Domain() string                   { return t.domain }
func (t *Tenant) Email() string                    { return t.email }
func (t *Tenant) Phone() string                    { return t.phone }
func (t *Tenant) Address() string                  { return t.address }
func (t *Tenant) LogoURL() string                  { return t.logoURL }
func (t *Tenant) SubscriptionPlanID() string       { return t.subscriptionPlanID }
func (t *Tenant) SubscriptionStatus() SubscriptionStatus { return t.subscriptionStatus }
func (t *Tenant) TrialEndsAt() *time.Time          { return t.trialEndsAt }
func (t *Tenant) SubscriptionStartsAt() *time.Time { return t.subscriptionStartsAt }
func (t *Tenant) SubscriptionEndsAt() *time.Time   { return t.subscriptionEndsAt }
func (t *Tenant) IsActive() bool                   { return t.isActive }
func (t *Tenant) Settings() map[string]interface{} { return t.settings }
func (t *Tenant) CreatedAt() time.Time             { return t.createdAt }
func (t *Tenant) UpdatedAt() time.Time             { return t.updatedAt }
func (t *Tenant) CreatedBy() string                { return t.createdBy }

// Ubiquitous language methods for tenant operations

// UpdateProfile updates tenant profile information
func (t *Tenant) UpdateProfile(companyName, email, phone, address, logoURL string) error {
	if companyName == "" {
		return fmt.Errorf("nama perusahaan tidak boleh kosong")
	}
	if email == "" {
		return fmt.Errorf("email tidak boleh kosong")
	}

	t.companyName = companyName
	t.email = email
	t.phone = phone
	t.address = address
	t.logoURL = logoURL
	t.updatedAt = time.Now()
	return nil
}

// UpdateDomain updates tenant custom domain
func (t *Tenant) UpdateDomain(domain string) error {
	t.domain = domain
	t.updatedAt = time.Now()
	return nil
}

// ActivateSubscription activates a subscription
func (t *Tenant) ActivateSubscription(planID string, startsAt, endsAt time.Time) error {
	t.subscriptionPlanID = planID
	t.subscriptionStatus = SubscriptionStatusActive
	t.subscriptionStartsAt = &startsAt
	t.subscriptionEndsAt = &endsAt
	t.updatedAt = time.Now()
	return nil
}

// SuspendSubscription suspends the subscription
func (t *Tenant) SuspendSubscription() error {
	t.subscriptionStatus = SubscriptionStatusSuspended
	t.updatedAt = time.Now()
	return nil
}

// CancelSubscription cancels the subscription
func (t *Tenant) CancelSubscription() error {
	t.subscriptionStatus = SubscriptionStatusCancelled
	t.updatedAt = time.Now()
	return nil
}

// ExpireSubscription marks subscription as expired
func (t *Tenant) ExpireSubscription() error {
	t.subscriptionStatus = SubscriptionStatusExpired
	t.updatedAt = time.Now()
	return nil
}

// Activate activates the tenant account
func (t *Tenant) Activate() {
	t.isActive = true
	t.updatedAt = time.Now()
}

// Deactivate deactivates the tenant account
func (t *Tenant) Deactivate() {
	t.isActive = false
	t.updatedAt = time.Now()
}

// UpdateSettings updates tenant settings
func (t *Tenant) UpdateSettings(settings map[string]interface{}) error {
	if t.settings == nil {
		t.settings = make(map[string]interface{})
	}
	for key, value := range settings {
		t.settings[key] = value
	}
	t.updatedAt = time.Now()
	return nil
}

// GetSetting retrieves a specific setting value
func (t *Tenant) GetSetting(key string) (interface{}, bool) {
	if t.settings == nil {
		return nil, false
	}
	value, exists := t.settings[key]
	return value, exists
}

// IsTrial checks if tenant is in trial period
func (t *Tenant) IsTrial() bool {
	if t.subscriptionStatus != SubscriptionStatusTrial {
		return false
	}
	if t.trialEndsAt == nil {
		return false
	}
	return time.Now().Before(*t.trialEndsAt)
}

// IsSubscriptionActive checks if subscription is active and not expired
func (t *Tenant) IsSubscriptionActive() bool {
	if t.subscriptionStatus != SubscriptionStatusActive {
		return false
	}
	if t.subscriptionEndsAt == nil {
		return true
	}
	return time.Now().Before(*t.subscriptionEndsAt)
}

// HasFeatureAccess checks if tenant has access to a specific feature
// This should be used with subscription plan features from repository
func (t *Tenant) HasFeatureAccess(features SubscriptionFeatures, feature string) bool {
	switch feature {
	case "pos":
		return features.POS
	case "inventory_management":
		return features.InventoryManagement
	case "basic_reports":
		return features.BasicReports
	case "multi_user":
		return features.MultiUser
	case "qr_ordering":
		return features.QROrdering
	case "advanced_reports":
		return features.AdvancedReports
	case "api_access":
		return features.APIAccess
	case "custom_branding":
		return features.CustomBranding
	case "multi_store":
		return features.MultiStore
	case "raw_material_management":
		return features.RawMaterialManagement
	case "priority_support":
		return features.PrioritySupport
	case "custom_integrations":
		return features.CustomIntegrations
	case "dedicated_server":
		return features.DedicatedServer
	default:
		return false
	}
}

// SubscriptionPlanDetail represents a subscription plan entity
type SubscriptionPlanDetail struct {
	id                       string
	name                     string
	description              string
	priceMonthly             float64
	priceYearly              float64
	maxUsers                 int
	maxStores                int
	maxProducts              int
	maxTransactionsPerDay    int
	features                 SubscriptionFeatures
	isActive                 bool
	createdAt                time.Time
	updatedAt                time.Time
}

// ReconstructSubscriptionPlanDetail recreates a subscription plan from database (trusted data)
func ReconstructSubscriptionPlanDetail(
	id, name, description string,
	priceMonthly, priceYearly float64,
	maxUsers, maxStores, maxProducts, maxTransactionsPerDay int,
	featuresJSON string,
	isActive bool,
	createdAt, updatedAt time.Time,
) *SubscriptionPlanDetail {
	var features SubscriptionFeatures
	if featuresJSON != "" {
		json.Unmarshal([]byte(featuresJSON), &features)
	}

	return &SubscriptionPlanDetail{
		id:                    id,
		name:                  name,
		description:           description,
		priceMonthly:          priceMonthly,
		priceYearly:           priceYearly,
		maxUsers:              maxUsers,
		maxStores:             maxStores,
		maxProducts:           maxProducts,
		maxTransactionsPerDay: maxTransactionsPerDay,
		features:              features,
		isActive:              isActive,
		createdAt:             createdAt,
		updatedAt:             updatedAt,
	}
}

// Accessor methods for SubscriptionPlanDetail
func (p *SubscriptionPlanDetail) ID() string                 { return p.id }
func (p *SubscriptionPlanDetail) Name() string               { return p.name }
func (p *SubscriptionPlanDetail) Description() string        { return p.description }
func (p *SubscriptionPlanDetail) PriceMonthly() float64      { return p.priceMonthly }
func (p *SubscriptionPlanDetail) PriceYearly() float64       { return p.priceYearly }
func (p *SubscriptionPlanDetail) MaxUsers() int              { return p.maxUsers }
func (p *SubscriptionPlanDetail) MaxStores() int             { return p.maxStores }
func (p *SubscriptionPlanDetail) MaxProducts() int           { return p.maxProducts }
func (p *SubscriptionPlanDetail) MaxTransactionsPerDay() int { return p.maxTransactionsPerDay }
func (p *SubscriptionPlanDetail) Features() SubscriptionFeatures { return p.features }
func (p *SubscriptionPlanDetail) IsActive() bool             { return p.isActive }
func (p *SubscriptionPlanDetail) CreatedAt() time.Time       { return p.createdAt }
func (p *SubscriptionPlanDetail) UpdatedAt() time.Time       { return p.updatedAt }

// HasUnlimitedUsers checks if plan has unlimited users
func (p *SubscriptionPlanDetail) HasUnlimitedUsers() bool {
	return p.maxUsers < 0
}

// HasUnlimitedStores checks if plan has unlimited stores
func (p *SubscriptionPlanDetail) HasUnlimitedStores() bool {
	return p.maxStores < 0
}

// HasUnlimitedProducts checks if plan has unlimited products
func (p *SubscriptionPlanDetail) HasUnlimitedProducts() bool {
	return p.maxProducts < 0
}

// HasUnlimitedTransactions checks if plan has unlimited transactions
func (p *SubscriptionPlanDetail) HasUnlimitedTransactions() bool {
	return p.maxTransactionsPerDay < 0
}
