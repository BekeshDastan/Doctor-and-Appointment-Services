package config

import (
	"fmt"
	"os"
)

type AppointmentConfig struct {
	DatabaseURL      string
	GRPCPort         string
	DoctorServiceURL string
}

func LoadAppointmentConfig() *AppointmentConfig {
	dbURL := os.Getenv("APPOINTMENT_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:12345678@localhost:5432/appointment?sslmode=disable"
	}

	port := os.Getenv("APPOINTMENT_GRPC_PORT")
	if port == "" {
		port = "50052"
	}

	doctorURL := os.Getenv("DOCTOR_SERVICE_URL")
	if doctorURL == "" {
		doctorURL = "localhost:50051"
	}

	return &AppointmentConfig{
		DatabaseURL:      dbURL,
		GRPCPort:         fmt.Sprintf(":%s", port),
		DoctorServiceURL: doctorURL,
	}
}
