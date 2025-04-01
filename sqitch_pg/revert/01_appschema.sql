-- Revert tezos-delegation-service:01_appschema from pg

BEGIN;

DROP SCHEMA IF EXISTS app CASCADE;

COMMIT;
