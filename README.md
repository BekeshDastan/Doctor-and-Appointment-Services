# Medical Scheduling Platform — Assignment 3

## 1. Project Overview

This project extends the Medical Scheduling Platform (AP2 Assignment 2) with two additional infrastructure layers:

**What changed vs Assignment 2:**
- **PostgreSQL schema** is now managed exclusively through versioned migration files (`golang-migrate`). No raw DDL exists in application code.
- **NATS message broker** added: every successful write operation publishes a domain event.
- **Notification Service** — a new third binary that subscribes to all events and logs structured JSON to stdout.
- Directory names updated: `doctor/` → `doctor-service/`, `appointment/` → `appointment-service/`, plus new `notification-service/`.
- Database transactions added to `CreateDoctor` (email uniqueness check + insert in one transaction).
- Each service has its own **Dockerfile**; the entire system can be started with a single `docker compose up --build`.

**What did NOT change:**
- Domain models (`Doctor`, `Appointment`, `Status`) are identical.
- Use-case business logic and rules are identical.
- gRPC service contracts and generated proto stubs are identical.
- Clean Architecture layering and dependency direction are identical.

---

## 2. Broker Choice — NATS

**Chosen broker: NATS (Core)**

**Reason:** NATS fits the use case perfectly — the Notification Service is stateless and only needs to log events. Core NATS provides simple fire-and-forget pub/sub with minimal setup (single Docker image, no queues to configure). There is no need for durable delivery because a missed log event is not a critical failure.

**What would change if we switched to RabbitMQ:**
- Replace `github.com/nats-io/nats.go` with `github.com/rabbitmq/amqp091-go`
- Declare a fanout exchange `ap2.events`; each subscriber binds its own exclusive queue
- Replace `NATS_URL` with `AMQP_URL` in all three services
- Add explicit message acknowledgment (`d.Ack(false)`) in the notification handler

**When durable delivery becomes necessary in production:**
If event loss is unacceptable (e.g., billing events, audit logs), switch to **RabbitMQ with persistent queues + publisher confirms**, or **NATS JetStream** with subject-level persistence.

---

## 3. Environment Variables

| Service | Variable | Default | Description |
|---|---|---|---|
| doctor-service | `DOCTOR_DATABASE_URL` | `postgres://postgres:12345678@localhost:5432/doctor?sslmode=disable` | PostgreSQL DSN |
| doctor-service | `DOCTOR_GRPC_PORT` | `50051` | gRPC listen port |
| doctor-service | `NATS_URL` | `nats://localhost:4222` | NATS connection URL |
| appointment-service | `APPOINTMENT_DATABASE_URL` | `postgres://postgres:12345678@localhost:5432/appointment?sslmode=disable` | PostgreSQL DSN |
| appointment-service | `APPOINTMENT_GRPC_PORT` | `50052` | gRPC listen port |
| appointment-service | `DOCTOR_SERVICE_URL` | `localhost:50051` | Doctor Service gRPC address |
| appointment-service | `NATS_URL` | `nats://localhost:4222` | NATS connection URL |
| notification-service | `NATS_URL` | `nats://localhost:4222` | NATS connection URL |

> **In Docker Compose** the defaults are overridden automatically — services communicate by container name (`postgres`, `nats`, `doctor-service`) instead of `localhost`.

---

## 4. Running with Docker (recommended)

All five containers (Postgres, NATS, doctor-service, appointment-service, notification-service) start with one command:

```bash
docker compose up --build
```

Run in background:
```bash
docker compose up --build -d
```

View logs of a specific service:
```bash
docker compose logs -f notification-service
docker compose logs -f doctor-service
docker compose logs -f appointment-service
```

Stop everything:
```bash
docker compose down
```

Stop and delete all data (volumes):
```bash
docker compose down -v
```

### What happens automatically
1. Postgres starts and runs `scripts/init-db.sh` — creates `doctor` and `appointment` databases.
2. `doctor-service` starts, runs migrations (`000001_create_doctors.up.sql`), then serves gRPC on `:50051`.
3. `appointment-service` starts after Postgres and `doctor-service` are ready, runs migrations, serves gRPC on `:50052`.
4. `notification-service` starts, connects to NATS, subscribes to all three subjects, and waits for events.

### Dockerfile overview

| Service | Build context | Notes |
|---|---|---|
| `doctor-service/Dockerfile` | `./doctor-service` | Multi-stage: `golang:1.25-alpine` → `alpine:3.20`. Binary + `migrations/` copied to runtime image. |
| `appointment-service/Dockerfile` | `.` (project root) | Build context must include `doctor-service/` due to the `replace ../doctor-service` directive in `go.mod`. |
| `notification-service/Dockerfile` | `./notification-service` | Binary only — no migrations, no database. |

---

## 5. Running Locally (without Docker)

### Prerequisites
- Go 1.25+
- PostgreSQL 14+ running on `localhost:5432`
- NATS server on `localhost:4222`

### Start NATS locally
```bash
# Docker (recommended)
docker run -d -p 4222:4222 nats:2.10-alpine

# or native binary
nats-server
```

### Create databases manually (first time only)
```bash
psql -U postgres -c "CREATE DATABASE doctor;"
psql -U postgres -c "CREATE DATABASE appointment;"
```
Migrations are applied automatically when each service starts.

### Start services (in order)

**Terminal 1 — Doctor Service:**
```bash
cd doctor-service
go run ./cmd/service-doctor
```

**Terminal 2 — Appointment Service:**
```bash
cd appointment-service
go run ./cmd/service-appointment
```

**Terminal 3 — Notification Service:**
```bash
cd notification-service
go run ./cmd/notification-service
```

> **Why this order?** Appointment Service dials Doctor Service at startup. If Doctor Service is not running, Appointment Service exits with a fatal error.

---

## 6. Database Migrations

Migrations run **automatically on service startup** before the gRPC server begins accepting requests.

### Manual rollback (golang-migrate CLI)

Install the CLI:
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

**Roll back one step (tested during defense):**
```bash
migrate -path doctor-service/migrations \
  -database "postgres://postgres:12345678@localhost:5432/doctor?sslmode=disable" down 1

migrate -path appointment-service/migrations \
  -database "postgres://postgres:12345678@localhost:5432/appointment?sslmode=disable" down 1
```

**Re-apply:**
```bash
migrate -path doctor-service/migrations \
  -database "postgres://postgres:12345678@localhost:5432/doctor?sslmode=disable" up

migrate -path appointment-service/migrations \
  -database "postgres://postgres:12345678@localhost:5432/appointment?sslmode=disable" up
```

**Verify schema:**
```bash
psql -U postgres -d doctor -c '\d doctors'
psql -U postgres -d appointment -c '\d appointments'
```

### Migration files

```
doctor-service/migrations/
  000001_create_doctors.up.sql    — CREATE TABLE doctors (...)
  000001_create_doctors.down.sql  — DROP TABLE IF EXISTS doctors

appointment-service/migrations/
  000001_create_appointments.up.sql    — CREATE TABLE appointments (...)
  000001_create_appointments.down.sql  — DROP TABLE IF EXISTS appointments
```

---

## 7. Event Contract

All events are serialized as JSON and published to NATS.

| Subject | Published by | Trigger | Required JSON Fields |
|---|---|---|---|
| `doctors.created` | doctor-service | `CreateDoctor` succeeds | `event_type`, `occurred_at`, `id`, `full_name`, `specialization`, `email` |
| `appointments.created` | appointment-service | `CreateAppointment` succeeds | `event_type`, `occurred_at`, `id`, `title`, `doctor_id`, `status` |
| `appointments.status_updated` | appointment-service | `UpdateAppointmentStatus` succeeds | `event_type`, `occurred_at`, `id`, `old_status`, `new_status` |

### Example payloads

**doctors.created:**
```json
{
  "event_type": "doctors.created",
  "occurred_at": "2026-05-02T10:23:44Z",
  "id": "doc-1",
  "full_name": "Dr. Aisha Seitkali",
  "specialization": "Cardiology",
  "email": "a.seitkali@clinic.kz"
}
```

**appointments.created:**
```json
{
  "event_type": "appointments.created",
  "occurred_at": "2026-05-02T10:24:01Z",
  "id": "appt-1",
  "title": "Initial cardiac consultation",
  "doctor_id": "doc-1",
  "status": "new"
}
```

**appointments.status_updated:**
```json
{
  "event_type": "appointments.status_updated",
  "occurred_at": "2026-05-02T10:25:10Z",
  "id": "appt-1",
  "old_status": "new",
  "new_status": "in_progress"
}
```

---

## 8. Notification Service

The Notification Service is a standalone subscriber binary with **no gRPC server, no HTTP server, and no database**.

**What it does:**
1. On startup: connects to NATS with exponential backoff (1s → 2s → 4s → 8s → 16s). After 5 failed attempts, exits with non-zero code.
2. Subscribes to three subjects: `doctors.created`, `appointments.created`, `appointments.status_updated`.
3. For each message: deserializes the JSON payload and prints one structured JSON log line to stdout.
4. On `SIGINT`/`SIGTERM`: drains in-flight messages, closes the NATS connection, exits with code 0.

**Log output format (one line per event):**
```json
{"time":"2026-05-02T10:23:44Z","subject":"doctors.created","event":{"email":"a.seitkali@clinic.kz","event_type":"doctors.created","full_name":"Dr. Aisha Seitkali","id":"...","occurred_at":"2026-05-02T10:23:44Z","specialization":"Cardiology"}}
```

**How to verify during a live demo:**
```bash
# In one terminal watch the notification-service logs
docker compose logs -f notification-service

# In another terminal call CreateDoctor
grpcurl -plaintext -d '{"full_name":"Dr. Test","specialization":"X","email":"t@t.com"}' \
  localhost:50051 doctor.DoctorService/CreateDoctor
```
Within 1–2 seconds a JSON line with `"subject":"doctors.created"` appears in the notification-service terminal.

---

## 9. gRPC Test Commands

### Create a Doctor
```bash
grpcurl -plaintext -d '{
  "full_name": "Dr. Aisha Seitkali",
  "specialization": "Cardiology",
  "email": "a.seitkali@clinic.kz"
}' localhost:50051 doctor.DoctorService/CreateDoctor
```
**Expected notification-service output:**
```json
{"time":"...","subject":"doctors.created","event":{"email":"a.seitkali@clinic.kz","event_type":"doctors.created","full_name":"Dr. Aisha Seitkali","id":"<uuid>","occurred_at":"...","specialization":"Cardiology"}}
```

### Get a Doctor
```bash
grpcurl -plaintext -d '{"id": "<doctor-id>"}' localhost:50051 doctor.DoctorService/GetDoctor
```

### List All Doctors
```bash
grpcurl -plaintext -d '{}' localhost:50051 doctor.DoctorService/ListDoctors
```

### Create an Appointment
```bash
grpcurl -plaintext -d '{
  "title": "Initial cardiac consultation",
  "description": "First visit",
  "doctor_id": "<doctor-id>"
}' localhost:50052 appointment.AppointmentService/CreateAppointment
```
**Expected notification-service output:**
```json
{"time":"...","subject":"appointments.created","event":{"doctor_id":"<doctor-id>","event_type":"appointments.created","id":"<uuid>","occurred_at":"...","status":"new","title":"Initial cardiac consultation"}}
```

### Update Appointment Status
```bash
grpcurl -plaintext -d '{
  "id": "<appointment-id>",
  "status": "in_progress"
}' localhost:50052 appointment.AppointmentService/UpdateAppointmentStatus
```
**Expected notification-service output:**
```json
{"time":"...","subject":"appointments.status_updated","event":{"event_type":"appointments.status_updated","id":"<appointment-id>","new_status":"in_progress","occurred_at":"...","old_status":"new"}}
```

### List All Appointments
```bash
grpcurl -plaintext -d '{}' localhost:50052 appointment.AppointmentService/ListAppointments
```

---

## 10. Error Handling

| Situation | Expected Behaviour |
|---|---|
| Database unavailable on startup | Service exits with non-zero code and descriptive log message |
| Database query fails at runtime | Returns `codes.Internal` with descriptive message |
| Broker unavailable at startup (Doctor/Appointment) | Service starts normally, logs WARN, RPC still succeeds |
| Broker publish fails during RPC | Logs error, RPC response unaffected |
| Broker unavailable at startup (Notification Service) | Retries with 1s/2s/4s/8s/16s backoff, exits non-zero after 5 attempts |
| Duplicate email in DB | Returns `codes.AlreadyExists` |
| Row not found in DB | Returns `codes.NotFound` |
| Doctor not found via gRPC | Returns `codes.FailedPrecondition` |
| Doctor Service unreachable | Returns `codes.Unavailable` |
| Invalid status transition (done→new) | Returns `codes.InvalidArgument` |

---

## 11. Consistency Trade-offs

Because broker publishing is **best-effort**, a process crash between the DB commit and `nc.Publish()` will silently lose the event.

**Concrete failure scenario:**
1. `CreateDoctor` inserts a row in PostgreSQL — committed.
2. Process crashes before `nc.Publish("doctors.created", ...)` executes.
3. Notification Service never receives the event.
4. The doctor exists in the DB, but no log was produced.

**How to fix in production:**

| Pattern | How it helps |
|---|---|
| **Outbox pattern** | Write event to an `outbox` table in the same DB transaction; a separate relay process reads and publishes. Guarantees at-least-once delivery. |
| **NATS JetStream** | Adds persistent streams and consumer acknowledgment to NATS. Publisher can wait for `PubAck` before returning success. |
| **RabbitMQ publisher confirms** | Exchange-level delivery confirmation; publisher blocks until broker persists the message. |

---

## 12. Broker Comparison: NATS vs RabbitMQ

| Feature | NATS Core | RabbitMQ |
|---|---|---|
| **Persistence** | None — fire-and-forget | Queue-level durability; messages survive broker restart |
| **Delivery guarantee** | At-most-once | At-least-once (with persistent queues + publisher confirms) |
| **Setup complexity** | Single binary or one Docker image | Requires Erlang runtime, management plugin, exchange/queue setup |
| **Throughput** | Very high (~millions of msgs/sec) | High (~hundreds of thousands of msgs/sec) |
| **Use case** | Stateless notifications, metrics streaming | Financial transactions, task queues, guaranteed delivery |

**When to choose NATS:** Stateless notifications, low-latency internal events where occasional loss is acceptable.

**When to choose RabbitMQ:** Order processing, payment events, any domain where losing a message means losing money or data integrity.

---

## 13. Architecture Diagram

```
 ┌─────────────────────────────────────────────┐
 │              Client / CLI                    │
 │           (grpcurl commands)                 │
 └────────┬─────────────────────┬───────────────┘
          │ gRPC                │ gRPC
          ▼                     ▼
 ┌─────────────────┐    ┌───────────────────┐
 │  DOCTOR SERVICE │    │ APPOINTMENT SVC   │
 │   :50051 (gRPC) │◄───│   :50052 (gRPC)   │
 │  [migrations]   │    │  [migrations]     │
 └────────┬────────┘    └────────┬──────────┘
          │ SQL                  │ SQL
          ▼                      ▼
 ┌────────────────┐    ┌────────────────────┐
 │  PostgreSQL    │    │  PostgreSQL        │
 │  doctor DB     │    │  appointment DB    │
 └────────────────┘    └────────────────────┘
          │ doctors.created      │ appointments.created
          │                      │ appointments.status_updated
          └──────────┬───────────┘
                     │ NATS Publish
                     ▼
              ┌─────────────┐
              │    NATS     │
              │  :4222      │
              └──────┬──────┘
                     │ Subscribe (all 3 subjects)
                     ▼
        ┌────────────────────────┐
        │  NOTIFICATION SERVICE  │
        │  (no port, no DB)      │
        │  → stdout JSON logs    │
        └────────────────────────┘
```

---

## 14. Project Structure

```
Assignment 2/
├── doctor-service/
│   ├── cmd/service-doctor/main.go
│   ├── internal/
│   │   ├── app/migrate.go           ← golang-migrate runner
│   │   ├── config/config.go         ← reads env vars incl. NATS_URL
│   │   ├── event/publisher.go       ← Publisher interface + NATS impl + Noop
│   │   ├── model/doctor.go
│   │   ├── repository/doctor_repository.go  ← CreateWithEmailCheck (tx)
│   │   ├── transport/grpc/doctor_server.go
│   │   └── usecase/
│   ├── migrations/
│   │   ├── 000001_create_doctors.up.sql
│   │   └── 000001_create_doctors.down.sql
│   ├── proto/
│   ├── Dockerfile                   ← multi-stage, context: ./doctor-service
│   ├── go.mod
│   └── go.sum
│
├── appointment-service/
│   ├── cmd/service-appointment/main.go
│   ├── internal/
│   │   ├── app/migrate.go
│   │   ├── client/doctor_grpc_client.go
│   │   ├── config/config.go
│   │   ├── event/publisher.go
│   │   ├── model/appointment.go
│   │   ├── repository/appointment_repository.go
│   │   ├── transport/grpc/appointment_server.go
│   │   └── usecase/
│   ├── migrations/
│   │   ├── 000001_create_appointments.up.sql
│   │   └── 000001_create_appointments.down.sql
│   ├── proto/
│   ├── Dockerfile                   ← build context is project root (needs doctor-service/)
│   ├── go.mod
│   └── go.sum
│
├── notification-service/
│   ├── cmd/notification-service/main.go
│   ├── internal/subscriber/subscriber.go
│   ├── Dockerfile                   ← context: ./notification-service
│   ├── go.mod
│   └── go.sum
│
├── scripts/
│   └── init-db.sh                   ← creates doctor + appointment DBs on first Postgres start
├── docker-compose.yml               ← all 5 services (Postgres, NATS, 3 Go services)
├── .dockerignore
└── README.md
```

---

## 15. Clean Architecture Preservation

**Dependency rule is fully preserved:**

```
transport/grpc  →  usecase  →  repository  →  model
                      ↓
                  event.Publisher (interface)
```

- `transport/grpc` handlers remain thin — only proto↔domain mapping and error translation.
- Use cases contain the business logic and are the only layer that publishes events.
- `event.Publisher` is an interface injected at startup — use cases never import NATS directly.
- Domain models (`model/`) have zero external dependencies.
- No infrastructure types (NATS, `*sql.DB`, `migrate.Migrate`) leak into domain or use-case layers.
