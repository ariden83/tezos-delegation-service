-- Verify tezos-delegation-service:02_delegations on pg

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

ROLLBACK;
