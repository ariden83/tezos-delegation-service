-- Revert tezos-delegation-service:02_accounts from pg

BEGIN;

DROP INDEX IF EXISTS app.idx_accounts_address;
DROP INDEX IF EXISTS app.idx_accounts_type;

DROP TABLE IF EXISTS app.accounts;

COMMIT;
