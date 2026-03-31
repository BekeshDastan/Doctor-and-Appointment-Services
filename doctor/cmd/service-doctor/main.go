package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/repository"
	httptransport "github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/transport/http"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/usecase"
)

func main() {
	dsn := "postgres://postgres:12345678@localhost:5432/doctor?sslmode=disable"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL!")

	doctorRepo := repository.NewPostgresDoctorRepository(db)

	createUC := usecase.NewCreateDoctorUseCase(doctorRepo)
	getByIdUC := usecase.NewGetDoctorByIdUseCase(doctorRepo)
	getAllUC := usecase.NewGetAllDoctorsUseCase(doctorRepo)

	router := gin.Default()

	doctorHandler := httptransport.NewDoctorHandler(createUC, getByIdUC, getAllUC)

	doctorHandler.RegisterRoutes(router)

	log.Println("Doctor Service is running on port :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
