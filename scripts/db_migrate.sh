#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

# Check if sqitch is installed
if ! command -v sqitch &> /dev/null; then
    echo "Error: sqitch is not installed."
    echo "Please install sqitch: https://sqitch.org/download/"
    exit 1
fi

# Check if libdbd-pg-perl is installed
if ! dpkg -l | grep -q libdbd-pg-perl; then
    echo "Warning: libdbd-pg-perl might not be installed."
    echo "Please install it: apt-get install libdbd-pg-perl"
fi

# Change to sqitch_pg directory
cd "$REPO_ROOT/sqitch_pg" || exit 1

# Set PostgreSQL environment variables if not provided
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_NAME=${DB_NAME:-"tezos_delegations"}
DB_USER=${DB_USER:-"postgres"}
DB_PASSWORD=${DB_PASSWORD:-"postgres"}

# Set environment variables for sqitch
export PGHOST="$DB_HOST"
export PGPORT="$DB_PORT"
export PGDATABASE="$DB_NAME"
export PGUSER="$DB_USER"
export PGPASSWORD="$DB_PASSWORD"

# Deploy sqitch changes
# Note: Using --verify makes sure our verify scripts work properly
sqitch deploy --verify

echo "PostgreSQL database migrations completed successfully"