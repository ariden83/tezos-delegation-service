-- Deploy tezos-delegation-service:05_rewards to pg
-- requires: 01_appschema 02_accounts 03_delegations 04_staking_operations

BEGIN;

CREATE TABLE IF NOT EXISTS app.rewards (
    id SERIAL PRIMARY KEY,
    recipient_id BIGINT NOT NULL REFERENCES app.accounts(id),
    source_id BIGINT NOT NULL,
    cycle BIGINT NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    timestamp BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rewards_recipient_id ON app.rewards (recipient_id);
CREATE INDEX IF NOT EXISTS idx_rewards_source_id ON app.rewards (source_id);
CREATE INDEX IF NOT EXISTS idx_rewards_cycle ON app.rewards (cycle);
CREATE INDEX IF NOT EXISTS idx_rewards_timestamp ON app.rewards (timestamp DESC);

COMMIT;
