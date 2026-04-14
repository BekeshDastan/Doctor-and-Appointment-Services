package main

import (
	"database/sql"
	"log"
	"net"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/client"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/config"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/repository"
	grpctransport "github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/transport/grpc"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/usecase"
	proto "github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/proto"
)

func main() {
	cfg := config.LoadAppointmentConfig()

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL!")

	// Connect to Doctor Service via gRPC
	doctorConn, err := grpc.Dial(cfg.DoctorServiceURL, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Doctor Service at %s: %v", cfg.DoctorServiceURL, err)
	}
	defer doctorConn.Close()

	doctorClient := client.NewGRPCDoctorClient(doctorConn)

	appointmentRepo := repository.NewPostgresAppointmentRepository(db)

	createUC := usecase.NewCreateAppointmentUseCase(appointmentRepo, doctorClient)
	getByIdUC := usecase.NewGetAppointmentUseCase(appointmentRepo)
	getAllUC := usecase.NewListAppointmentsUseCase(appointmentRepo)
	updateStatusUC := usecase.NewUpdateStatusUseCase(appointmentRepo)

	// Create gRPC listener
	listener, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	appointmentServer := grpctransport.NewAppointmentServer(
		createUC,
		getByIdUC,
		getAllUC,
		updateStatusUC,
		doctorClient,
	)
	proto.RegisterAppointmentServiceServer(grpcServer, appointmentServer)

	log.Printf("Appointment Service is running on port %s (gRPC)", cfg.GRPCPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
