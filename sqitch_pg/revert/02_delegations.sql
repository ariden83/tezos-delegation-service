-- Revert tezos-delegation-service:02_delegations from pg

BEGIN;

DROP INDEX IF EXISTS app.idx_delegations_level;
DROP INDEX IF EXISTS app.idx_delegations_delegator;
DROP INDEX IF EXISTS app.idx_delegations_timestamp;
DROP TABLE IF EXISTS app.delegations;

COMMIT;
