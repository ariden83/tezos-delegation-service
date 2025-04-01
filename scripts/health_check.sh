#!/bin/bash

# Health check script for the Tezos Delegation Service
# This script checks the health, liveness, and readiness endpoints

# Configuration
SERVER_URL=${1:-"http://localhost:8080"}
HEALTH_ENDPOINT="${SERVER_URL}/health"
LIVENESS_ENDPOINT="${SERVER_URL}/health/live"
READINESS_ENDPOINT="${SERVER_URL}/health/ready"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to check an endpoint
check_endpoint() {
  local endpoint=$1
  local name=$2
  
  echo -e "${YELLOW}Checking ${name}...${NC}"
  
  # Make the HTTP request and capture the status code and response
  response=$(curl -s -o /dev/null -w "%{http_code}|%{size_download}" ${endpoint})
  status_code=$(echo $response | cut -d'|' -f1)
  response_size=$(echo $response | cut -d'|' -f2)
  
  # Display detailed information
  echo "Endpoint: ${endpoint}"
  echo "Status code: ${status_code}"
  echo "Response size: ${response_size} bytes"
  
  # Check if the response was successful (200 OK)
  if [ "$status_code" -eq 200 ]; then
    echo -e "${GREEN}✓ ${name} check passed${NC}"
    
    # If requested, show the response content
    if [ "$3" == "verbose" ]; then
      echo "Response content:"
      curl -s ${endpoint} | jq '.'
    fi
  else
    echo -e "${RED}✗ ${name} check failed${NC}"
    echo "Response content:"
    curl -s ${endpoint} | jq '.'
  fi
  
  echo ""
}

# Main script
echo -e "${YELLOW}Tezos Delegation Service Health Check${NC}"
echo "Server URL: ${SERVER_URL}"
echo ""

# Check all endpoints
check_endpoint "${HEALTH_ENDPOINT}" "Health" "verbose"
check_endpoint "${LIVENESS_ENDPOINT}" "Liveness" "verbose"
check_endpoint "${READINESS_ENDPOINT}" "Readiness" "verbose"

# Check metrics endpoint
echo -e "${YELLOW}Checking Metrics...${NC}"
metrics_response=$(curl -s -o /dev/null -w "%{http_code}" ${SERVER_URL}/metrics)
if [ "$metrics_response" -eq 200 ]; then
  echo -e "${GREEN}✓ Metrics endpoint is available${NC}"
else
  echo -e "${RED}✗ Metrics endpoint check failed (Status: ${metrics_response})${NC}"
fi

echo ""
echo -e "${YELLOW}Health check completed${NC}"