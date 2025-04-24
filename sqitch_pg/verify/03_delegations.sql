-- Verify tezos-delegation-service:03_delegations to pg

BEGIN;

SELECT id, delegator, timestamp, amount, level, created_at
FROM app.delegations
WHERE FALSE;

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'delegations' AND indexname = 'idx_delegations_timestamp';

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'delegations' AND indexname = 'idx_delegations_delegator';

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'delegations' AND indexname = 'idx_delegations_level';

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'delegations' AND indexname = 'idx_delegations_sender_address';

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'delegations' AND indexname = 'idx_delegations_delegate_address';

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'delegations' AND indexname = 'idx_delegations_status';

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'delegations' AND indexname = 'idx_delegations_block';

ROLLBACK;
