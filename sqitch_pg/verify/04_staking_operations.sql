-- Verify tezos-delegation-service:04_staking_operations to pg

BEGIN;

SELECT id, sender_id, contract_id, entrypoint, amount, block, timestamp, status, created_at
FROM app.staking_operations
WHERE FALSE;

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'staking_operations' AND indexname = 'idx_staking_operations_sender_address';

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'staking_operations' AND indexname = 'idx_staking_operations_contract_address';

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'staking_operations' AND indexname = 'idx_staking_operations_timestamp';

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'staking_operations' AND indexname = 'idx_staking_operations_status';

COMMIT;
