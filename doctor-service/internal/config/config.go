package config

import (
	"fmt"
	"os"
)

type DoctorConfig struct {
	DatabaseURL string
	GRPCPort    string
	NATSURL     string
}

func LoadDoctorConfig() *DoctorConfig {
	dbURL := os.Getenv("DOCTOR_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:12345678@localhost:5432/doctor?sslmode=disable"
	}

	port := os.Getenv("DOCTOR_GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	return &DoctorConfig{
		DatabaseURL: dbURL,
		GRPCPort:    fmt.Sprintf(":%s", port),
		NATSURL:     natsURL,
	}
}
