package main

import (
	"database/sql"
	"log"
	"net"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/config"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/repository"
	grpctransport "github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/transport/grpc"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/usecase"
	proto "github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/proto"
)

func main() {
	cfg := config.LoadDoctorConfig()

	db, err := sql.Open("pgx", cfg.DatabaseURL)
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

	listener, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer()
	doctorServer := grpctransport.NewDoctorServer(createUC, getByIdUC, getAllUC)
	proto.RegisterDoctorServiceServer(grpcServer, doctorServer)

	log.Printf("Doctor Service is running on port %s (gRPC)", cfg.GRPCPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
