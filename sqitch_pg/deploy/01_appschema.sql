-- Deploy tezos-delegation-service:01_appschema to pg

BEGIN;

-- Create main application schema
CREATE SCHEMA IF NOT EXISTS app;

COMMIT;
