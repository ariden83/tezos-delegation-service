-- Deploy tezos-delegation-service:02_accounts to pg
-- requires: 01_appschema

BEGIN;

CREATE TABLE IF NOT EXISTS app.accounts (
    id SERIAL PRIMARY KEY,
    address TEXT NOT NULL UNIQUE,
    alias TEXT,
    type TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_accounts_address ON app.accounts (address);
CREATE INDEX IF NOT EXISTS idx_accounts_type ON app.accounts (type);

COMMIT;