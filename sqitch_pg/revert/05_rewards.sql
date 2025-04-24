-- Revert tezos-delegation-service:05_rewards to pg

BEGIN;

DROP INDEX IF EXISTS app.idx_rewards_recipient_address;
DROP INDEX IF EXISTS app.idx_rewards_source_address;
DROP INDEX IF EXISTS app.idx_rewards_cycle;
DROP INDEX IF EXISTS app.idx_rewards_timestamp;

DROP TABLE IF EXISTS app.rewards;

COMMIT;
