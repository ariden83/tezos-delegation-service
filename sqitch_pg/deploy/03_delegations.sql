-- Deploy tezos-delegation-service:03_delegations to pg
-- requires: 01_appschema 02_accounts

BEGIN;

CREATE TABLE IF NOT EXISTS app.delegations (
    id SERIAL PRIMARY KEY,
    sender_id BIGINT NOT NULL REFERENCES app.accounts(id),
    delegate_id BIGINT NOT NULL REFERENCES app.accounts(id),
    delegator TEXT NOT NULL,
    delegate TEXT NOT NULL DEFAULT '',
    level BIGINT NOT NULL,
    block TEXT NOT NULL,
    timestamp BIGINT NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_delegations_timestamp ON app.delegations (timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_delegations_delegator ON app.delegations (delegator);
CREATE INDEX IF NOT EXISTS idx_delegations_level ON app.delegations (level);
CREATE INDEX IF NOT EXISTS idx_delegations_sender_id ON app.delegations (sender_id);
CREATE INDEX IF NOT EXISTS idx_delegations_delegate_id ON app.delegations (delegate_id);
CREATE INDEX IF NOT EXISTS idx_delegations_status ON app.delegations (status);
CREATE INDEX IF NOT EXISTS idx_delegations_block ON app.delegations (block);

COMMIT;