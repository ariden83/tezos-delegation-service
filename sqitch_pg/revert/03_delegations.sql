-- Revert tezos-delegation-service:03_delegations to pg

BEGIN;

DROP INDEX IF EXISTS app.idx_delegations_level;
DROP INDEX IF EXISTS app.idx_delegations_delegator;
DROP INDEX IF EXISTS app.idx_delegations_timestamp;
DROP INDEX IF EXISTS app.idx_delegations_sender_address;
DROP INDEX IF EXISTS app.idx_delegations_delegate_address;
DROP INDEX IF EXISTS app.idx_delegations_status;
DROP INDEX IF EXISTS app.idx_delegations_block;

DROP TABLE IF EXISTS app.delegations;

COMMIT;