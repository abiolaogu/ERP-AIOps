-- Sovereign tenant isolation template for ERP-AIOps
-- Generated 2026-02-28

CREATE SCHEMA IF NOT EXISTS sovereign_core;

CREATE TABLE IF NOT EXISTS sovereign_core.entity_store (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_entity_store_tenant_created
  ON sovereign_core.entity_store (tenant_id, created_at DESC);

ALTER TABLE sovereign_core.entity_store ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS entity_store_tenant_isolation ON sovereign_core.entity_store;
CREATE POLICY entity_store_tenant_isolation
  ON sovereign_core.entity_store
  USING (tenant_id = current_setting('app.tenant_id', true));

-- Ensure session bootstrapping always sets app.tenant_id before data access.
