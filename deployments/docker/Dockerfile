FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o /app/bin/poll-service ../../cmd/main.go

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/bin/poll-service ./poll-service
COPY config.yaml ./config.yaml

CMD ["./poll-service"]