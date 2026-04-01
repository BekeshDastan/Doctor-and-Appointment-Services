package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/repository"
	httptransport "github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/transport/http"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/usecase"
)

func main() {
	dsn := "postgres://postgres:12345678@localhost:5432/appointment?sslmode=disable"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL!")

	doctorClient := usecase.NewRESTDoctorClient("http://localhost:8080")

	appointmentRepo := repository.NewPostgresAppointmentRepository(db)

	createUC := usecase.NewCreateAppointmentUseCase(appointmentRepo, doctorClient)
	getByIdUC := usecase.NewGetAppointmentUseCase(appointmentRepo)
	getAllUC := usecase.NewListAppointmentsUseCase(appointmentRepo)
	updateStatusUC := usecase.NewUpdateStatusUseCase(appointmentRepo)

	router := gin.Default()

	appointmentHandler := httptransport.NewAppointmentHandler(
		createUC,
		getByIdUC,
		getAllUC,
		updateStatusUC,
	)

	appointmentHandler.RegisterRoutes(router)

	log.Println("Appointment Service is running on port :8081")
	if err := router.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
