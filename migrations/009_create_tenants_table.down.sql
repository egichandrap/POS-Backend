-- +migrate Down
-- Drop views
DROP VIEW IF EXISTS product_availability;

-- Drop triggers
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
DROP TRIGGER IF EXISTS update_raw_materials_updated_at ON raw_materials;
DROP TRIGGER IF EXISTS update_product_recipes_updated_at ON product_recipes;
DROP TRIGGER IF EXISTS update_tenant_usage_updated_at ON tenant_usage;
DROP TRIGGER IF EXISTS update_subscription_plans_updated_at ON subscription_plans;
DROP TRIGGER IF EXISTS update_subscription_invoices_updated_at ON subscription_invoices;

-- Drop tables (in correct order due to foreign key constraints)
DROP TABLE IF EXISTS subscription_invoices CASCADE;
DROP TABLE IF EXISTS tenant_audit_logs CASCADE;
DROP TABLE IF EXISTS tenant_usage CASCADE;
DROP TABLE IF EXISTS product_recipes CASCADE;
DROP TABLE IF EXISTS raw_materials CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;
DROP TABLE IF EXISTS subscription_plans CASCADE;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column CASCADE;

-- Remove tenant_id columns from existing tables (keep data but remove constraint)
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_tenant_id_fkey;
ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;

ALTER TABLE stores DROP CONSTRAINT IF EXISTS stores_tenant_id_fkey;
ALTER TABLE stores DROP COLUMN IF EXISTS tenant_id;

ALTER TABLE inventories DROP CONSTRAINT IF EXISTS inventories_tenant_id_fkey;
ALTER TABLE inventories DROP COLUMN IF EXISTS tenant_id;

ALTER TABLE categories DROP CONSTRAINT IF EXISTS categories_tenant_id_fkey;
ALTER TABLE categories DROP COLUMN IF EXISTS tenant_id;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'customers') THEN
        ALTER TABLE customers DROP CONSTRAINT IF EXISTS customers_tenant_id_fkey;
        ALTER TABLE customers DROP COLUMN IF EXISTS tenant_id;
    END IF;
END $$;

ALTER TABLE tables DROP CONSTRAINT IF EXISTS tables_tenant_id_fkey;
ALTER TABLE tables DROP COLUMN IF EXISTS tenant_id;

-- Drop indexes
DROP INDEX IF EXISTS idx_users_tenant_id;
DROP INDEX IF EXISTS idx_stores_tenant_id;
DROP INDEX IF EXISTS idx_inventories_tenant_id;
DROP INDEX IF EXISTS idx_categories_tenant_id;
DROP INDEX IF EXISTS idx_tables_tenant_id;
