-- Revert tezos-delegation-service:07_sync_state to pg

BEGIN;

DROP TABLE IF EXISTS app.sync_state;

COMMIT;
