-- Verify tezos-delegation-service:06_staking_pools to pg

BEGIN;

SELECT id, address, name, staking_token, created_at
FROM app.staking_pools
WHERE FALSE;

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'staking_pools' AND indexname = 'idx_staking_pools_address';

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'staking_pools' AND indexname = 'idx_staking_pools_token';

COMMIT;
