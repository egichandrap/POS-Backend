-- +migrate Up
-- ============================================
-- TENANTS & SUBSCRIPTIONS
-- ============================================

CREATE TABLE IF NOT EXISTS subscription_plans (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price_monthly DECIMAL(10,2) NOT NULL DEFAULT 0,
    price_yearly DECIMAL(10,2) NOT NULL DEFAULT 0,
    max_users INTEGER NOT NULL DEFAULT 5,
    max_stores INTEGER NOT NULL DEFAULT 1,
    max_products INTEGER NOT NULL DEFAULT 100,
    max_transactions_per_day INTEGER NOT NULL DEFAULT 1000,
    features JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert default subscription plans
INSERT INTO subscription_plans (id, name, description, price_monthly, price_yearly, max_users, max_stores, max_products, max_transactions_per_day, features) VALUES
(
    'plus',
    'Plus Plan',
    'Perfect for small businesses',
    499000, 4990000,
    5, 1, 100, 500,
    '{
        "pos": true,
        "inventory_management": true,
        "basic_reports": true,
        "multi_user": true,
        "qr_ordering": false,
        "advanced_reports": false,
        "api_access": false,
        "custom_branding": false,
        "multi_store": false,
        "raw_material_management": false
    }'::jsonb
),
(
    'pro',
    'Pro Plan',
    'For growing businesses',
    1499000, 14990000,
    20, 3, 500, 5000,
    '{
        "pos": true,
        "inventory_management": true,
        "basic_reports": true,
        "multi_user": true,
        "qr_ordering": true,
        "advanced_reports": true,
        "api_access": true,
        "custom_branding": true,
        "multi_store": true,
        "raw_material_management": true
    }'::jsonb
),
(
    'enterprise',
    'Enterprise Plan',
    'For large organizations',
    4999000, 49990000,
    -1, -1, -1, -1,
    '{
        "pos": true,
        "inventory_management": true,
        "basic_reports": true,
        "multi_user": true,
        "qr_ordering": true,
        "advanced_reports": true,
        "api_access": true,
        "custom_branding": true,
        "multi_store": true,
        "raw_material_management": true,
        "priority_support": true,
        "custom_integrations": true,
        "dedicated_server": true
    }'::jsonb
) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_name VARCHAR(255) NOT NULL,
    company_slug VARCHAR(100) NOT NULL UNIQUE,
    domain VARCHAR(255) UNIQUE,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    address TEXT,
    logo_url VARCHAR(500),
    subscription_plan_id VARCHAR(50) NOT NULL REFERENCES subscription_plans(id),
    subscription_status VARCHAR(20) NOT NULL DEFAULT 'TRIAL',
    trial_ends_at TIMESTAMP WITH TIME ZONE,
    subscription_starts_at TIMESTAMP WITH TIME ZONE,
    subscription_ends_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id)
);

-- Create indexes for tenants
CREATE INDEX IF NOT EXISTS idx_tenants_company_slug ON tenants(company_slug);
CREATE INDEX IF NOT EXISTS idx_tenants_domain ON tenants(domain);
CREATE INDEX IF NOT EXISTS idx_tenants_subscription_status ON tenants(subscription_status);
CREATE INDEX IF NOT EXISTS idx_tenants_subscription_plan ON tenants(subscription_plan_id);

-- Add constraint for subscription status
ALTER TABLE tenants ADD CONSTRAINT chk_tenants_subscription_status
    CHECK (subscription_status IN ('TRIAL', 'ACTIVE', 'SUSPENDED', 'CANCELLED', 'EXPIRED'));

-- ============================================
-- RAW MATERIALS
-- ============================================

CREATE TABLE IF NOT EXISTS raw_materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    sku VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    unit VARCHAR(50) NOT NULL,
    quantity DECIMAL(10,3) NOT NULL DEFAULT 0,
    min_stock DECIMAL(10,3) NOT NULL DEFAULT 0,
    cost_per_unit DECIMAL(10,2) NOT NULL DEFAULT 0,
    supplier VARCHAR(255),
    location VARCHAR(100),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, sku)
);

-- Create indexes for raw materials
CREATE INDEX IF NOT EXISTS idx_raw_materials_tenant_id ON raw_materials(tenant_id);
CREATE INDEX IF NOT EXISTS idx_raw_materials_sku ON raw_materials(sku);
CREATE INDEX IF NOT EXISTS idx_raw_materials_is_active ON raw_materials(is_active);

-- ============================================
-- PRODUCT RECIPES (RAW MATERIAL RELATIONSHIPS)
-- ============================================

CREATE TABLE IF NOT EXISTS product_recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    inventory_id TEXT NOT NULL REFERENCES inventories(id) ON DELETE CASCADE,
    raw_material_id UUID NOT NULL REFERENCES raw_materials(id) ON DELETE CASCADE,
    quantity_required DECIMAL(10,3) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, inventory_id, raw_material_id)
);

-- Create indexes for product recipes
CREATE INDEX IF NOT EXISTS idx_product_recipes_tenant_id ON product_recipes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_product_recipes_inventory_id ON product_recipes(inventory_id);
CREATE INDEX IF NOT EXISTS idx_product_recipes_raw_material_id ON product_recipes(raw_material_id);
CREATE INDEX IF NOT EXISTS idx_product_recipes_is_active ON product_recipes(is_active);

-- ============================================
-- UPDATE EXISTING TABLES WITH TENANT_ID
-- ============================================

-- Add tenant_id to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Add tenant_id to stores table
ALTER TABLE stores ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Add tenant_id to inventories table
ALTER TABLE inventories ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Add tenant_id to categories table
ALTER TABLE categories ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Add tenant_id to customers table (if exists)
-- Note: customers table may not exist in all installations, so we use a safer approach
-- This will be silently ignored if the table doesn't exist
-- ALTER TABLE customers ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Add tenant_id to tables table
ALTER TABLE tables ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Create indexes for tenant_id in existing tables
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_stores_tenant_id ON stores(tenant_id);
CREATE INDEX IF NOT EXISTS idx_inventories_tenant_id ON inventories(tenant_id);
CREATE INDEX IF NOT EXISTS idx_categories_tenant_id ON categories(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tables_tenant_id ON tables(tenant_id);

-- ============================================
-- USAGE TRACKING FOR SUBSCRIPTION LIMITS
-- ============================================

CREATE TABLE IF NOT EXISTS tenant_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    current_users INTEGER NOT NULL DEFAULT 0,
    current_stores INTEGER NOT NULL DEFAULT 0,
    current_products INTEGER NOT NULL DEFAULT 0,
    transactions_today INTEGER NOT NULL DEFAULT 0,
    last_reset_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- FUNCTIONS AND TRIGGERS
-- ============================================

-- Note: Automatic updated_at triggers require plpgsql functions
-- which are difficult to create with the current migrate library setup
-- The updated_at columns have DEFAULT CURRENT_TIMESTAMP for inserts
-- Application code should handle updates, or we can add triggers in a separate migration

-- Triggers for updated_at will be added in a follow-up migration
-- CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants
--     FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- VIEWS FOR PRODUCT AVAILABILITY
-- ============================================

-- View to check product availability based on raw materials
CREATE OR REPLACE VIEW product_availability AS
SELECT
    i.id AS inventory_id,
    i.tenant_id,
    i.name AS product_name,
    i.sku AS product_sku,
    i.quantity AS product_quantity,
    COUNT(pr.id) AS total_ingredients,
    SUM(CASE WHEN rm.quantity >= pr.quantity_required THEN 1 ELSE 0 END) AS available_ingredients,
    CASE
        WHEN COUNT(pr.id) = 0 THEN TRUE
        WHEN SUM(CASE WHEN rm.quantity >= pr.quantity_required THEN 1 ELSE 0 END) = COUNT(pr.id) THEN TRUE
        ELSE FALSE
    END AS is_available,
    jsonb_agg(
        jsonb_build_object(
            'material_id', rm.id,
            'material_name', rm.name,
            'material_sku', rm.sku,
            'required', pr.quantity_required,
            'available', rm.quantity,
            'unit', rm.unit,
            'is_sufficient', rm.quantity >= pr.quantity_required
        )
    ) FILTER (WHERE pr.id IS NOT NULL) AS materials_status
FROM inventories i
LEFT JOIN product_recipes pr ON i.id = pr.inventory_id AND pr.is_active = TRUE
LEFT JOIN raw_materials rm ON pr.raw_material_id = rm.id AND rm.is_active = TRUE
GROUP BY i.id, i.tenant_id, i.name, i.sku, i.quantity;

-- ============================================
-- AUDIT LOGS FOR TENANT OPERATIONS
-- ============================================

CREATE TABLE IF NOT EXISTS tenant_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id VARCHAR(255),
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tenant_audit_logs_tenant_id ON tenant_audit_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_audit_logs_user_id ON tenant_audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_audit_logs_created_at ON tenant_audit_logs(created_at);

-- ============================================
-- SUBSCRIPTION INVOICES
-- ============================================

CREATE TABLE IF NOT EXISTS subscription_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    invoice_number VARCHAR(50) NOT NULL UNIQUE,
    subscription_plan_id VARCHAR(50) NOT NULL REFERENCES subscription_plans(id),
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    due_date DATE,
    paid_at TIMESTAMP WITH TIME ZONE,
    payment_method VARCHAR(50),
    payment_reference VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_subscription_invoices_tenant_id ON subscription_invoices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_subscription_invoices_status ON subscription_invoices(status);
CREATE INDEX IF NOT EXISTS idx_subscription_invoices_invoice_number ON subscription_invoices(invoice_number);

ALTER TABLE subscription_invoices ADD CONSTRAINT chk_invoice_status
    CHECK (status IN ('PENDING', 'PAID', 'OVERDUE', 'CANCELLED'));

-- Note: Trigger commented out due to migrate library limitations
-- CREATE TRIGGER update_subscription_invoices_updated_at BEFORE UPDATE ON subscription_invoices
--     FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
