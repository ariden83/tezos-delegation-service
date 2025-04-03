FROM golang:1.18-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev git sqlite-dev
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.2

COPY go.mod ./

RUN go mod download && go mod verify

COPY . .

RUN golangci-lint run ./... --timeout=5m

RUN CGO_ENABLED=1 go test -mod=mod -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o tezos-delegation-api ./cmd/tezos-delegation-api

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/tezos-delegation-api /app/tezos-delegation-api
COPY --from=builder /app/config /app/config
COPY --from=builder /app/scripts /app/scripts

RUN mkdir -p /app/data

WORKDIR /app

EXPOSE 8080

CMD ["./tezos-delegation-api"]