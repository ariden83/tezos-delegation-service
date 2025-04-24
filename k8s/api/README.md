# Kubernetes Configuration for Tezos Delegation Service

This directory contains the Kubernetes configuration files for deploying the Tezos Delegation Service in a Kubernetes cluster.

## Files Structure

- `00-namespace.yaml`: Defines the Kubernetes namespace for the application
- `01-configmap.yaml`: Contains the application configuration
- `02-secret.yaml`: Contains sensitive information (replace with your real secrets)
- `03-service.yaml`: Defines the Kubernetes service to expose the application
- `04-deployment.yaml`: Defines the deployment specification for the application
- `05-pvc.yaml`: Persistent Volume Claim for data storage
- `06-hpa.yaml`: HorizontalPodAutoscaler for automatic scaling
- `07-ingress.yaml`: Ingress configuration for external access
- `08-network-policy.yaml`: Network policies for security
- `09-service-monitor.yaml`: Prometheus ServiceMonitor configuration
- `kustomization.yaml`: Kustomize configuration for managing resources

## Deployment

### Prerequisites

- A Kubernetes cluster
- kubectl installed and configured
- Helm (optional, for advanced deployments)
- Prometheus Operator (for ServiceMonitor)

### Quick Deployment

You can deploy the application using:

```bash
./scripts/k8s-apply.sh
```

Or manually:

```bash
kubectl apply -k k8s/
```

### Configuration

Before deploying, you should customize:

1. **Domain Name**: In `07-ingress.yaml`, replace `tezos-delegations.example.com` with your domain
2. **Secrets**: In `02-secret.yaml`, replace base64 encoded values with your real secrets
3. **Resources**: In `04-deployment.yaml`, adjust resource limits and requests based on your needs
4. **Storage**: In `05-pvc.yaml`, adjust storage size and class as needed

### Accessing the Application

Once deployed, the application will be available at:

- Internal: `http://tezos-delegation-service.tezos-delegation.svc.cluster.local`
- External: `https://tezos-delegations.example.com` (after configuring DNS)

### Monitoring

The service exposes Prometheus metrics at `/metrics`. A ServiceMonitor is included for integration with Prometheus Operator.

### Scaling

The deployment includes a HorizontalPodAutoscaler that will automatically scale the number of replicas based on CPU and memory usage. The configuration sets:

- Minimum replicas: 2
- Maximum replicas: 5
- Target CPU utilization: 70%
- Target memory utilization: 80%

### High Availability

The deployment is configured for high availability:
- Multiple replicas
- Rolling update strategy
- Liveness and readiness probes
- Persistent storage

## Maintenance

### Upgrading

To upgrade the application:

1. Update the image version in `kustomization.yaml`
2. Reapply the configuration:
   ```bash
   kubectl apply -k k8s/
   ```

### Monitoring Logs

```bash
kubectl logs -n tezos-delegation deployment/tezos-delegation
```

### Checking Application Status

```bash
kubectl get all -n tezos-delegation
```

### Accessing the Pod Shell

```bash
kubectl exec -it -n tezos-delegation deploy/tezos-delegation -- /bin/sh
```