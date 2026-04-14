
# Doctor and Appointment Services Platform - gRPC Migration

## 1. Project Overview and Purpose

This project implements a **Medical Scheduling Platform** consisting of two independent microservices: **Doctor Service** and **Appointment Service**. The system achieves **inter-service communication exclusively through gRPC** and follows **Clean Architecture** principles.

**Key Purpose:**
- Demonstrate service decomposition with **bounded contexts** (Doctor and Appointment domains)
- Implement **gRPC-based synchronous communication** between services
- Maintain strict **data ownership** — each service manages its own database
- Preserve **Clean Architecture layering** across the migration from REST to gRPC


## 2. Service Responsibilities and Data Ownership

### **Doctor Service** (gRPC on port `50051`)
- **Owns**: Doctor entity and doctor-related data
- **Responsibilities**: 
  - `CreateDoctor` — Create a doctor with full_name, specialization, and email (email must be unique)
  - `GetDoctor` — Retrieve doctor by ID (returns `NOT_FOUND` if ID does not exist)
  - `ListDoctors` — Return all stored doctors
- **Protocol**: gRPC
- **Database**: PostgreSQL (`doctor` database)

### **Appointment Service** (gRPC on port `50052`)
- **Owns**: Appointment entity and appointment-related data
- **Responsibilities**:
  - `CreateAppointment` — Create appointment with title, description, and doctor_id (validates doctor existence via Doctor Service)
  - `GetAppointment` — Retrieve appointment by ID
  - `ListAppointments` — Return all stored appointments
  - `UpdateAppointmentStatus` — Update status (new → in_progress → done; done → new is forbidden)
- **Protocol**: gRPC (calls Doctor Service via gRPC)
- **Database**: PostgreSQL (`appointment` database)


## 3. Clean Architecture Structure

Both services follow this strict layered architecture:

```
service/
├── cmd/service-<name>/
│   └── main.go                      # Application entry point and dependency wiring
├── internal/
│   ├── model/                       # Domain entities (no external dependencies)
│   │   ├── doctor.go
│   │   └── appointment.go
│   ├── repository/                  # Data persistence layer
│   │   ├── doctor_repository.go
│   │   └── appointment_repository.go
│   ├── usecase/                     # Business logic layer (interfaces only)
│   │   ├── create_doctor.go
│   │   ├── get_doctor.go
│   │   ├── list_doctor.go
│   │   ├── create_appointment.go
│   │   ├── get_appointment.go
│   │   ├── list_appointment.go
│   │   └── update_status.go
│   ├── transport/grpc/              # Delivery layer (gRPC handlers)
│   │   └── doctor_server.go
│   │   └── appointment_server.go
│   └── client/                      # External service clients (Appointment only)
│       └── doctor_grpc_client.go
└── proto/
    ├── doctor.proto
    ├── doctor.pb.go
    └── doctor_grpc.pb.go
```

**Dependency Rule:** Dependencies only point inward.  
`transport/grpc` → `usecase` → `repository` → `model`

**Critical Constraint:** 
- ✅ Domain models (`model/`) must NOT import protobuf-generated types
- ✅ Use cases must NOT import protobuf-generated types
- ✅ Protobuf types belong ONLY in the `transport/grpc/` layer
- ✅ Mapping between proto messages and domain models is the responsibility of the delivery layer


## 4. gRPC Service Contracts

### **Doctor Service** (`doctor/proto/doctor.proto`)

```protobuf
service DoctorService {
  rpc CreateDoctor(CreateDoctorRequest) returns (DoctorResponse);
  rpc GetDoctor(GetDoctorRequest) returns (DoctorResponse);
  rpc ListDoctors(ListDoctorsRequest) returns (ListDoctorsResponse);
}

message CreateDoctorRequest {
  string full_name = 1;
  string specialization = 2;
  string email = 3;
}

message GetDoctorRequest {
  string id = 1;
}

message ListDoctorsRequest {}

message DoctorResponse {
  string id = 1;
  string full_name = 2;
  string specialization = 3;
  string email = 4;
}

message ListDoctorsResponse {
  repeated DoctorResponse doctors = 1;
}
```

### **Appointment Service** (`appointment/proto/appointment.proto`)

```protobuf
service AppointmentService {
  rpc CreateAppointment(CreateAppointmentRequest) returns (AppointmentResponse);
  rpc GetAppointment(GetAppointmentRequest) returns (AppointmentResponse);
  rpc ListAppointments(ListAppointmentsRequest) returns (ListAppointmentsResponse);
  rpc UpdateAppointmentStatus(UpdateStatusRequest) returns (AppointmentResponse);
}

message CreateAppointmentRequest {
  string title = 1;
  string description = 2;
  string doctor_id = 3;
}

message GetAppointmentRequest {
  string id = 1;
}

message ListAppointmentsRequest {}

message UpdateStatusRequest {
  string id = 1;
  string status = 2;
}

message AppointmentResponse {
  string id = 1;
  string title = 2;
  string description = 3;
  string doctor_id = 4;
  string status = 5;
  string created_at = 6;
  string updated_at = 7;
}

message ListAppointmentsResponse {
  repeated AppointmentResponse appointments = 1;
}
```

---

## 5. Inter-Service Communication (gRPC)

**Appointment Service → Doctor Service Call Flow:**

1. Client calls `AppointmentService.CreateAppointment(title, doctor_id, ...)`
2. gRPC handler unpacks the proto message
3. Use case `CreateAppointmentUseCase.Execute()` is invoked
4. Use case calls `DoctorServiceClient.CheckDoctorExists(doctor_id)` (injected interface)
5. Client implementation calls `DoctorService.GetDoctor()` via gRPC
6. If doctor exists (returns `DoctorResponse`), appointment is created
7. If doctor does NOT exist (gRPC returns `codes.NotFound`), use case returns error
8. gRPC handler maps error to appropriate status code and returns to client

**Key Design Decision:**
- The gRPC client is **injected as an interface** into the use case
- The use case NEVER imports protobuf types from doctor service
- Mapping happens in the delivery layer only

---

## 6. gRPC Error Handling

All errors use **standard gRPC status codes** from `google.golang.org/grpc/codes`:

| Scenario | gRPC Status Code | Example |
|----------|------------------|---------|
| Required field missing | `codes.InvalidArgument` | "Title and doctor_id are required" |
| Email already in use | `codes.AlreadyExists` | "Email already in use" |
| Doctor ID not found (local) | `codes.NotFound` | "Doctor not found" |
| Doctor Service unreachable | `codes.Unavailable` | "Doctor service is unavailable" |
| Doctor does not exist (remote check) | `codes.FailedPrecondition` | "Doctor does not exist" |
| Invalid status transition (done→new) | `codes.InvalidArgument` | "Cannot revert from done to new" |
| Internal error | `codes.Internal` | "Failed to create appointment" |

---

## 7. Installation and Setup

### **Prerequisites**
- **Go 1.20+** (project uses Go 1.25.0)
- **PostgreSQL** (tested on PostgreSQL 14+)
- **protoc** (Protocol Buffer compiler)
- **Go gRPC plugins**

### **Install protoc and Go Plugins**

**Windows (using Chocolatey):**
```bash
choco install protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

**macOS (using Homebrew):**
```bash
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

**Linux (Ubuntu/Debian):**
```bash
apt-get install -y protobuf-compiler
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### **Verify Installation**
```bash
protoc --version
which protoc-gen-go
which protoc-gen-go-grpc
```

---

## 8. Regenerating Proto Stubs

If you modify `.proto` files, regenerate the Go stubs:

**For Doctor Service:**
```bash
cd doctor/proto
protoc --go_out=. --go-grpc_out=. doctor.proto
```

**For Appointment Service:**
```bash
cd appointment/proto
protoc --go_out=. --go-grpc_out=. appointment.proto
```

---

## 9. Running the Services Locally

### **Database Setup**

Create two PostgreSQL databases:

```sql
-- Connect to PostgreSQL as admin
psql -U postgres

-- Create doctor database
CREATE DATABASE doctor;

-- Create appointment database
CREATE DATABASE appointment;

-- Create tables
\c doctor
CREATE TABLE doctors (
  id VARCHAR(36) PRIMARY KEY,
  full_name VARCHAR(255) NOT NULL,
  specialization VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

\c appointment
CREATE TABLE appointments (
  id VARCHAR(36) PRIMARY KEY,
  doctor_id VARCHAR(36) NOT NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  status VARCHAR(50) DEFAULT 'new',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### **Start Services**

**Terminal 1 - Doctor Service (gRPC on :50051):**
```bash
cd doctor
go run cmd/service-doctor/main.go
# Output: Doctor Service is running on port :50051 (gRPC)
```

**Terminal 2 - Appointment Service (gRPC on :50052):**
```bash
cd appointment
go run cmd/service-appointment/main.go
# Output: Appointment Service is running on port :50052 (gRPC)
```

Both services will attempt to connect to PostgreSQL on `localhost:5432` with credentials `postgres:12345678`.

---

## 10. Testing 



<img width="888" height="820" alt="image" src="https://github.com/user-attachments/assets/d3b2d04c-d2fe-4a7a-bb7d-0649117cd8d2" />
<img width="892" height="607" alt="image" src="https://github.com/user-attachments/assets/14500751-74ea-4c43-9d36-bc75896af5d7" />
<img width="860" height="569" alt="image" src="https://github.com/user-attachments/assets/efced08b-fb67-4267-af42-a1ffe2255566" />

<img width="873" height="830" alt="image" src="https://github.com/user-attachments/assets/fd2e0ec1-1648-4afa-96bd-e714728a29cb" />
<img width="894" height="655" alt="image" src="https://github.com/user-attachments/assets/e83a80bf-4973-4efc-b4ad-c83aa3987428" />
<img width="879" height="629" alt="image" src="https://github.com/user-attachments/assets/43770b4d-523b-4345-9755-a308fffdc146" />
<img width="899" height="635" alt="image" src="https://github.com/user-attachments/assets/849c9d21-d16e-4afb-b26d-7d21cfb44faa" />



## 11. Failure Scenarios and Resilience

### **Scenario 1: Doctor Service Unreachable**

**What happens:**
1. Appointment Service starts but cannot connect to Doctor Service on `:50051`
2. Main function logs: `"Failed to connect to Doctor Service: ..."`
3. Appointment Service startup fails with exit code 1

**How to test:**
- Start Appointment Service without starting Doctor Service
- Observe error in logs

**gRPC Status Code Returned:** `codes.Unavailable`

### **Scenario 2: Doctor Does Not Exist (Valid Doctor Service)**

**What happens:**
1. Client calls `CreateAppointment` with non-existent `doctor_id`
2. Appointment Service calls `Doctor Service.GetDoctor(doctor_id)`
3. Doctor Service returns `codes.NotFound`
4. Appointment Service catches the error and returns `codes.FailedPrecondition`

**Example error response:**
```json
{
  "code": 9,
  "message": "Doctor does not exist"
}
```

**How to test:**
```bash
grpcurl -plaintext -d '{"title":"Test","description":"Test","doctor_id":"invalid-id"}' localhost:50052 appointment.AppointmentService/CreateAppointment
```

### **Scenario 3: Email Already Exists**

**What happens:**
1. Client calls `CreateDoctor` with duplicate email
2. Repository detects unique constraint violation
3. gRPC handler catches error and returns `codes.AlreadyExists`

**Example error response:**
```json
{
  "code": 6,
  "message": "Email already in use"
}
```

### **Scenario 4: Invalid Status Transition**

**What happens:**
1. Appointment status is `done`
2. Client calls `UpdateAppointmentStatus` with `status="new"`
3. Use case rejects transition: "cannot revert status from 'done' to 'new'"
4. gRPC handler returns `codes.InvalidArgument`

**Example error response:**
```json
{
  "code": 3,
  "message": "Cannot revert from done to new"
}
```

---

## 12. REST vs gRPC: Trade-offs and When to Use Each

### **Trade-off 1: Protocol Efficiency**

| Aspect | REST | gRPC |
|--------|------|------|
| **Encoding** | JSON (text-based) | Protocol Buffers (binary) |
| **Message Size** | Large (human-readable) | Small (compact binary) |
| **Network Bandwidth** | High | Low (-70% in practice) |
| **Best For** | Public APIs, web browsers | Microservice-to-microservice |

**When to use:**
- Use **REST** for client-facing APIs (mobile apps, web frontends)
- Use **gRPC** for internal service communication (faster, smaller messages)

### **Trade-off 2: Development Speed vs Type Safety**

| Aspect | REST | gRPC |
|--------|------|------|
| **Type Definition** | Loose (JSON can be anything) | Strict (Proto contracts) |
| **Code Generation** | Manual (or tools) | Automatic (protoc) |
| **Breaking Changes** | Easier to ignore versioning | Enforced by proto evolution rules |
| **Best For** | Rapid prototyping | Production systems with stability guarantees |

**When to use:**
- Use **REST** for quick iterations and experimental APIs
- Use **gRPC** when you need contract consistency across teams

### **Trade-off 3: Client Ecosystem and Tooling**

| Aspect | REST | gRPC |
|--------|------|------|
| **Browser Support** | Native (fetch, XMLHttpRequest) | No (requires gRPC-web for browsers) |
| **Debugging Tools** | cURL, Postman, browser DevTools | grpcurl, Postman (limited), custom tools |
| **Load Balancing** | Standard HTTP load balancers | Requires aware load balancers |
| **Best For** | Web APIs with diverse clients | Polyglot microservices (Go, Java, Python, etc.) |

**When to use:**
- Use **REST** when clients are browsers or mobile apps with limited gRPC support
- Use **gRPC** when all clients are microservices you control (Go, Node, Python, Java, etc.)

### **Project Decision: Why gRPC?**

This project **migrated to gRPC** because:
1. **Both clients are microservices** (Appointment ↔ Doctor service communication)
2. **High performance** needed for medical scheduling platform (low latency)
3. **Type safety** ensures appointment/doctor contracts don't break
4. **No browser clients** accessing internal service-to-service APIs
5. **Polyglot ready** — future services can be written in any language supporting gRPC


## 13. Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client / CLI                              │
│                     (grpcurl / test client)                     │
└───────────────────┬──────────────────────────────────────────────┘
                    │
                    │ (gRPC requests)
                    │
        ┌───────────┴────────────┬───────────────┬────────────────┐
        │                        │               │                │
   ┌────▼──────────────┐  ┌──────▼──────┐   ┌───▼───────┐    ┌───▼───────┐
   │  DOCTOR SERVICE   │  │   PORT      │   │  DATABASE │    │   PROTO   │
   │   (gRPC Server)   │  │   :50051    │   │ doctor DB │    │  contracts│
   └────┬──────────────┘  └─────────────┘   └───────────┘    └───────────┘
        │
        ├── CreateDoctor(CreateDoctorRequest)
        ├── GetDoctor(GetDoctorRequest) → GetDoctor status
        └── ListDoctors(ListDoctorsRequest)
        
        ▲
        │ (gRPC client call)
        │ from Appointment Service
        │
   ┌────┴──────────────┐  ┌──────────────┐   ┌───────────┐    ┌───────────┐
   │APPOINTMENT SERVICE│  │   PORT       │   │ DATABASE  │    │    PROTO  │
   │   (gRPC Server)   │  │   :50052     │   │appoint DB │    │ contracts │
   └────┬──────────────┘  └──────────────┘   └───────────┘    └───────────┘
        │
        ├── CreateAppointment(CreateAppointmentRequest)
        │   └─→ Calls Doctor Service.GetDoctor(doctor_id)
        │       └─→ Validates doctor exists
        │           └─→ If doctor_id invalid → codes.FailedPrecondition
        │
        ├── GetAppointment(GetAppointmentRequest)
        ├── ListAppointments(ListAppointmentsRequest)
        └── UpdateAppointmentStatus(UpdateStatusRequest)
            └─→ Validates status transition
                └─→ If done→new → codes.InvalidArgument
```


## 14. Project Structure Summary

```
ap2-assignment2/
├── doctor/
│   ├── cmd/service-doctor/
│   │   └── main.go                          # Server startup, DB init
│   ├── internal/
│   │   ├── model/doctor.go                  # Domain entity
│   │   ├── repository/doctor_repository.go  # PostgreSQL impl
│   │   ├── usecase/
│   │   │   ├── create_doctor.go
│   │   │   ├── get_doctor.go
│   │   │   └── list_doctor.go
│   │   └── transport/grpc/doctor_server.go  # gRPC handlers
│   ├── proto/
│   │   ├── doctor.proto                     # Proto contract
│   │   ├── doctor.pb.go                     # Generated code
│   │   └── doctor_grpc.pb.go                # Generated code
│   ├── go.mod
│   └── go.sum
│
├── appointment/
│   ├── cmd/service-appointment/
│   │   └── main.go                          # Server startup, DB init
│   ├── internal/
│   │   ├── model/appointment.go             # Domain entity
│   │   ├── repository/appointment_repository.go  # PostgreSQL impl
│   │   ├── usecase/
│   │   │   ├── create_appointment.go
│   │   │   ├── get_appointment.go
│   │   │   ├── list_appointment.go
│   │   │   └── update_status.go
│   │   ├── client/doctor_grpc_client.go     # Doctor Service client
│   │   └── transport/grpc/appointment_server.go  # gRPC handlers
│   ├── proto/
│   │   ├── appointment.proto                # Proto contract
│   │   ├── appointment.pb.go                # Generated code
│   │   └── appointment_grpc.pb.go           # Generated code
│   ├── go.mod
│   └── go.sum
│
└── README.md                                # This file
```


## 15. Summary of gRPC Benefits in This Project

✅ **Type-Safe Contracts**: Proto files enforce data shapes between services  
✅ **Efficient Communication**: Binary protocol reduces network overhead  
✅ **Language Agnostic**: Can add services in Python, Java, etc. without changes  
✅ **Built-in Error Handling**: Standard gRPC codes for all error scenarios  
✅ **Method Generation**: protoc generates all server/client stubs automatically  

