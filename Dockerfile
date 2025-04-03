# Stage 1: Modules caching
FROM golang:1.20-alpine AS modules
WORKDIR /modules
COPY go.mod go.sum ./
RUN go mod download

# Stage 2: Linting
FROM golangci/golangci-lint:v1.51.2-alpine AS lint
WORKDIR /app
COPY --from=modules /go/pkg /go/pkg
COPY . .
RUN golangci-lint run --timeout=5m

# Stage 3: Testing
FROM golang:1.20-alpine AS test
WORKDIR /app
COPY --from=modules /go/pkg /go/pkg
COPY . .
RUN go test -mod=mod -v ./...

# Stage 4: Building
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY --from=modules /go/pkg /go/pkg
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o tezos-delegation-service ./cmd/tezos-delegation-service

# Stage 5: Final lightweight image
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata postgresql-client
COPY --from=builder /app/tezos-delegation-service /app/tezos-delegation-service
COPY --from=builder /app/config /app/config
COPY --from=builder /app/scripts /app/scripts
RUN mkdir -p /app/data
EXPOSE 8080
CMD ["./tezos-delegation-service"]