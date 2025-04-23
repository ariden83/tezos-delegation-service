-- Deploy tezos-delegation-service:06_staking_pools to pg
-- requires: 01_appschema 02_accounts 03_delegations 04_staking_operations 05_rewards

BEGIN;

CREATE TABLE IF NOT EXISTS app.staking_pools (
    id SERIAL PRIMARY KEY,
    address TEXT NOT NULL,
    name TEXT NOT NULL,
    staking_token TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_staking_pools_address ON app.staking_pools (address);
CREATE INDEX IF NOT EXISTS idx_staking_pools_token ON app.staking_pools (staking_token);

COMMIT;
