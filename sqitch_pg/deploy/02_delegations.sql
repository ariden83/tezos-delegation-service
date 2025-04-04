-- Deploy tezos-delegation-service:02_delegations to pg
-- requires: 01_appschema

BEGIN;

CREATE TABLE IF NOT EXISTS app.delegations (
    id SERIAL PRIMARY KEY,
    delegator TEXT NOT NULL,
    delegate TEXT NOT NULL DEFAULT '',
    timestamp BIGINT NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    level BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_delegations_timestamp ON app.delegations (timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_delegations_delegator ON app.delegations (delegator);
CREATE INDEX IF NOT EXISTS idx_delegations_level ON app.delegations (level);

COMMIT;
