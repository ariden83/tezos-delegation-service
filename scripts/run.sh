#!/bin/bash

# Set up environment
mkdir -p data

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed or not in PATH"
    echo "Please install Go 1.18+ and try again"
    exit 1
fi

# Build the application
echo "Building the application..."
go build -o tezos-delegation-api ./cmd/tezos-delegation-api

# Run the application
echo "Starting the Tezos Delegation API Service..."
./tezos-delegation-api