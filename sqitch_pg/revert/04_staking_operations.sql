-- Revert tezos-delegation-service:04_staking_operations to pg

BEGIN;

DROP INDEX IF EXISTS app.idx_staking_operations_sender_id;
DROP INDEX IF EXISTS app.idx_staking_operations_contract_id;
DROP INDEX IF EXISTS app.idx_staking_operations_timestamp;
DROP INDEX IF EXISTS app.idx_staking_operations_status;

DROP TABLE IF EXISTS app.staking_operations;

COMMIT;
