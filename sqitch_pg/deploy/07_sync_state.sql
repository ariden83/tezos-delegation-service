-- Deploy tezos-delegation-service:07_sync_state to pg
-- requires: 01_appschema 02_accounts 03_delegations 04_staking_operations 05_rewards 06_staking_pools

BEGIN;

CREATE TABLE IF NOT EXISTS app.sync_state (
    id SERIAL PRIMARY KEY,
    source TEXT NOT NULL UNIQUE,
    last_synced_level BIGINT NOT NULL,
    last_synced_timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMIT;
