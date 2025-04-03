#!/bin/bash
set -e

# Installation de Sqitch pour PostgreSQL dans un conteneur Alpine
apk update && apk add --no-cache perl perl-dbd-pg

# Exécution des migrations SQL directement
export PGHOST=localhost
export PGDATABASE=tezos_delegations
export PGUSER=postgres
export PGPASSWORD=$POSTGRES_PASSWORD

# Création du schéma app
psql -v ON_ERROR_STOP=1 -d "$PGDATABASE" -c "CREATE SCHEMA IF NOT EXISTS app;"

# Exécution des scripts SQL de migration directement 
for sql_file in /app/sqitch_pg/deploy/*.sql; do
  echo "Applying migration: $sql_file"
  psql -v ON_ERROR_STOP=1 -d "$PGDATABASE" -f "$sql_file"
done

echo "PostgreSQL database migrations completed successfully"