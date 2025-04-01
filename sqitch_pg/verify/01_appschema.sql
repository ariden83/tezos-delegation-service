-- Verify tezos-delegation-service:01_appschema on pg

BEGIN;

SELECT pg_catalog.has_schema_privilege('app', 'usage');

ROLLBACK;
