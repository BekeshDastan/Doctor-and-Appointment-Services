# Doctor and Appointment Services Platform

## 1. Project Overview and Purpose
This project implements a medical scheduling platform consisting of two independent microservices: **Doctor Service** and **Appointment Service**. The system is built using Go, Gin (for HTTP routing), and follows **Clean Architecture** principles. 

The primary purpose of this project is to demonstrate service decomposition, bounded contexts, separate data ownership, and synchronous inter-service communication via REST, while avoiding the pitfalls of a distributed monolith.

## 2. Service Responsibilities
* **Doctor Service**: Operates within its own bounded context to manage doctor profiles. It is solely responsible for creating doctors, enforcing email uniqueness, and providing doctor details to external consumers.
* **Appointment Service**: Manages patient appointments and their statuses (new, in_progress, done). It does not store doctor details directly but relies on the Doctor Service via REST to validate a doctor's existence before creating or updating an appointment.

## 3. Folder Structure and Dependency Flow
Both services follow a strict layered structure based on Clean Architecture:

```text
cmd/
  service-name/
    main.go                 # Application entry point and wiring
internal/
  model/                    # Core domain entities (e.g., Doctor, Appointment). No external dependencies.
  repository/               # Data persistence layer. Implements repository interfaces.
  usecase/                  # Business logic layer. Defines interfaces for repositories and external clients.
  transport/http/           # Delivery layer (Gin handlers). Parses JSON and delegates to usecases.

```

**Dependency Rule:** Dependencies only point inward. 
`transport/http` ➔ `usecase` ➔ `repository` ➔ `model`

The core business logic (`usecase`) never depends on specific web frameworks or database drivers, utilizing interfaces instead.

---

## 4. Inter-Service Communication

The services communicate synchronously using REST over HTTP. 
When a client requests to create a new appointment, the **Appointment Service** makes an internal `GET /doctors/{id}` request to the **Doctor Service**.

* **If the doctor exists (200 OK):** the appointment is created.
* **If the doctor does not exist (404 Not Found)** or the service is down:** the Appointment Service rejects the request.



## 5. How to Run the Project

To run the project locally, open two separate terminal windows.

**Start the Doctor Service:**
```bash
cd doctor
go run cmd/service-doctor/main.go
# The Doctor service will start on port 8080 (or your configured port)
```

**Start the Appointment Service:**
```bash
cd appointment
go run cmd/service-appointment/main.go
# The Appointment service will start on port 8081 (or your configured port)
```


## 6. Why a Shared Database Was Not Used (Data Ownership)

Each service manages its own **independent database/storage**. A shared database was intentionally avoided to ensure strict **data ownership** and **bounded contexts**. 

If the `Appointment Service` queried the Doctor tables directly, it would create a tight coupling at the database level (a *"distributed monolith"*). This would make it impossible to independently evolve the Doctor schema or swap its underlying database technology in the future.



## 7. Failure Scenarios and Resilience

**Current Implementation:**

If the `Doctor Service` is unavailable (e.g., due to crashes or a network partition), the Appointment Service's HTTP client will fail to connect. 

The `usecase` layer catches this error and returns a clear `503 Service Unavailable` HTTP status code to the user, gracefully aborting the appointment creation process without corrupting data.
## 8. Diagram
<img width="1241" height="662" alt="image" src="https://github.com/user-attachments/assets/e4290068-64f0-4448-a645-ba38294b73af" />

## 9. POSTMAN
# Doctors
<img width="914" height="696" alt="image" src="https://github.com/user-attachments/assets/8ada49de-7fd8-4f94-9027-49cba95eaf26" />
<img width="884" height="679" alt="image" src="https://github.com/user-attachments/assets/e6bec5d9-801c-4826-ba5a-01a6939f6f28" />
<img width="880" height="807" alt="image" src="https://github.com/user-attachments/assets/8f05780a-9c81-4f76-88e7-be4149718583" />
# Appointments
<img width="927" height="737" alt="image" src="https://github.com/user-attachments/assets/e9628074-b547-46fa-af22-ee6d956f703b" />
<img width="868" height="575" alt="image" src="https://github.com/user-attachments/assets/1c6f7db7-cb08-4148-93a6-0fcfe642b942" />
<img width="889" height="801" alt="image" src="https://github.com/user-attachments/assets/64ba4d2a-c0ae-4f5c-82ab-3433f45a336a" />
<img width="868" height="575" alt="image" src="https://github.com/user-attachments/assets/ca6f0959-a9b3-44a3-bae2-da56cceb3f1b" />
<img width="884" height="695" alt="image" src="https://github.com/user-attachments/assets/3466e39d-7f2a-4fc2-aec2-7a0989f32b51" />
<img width="895" height="737" alt="image" src="https://github.com/user-attachments/assets/dae6d48d-388a-44e6-ac83-e868c2b78e70" />
<img width="914" height="657" alt="image" src="https://github.com/user-attachments/assets/ec6f33d6-4c95-4941-871a-231547358748" />
<img width="865" height="764" alt="image" src="https://github.com/user-attachments/assets/27bdb791-70e4-41df-a68b-c9756a61bcc3" />
<img width="877" height="686" alt="image" src="https://github.com/user-attachments/assets/f1605acc-5bb8-4b71-a7cb-624cb7f1b5bc" />
<img width="885" height="702" alt="image" src="https://github.com/user-attachments/assets/e5c4344b-7db9-44c8-9416-5010e9b72512" />











