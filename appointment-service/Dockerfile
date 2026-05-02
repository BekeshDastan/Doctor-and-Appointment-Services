FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy doctor-service first — needed for the replace directive ../doctor-service
COPY doctor-service/ ./doctor-service/

# Copy appointment-service source
COPY appointment-service/ ./appointment-service/

WORKDIR /app/appointment-service

RUN go mod download

RUN go build -o service ./cmd/service-appointment

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/appointment-service/service .
COPY --from=builder /app/appointment-service/migrations ./migrations

CMD ["./service"]
