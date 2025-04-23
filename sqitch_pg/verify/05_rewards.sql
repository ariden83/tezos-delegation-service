-- Verify tezos-delegation-service:05_rewards to pg

BEGIN;

SELECT id, recipient_id, source_id, cycle, amount, timestamp, created_at
FROM app.rewards
WHERE FALSE;

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'rewards' AND indexname = 'idx_rewards_recipient_id';

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'rewards' AND indexname = 'idx_rewards_source_id';

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'rewards' AND indexname = 'idx_rewards_cycle';

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'rewards' AND indexname = 'idx_rewards_timestamp';

COMMIT;
