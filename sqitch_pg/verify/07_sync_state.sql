-- Verify tezos-delegation-service:07_sync_state to pg

BEGIN;

SELECT id, source, last_synced_level, last_synced_timestamp
FROM app.sync_state
WHERE FALSE;

COMMIT;
