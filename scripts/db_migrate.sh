#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

# Set PostgreSQL environment variables if not provided
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_NAME=${DB_NAME:-"tezos_delegations"}
DB_USER=${DB_USER:-"postgres"}
DB_PASSWORD=${DB_PASSWORD:-"postgres"}

# Set environment variables for PostgreSQL client
export PGHOST="$DB_HOST"
export PGPORT="$DB_PORT"
export PGDATABASE="$DB_NAME"
export PGUSER="$DB_USER"
export PGPASSWORD="$DB_PASSWORD"

# Check if PostgreSQL client is installed
if ! command -v psql &> /dev/null; then
    echo "Error: psql is not installed."
    echo "Please install PostgreSQL client"
    exit 1
fi

# Création du schéma app
psql -v ON_ERROR_STOP=1 -c "CREATE SCHEMA IF NOT EXISTS app;"

# Exécution des scripts SQL de migration directement
for sql_file in "$REPO_ROOT"/sqitch_pg/deploy/*.sql; do
  echo "Applying migration: $sql_file"
  psql -v ON_ERROR_STOP=1 -f "$sql_file"
done

echo "PostgreSQL database migrations completed successfully"