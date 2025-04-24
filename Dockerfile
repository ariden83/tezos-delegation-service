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
# On définit l'argument TARGET avec "api" comme valeur par défaut
ARG TARGET=api
WORKDIR /app
COPY --from=modules /go/pkg /go/pkg
COPY . .
# On utilise la valeur de TARGET pour déterminer quel binaire compiler
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o tezos-delegation-$TARGET ./cmd/tezos-delegation-$TARGET

# Stage 5: Final lightweight image
FROM alpine:latest
ARG TARGET=api
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata postgresql-client
COPY --from=builder /app/tezos-delegation-$TARGET /app/tezos-delegation-$TARGET
COPY --from=builder /app/config /app/config
COPY --from=builder /app/scripts /app/scripts

EXPOSE 8080
CMD ["./tezos-delegation-$TARGET"]