-- Verify tezos-delegation-service:02_accounts on pg

BEGIN;

SELECT id, address, alias, type, created_at
FROM app.accounts
WHERE FALSE;

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'accounts' AND indexname = 'idx_accounts_address';

SELECT 1/COUNT(*)
FROM pg_indexes
WHERE tablename = 'accounts' AND indexname = 'idx_accounts_type';

ROLLBACK;
