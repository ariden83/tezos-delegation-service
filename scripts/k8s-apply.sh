#!/bin/bash

# Script to apply Kubernetes configuration

set -e

# Define colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Directory containing k8s manifests
K8S_DIR=$(dirname "$0")/../k8s

echo -e "${YELLOW}Applying Kubernetes configuration for Tezos Delegation Service${NC}"
echo -e "${YELLOW}=================================${NC}"

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}kubectl could not be found. Please install kubectl and try again.${NC}"
    exit 1
fi

# Check if kustomize is installed
if ! command -v kustomize &> /dev/null; then
    echo -e "${YELLOW}kustomize could not be found. Using kubectl apply -k instead.${NC}"
    KUSTOMIZE_CMD="kubectl apply -k"
else
    KUSTOMIZE_CMD="kustomize build $K8S_DIR | kubectl apply -f -"
fi

# Check if the namespace exists, if not create it
if ! kubectl get namespace tezos-delegation &> /dev/null; then
    echo -e "${YELLOW}Creating namespace tezos-delegation${NC}"
    kubectl apply -f "$K8S_DIR/00-namespace.yaml"
fi

# Apply the configuration
echo -e "${YELLOW}Applying configuration...${NC}"
if [[ $KUSTOMIZE_CMD == *"kustomize"* ]]; then
    eval $KUSTOMIZE_CMD
else
    kubectl apply -k "$K8S_DIR"
fi

echo -e "${GREEN}Configuration applied successfully!${NC}"
echo -e "${YELLOW}=================================${NC}"

# Show resources
echo -e "${YELLOW}Deployed resources:${NC}"
kubectl get all -n tezos-delegation

echo -e "\n${YELLOW}Deployed ConfigMaps:${NC}"
kubectl get configmaps -n tezos-delegation

echo -e "\n${YELLOW}Deployed Secrets:${NC}"
kubectl get secrets -n tezos-delegation

echo -e "\n${YELLOW}Deployed PVCs:${NC}"
kubectl get pvc -n tezos-delegation

echo -e "\n${YELLOW}Deployed Ingresses:${NC}"
kubectl get ingress -n tezos-delegation

echo -e "\n${GREEN}Deployment complete!${NC}"