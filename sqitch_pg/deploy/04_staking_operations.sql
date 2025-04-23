-- Deploy tezos-delegation-service:04_staking_operations to pg
-- requires: 01_appschema 02_accounts 03_delegations

BEGIN;

CREATE TABLE IF NOT EXISTS app.staking_operations (
    id SERIAL PRIMARY KEY,
    sender_id BIGINT NOT NULL REFERENCES app.accounts(id),
    contract_id BIGINT NOT NULL REFERENCES app.accounts(id),
    entrypoint TEXT NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    block TEXT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_staking_operations_sender_id ON app.staking_operations (sender_id);
CREATE INDEX IF NOT EXISTS idx_staking_operations_contract_id ON app.staking_operations (contract_id);
CREATE INDEX IF NOT EXISTS idx_staking_operations_timestamp ON app.staking_operations (timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_staking_operations_status ON app.staking_operations (status);

COMMIT;
