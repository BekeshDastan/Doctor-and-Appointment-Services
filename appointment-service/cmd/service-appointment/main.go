package main

import (
	"database/sql"
	"log"
	"net"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/app"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/client"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/config"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/event"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/repository"
	grpctransport "github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/transport/grpc"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/usecase"
	proto "github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/proto"
)

func main() {
	cfg := config.LoadAppointmentConfig()

	if err := app.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations applied successfully")

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL!")

	doctorConn, err := grpc.Dial(cfg.DoctorServiceURL, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Doctor Service at %s: %v", cfg.DoctorServiceURL, err)
	}
	defer doctorConn.Close()

	pub := event.NewPublisher(cfg.NATSURL)

	doctorClient := client.NewGRPCDoctorClient(doctorConn)
	appointmentRepo := repository.NewPostgresAppointmentRepository(db)

	createUC := usecase.NewCreateAppointmentUseCase(appointmentRepo, doctorClient, pub)
	getByIdUC := usecase.NewGetAppointmentUseCase(appointmentRepo)
	getAllUC := usecase.NewListAppointmentsUseCase(appointmentRepo)
	updateStatusUC := usecase.NewUpdateStatusUseCase(appointmentRepo, pub)

	listener, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

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
