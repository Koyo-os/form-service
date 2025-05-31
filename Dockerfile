FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o /app/bin/form ./cmd/main.go

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/bin/form ./form


COPY --from=builder /app/config.yaml .

CMD ["./form"]