FROM golang:1.18-alpine AS builder

# Set the working directory
WORKDIR /app

# Install required tools for linting and SQLite support
RUN apk add --no-cache gcc musl-dev git sqlite-dev
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.2

# Copy go.mod and go.sum
COPY go.mod ./

# Download dependencies
RUN go mod download && go mod verify

# Copy the source code
COPY . .

# Run linting (will exit build if it fails)
RUN golangci-lint run ./... --timeout=5m

# Run tests (will exit build if any test fails)
RUN CGO_ENABLED=1 go test -mod=mod -v ./...

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o tezos-delegation-api ./cmd/tezos-delegation-api

# Use a small alpine image for the final image
FROM alpine:latest

# Install required packages
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/tezos-delegation-api /app/tezos-delegation-api
COPY --from=builder /app/config /app/config
COPY --from=builder /app/scripts /app/scripts

# Create data directory
RUN mkdir -p /app/data

# Set working directory
WORKDIR /app

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./tezos-delegation-api"]