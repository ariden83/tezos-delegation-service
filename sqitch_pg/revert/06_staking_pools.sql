-- Revert tezos-delegation-service:06_staking_pools to pg

BEGIN;

DROP INDEX IF EXISTS app.idx_staking_pools_address;
DROP INDEX IF EXISTS app.idx_staking_pools_token;

DROP TABLE IF EXISTS app.staking_pools;

COMMIT;
