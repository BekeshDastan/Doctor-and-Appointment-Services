FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o service ./cmd/notification-service

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/service .

CMD ["./service"]
