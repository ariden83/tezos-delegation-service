# Tezos Delegation Service

A Go service that collects and exposes delegations made on the Tezos protocol through a RESTful API, utilizing data from the TzKT API.
This service is designed to be lightweight and efficient, making it easy to integrate into existing systems or use as a standalone application.
It continuously polls the TzKT API for delegation data and stores it in a PostgreSQL database, providing a simple API for querying this data.
The service is built with a focus on performance, reliability, and ease of use, making it suitable for both development and production environments.
This project is part of the Tezos ecosystem and aims to provide developers and users with easy access to delegation data, enabling better insights into the Tezos network.


## Features

- Continuously polls and stores Tezos delegations from TzKT API
- Exposes delegation data through a REST API
- Supports filtering by year and pagination
- Includes a Swagger UI for API testing and documentation

## Requirements

- Docker and Docker Compose
- Git

## Quick Start with Docker

The easiest way to get started is using Docker Compose:

1. Clone this repository
   ```bash
   git clone https://github.com/tezos-delegation-service.git
   cd tezos-delegation-service
   ```

2. Build and start all services
   ```bash
   make docker-compose-up
   ```

3. Access the services:
   - **REST API**: http://localhost:8080
     - **Swagger UI**: http://localhost:8081/swagger/
   - **PostgreSQL**: localhost:5432 (username: postgres, password: postgres, database: tezos_delegations)

4. To stop all services:
   ```bash
   make docker-compose-down
   ```

### Using the Swagger UI for API Testing

After starting the services with `make docker-compose-up`, you can:

1. Open http://localhost:8081/swagger/ in your web browser
2. The Swagger UI will display all available API endpoints with documentation
3. Try out the endpoints directly from the UI:
   - Click on an endpoint (e.g., `/xtz/delegations`)
   - Click "Try it out"
   - Fill in any parameters
   - Click "Execute"
   - View the response

## Manual Setup (without Docker)

If you prefer to run the service directly:

1. Ensure you have Go 1.20+ and PostgreSQL installed
2. Configure the application in `config/config.yaml`
3. Install PostgreSQL client for migrations
4. Run migrations: `make db-migrate`
5. Build the application: `make build`
6. Run the application: `make run`

### Kubernetes Setup

For production deployment, we provide Kubernetes configuration:

1. Navigate to the `k8s` directory
2. Customize the configuration as needed (see `k8s/README.md`)
3. Deploy using: `./scripts/k8s-apply.sh`

See the [Kubernetes README](k8s/README.md) for detailed instructions.

## Project Structure

```
.
├── cmd/                         # Application entry points
│   └── tezos-delegation-service/  # Main application
├── config/                      # Configuration files
├── data/                        # Data storage (local development)
├── internal/                    # Private application code
│   ├── adapter/                 # External services adapters
│   │   ├── database/            # Database adapters
│   │   ├── metrics/             # Metrics adapters
│   │   └── tzktapi/             # TzKT API adapters
│   ├── model/                   # Domain models
│   └── usecase/                 # Business logic
├── k8s/                         # Kubernetes configuration
├── pkg/                         # Public packages
│   └── logger/                  # Logging utilities
├── scripts/                     # Utility scripts
├── sqitch_pg/                   # Database migrations (PostgreSQL)
└── docker-compose.yml           # Docker Compose configuration
```

## API

### GET /xtz/delegations

Returns a paginated list of Tezos delegations ordered by most recent first.

**Query Parameters:**
- `year` (optional): Filter delegations by year (format: YYYY)
- `page` (optional): Page number for pagination (default: 1)

**Response:**
```json
{
  "data": [
    {
      "timestamp": "2022-05-05T06:29:14Z",
      "amount": "125896",
      "delegator": "tz1a1SAaXRt9yoGMx29rh9FsBF4UzmvojdTL",
      "level": "2338084"
    },
    ...
  ]
}
```

### Health Check Endpoints

The service provides several health check endpoints for monitoring:

#### GET /health

Returns the overall health of the service.

**Response:**
```json
{
  "status": "ok",
  "uptime": "3h5m2s",
  "database": "ok",
  "ready": true,
  "shutdown": false
}
```

#### GET /health/live (Liveness Probe)

Indicates if the service is running. Kubernetes can use this to determine if the container needs to be restarted.

**Response:**
```json
{
  "status": "alive",
  "uptime": "3h5m2s",
  "started": "2023-04-10T12:00:00Z"
}
```

#### GET /health/ready (Readiness Probe)

Indicates if the service is ready to accept requests. Kubernetes can use this to determine if traffic should be routed to the container.

**Response:**
```json
{
  "status": "ready"
}
```

### Metrics

#### GET /metrics

Exposes Prometheus metrics for monitoring the service's performance and behavior. These metrics can be scraped by a Prometheus server.

The service collects the following types of metrics:
- API request metrics (count, duration, response size)
- Repository operation metrics (count, duration, errors)
- TzKT API metrics (requests, response time, sync statistics)
- Business metrics (total delegations, total amount delegated)

### Logging

The service supports structured logging with various outputs including console, file, and Graylog. Configure logging in the `config/config.yaml` file:

```yaml
logging:
  level: info  # Options: debug, info, warn, error
  format: json  # Options: json, text
  enable_file: false  # Set to true to write logs to a file
  file_path: /var/log/tezos-delegation-service.log  # Path for file logging
  graylog:
    enabled: false  # Set to true to enable Graylog logging
    url: graylog.example.com  # Graylog server URL
    port: 12201  # Graylog port (usually 12201 for GELF)
    facility: tezos-delegation-service  # Tag for identifying this service
```

Each log entry includes structured data like component name, timestamp, and contextual information to facilitate log analysis and troubleshooting.

## Testing

Run the tests with:
```
make test
```

## Database Migrations

This project uses [Sqitch](https://sqitch.org/) for database migrations.

```bash
# Install Sqitch first
# On Debian/Ubuntu:
# sudo apt-get install sqitch libdbd-sqlite3-perl

# Run migrations for SQLite (local development)
make db-migrate

# To create a new migration (example)
cd sqitch
sqitch add new_migration_name --requires previous_migration
# Then edit the created files in deploy/verify/revert directories
```

### PostgreSQL Support

For Docker/Kubernetes deployments using PostgreSQL, we provide separate Sqitch configuration in the `sqitch_pg` directory. The Docker setup automatically applies these migrations when the container starts.