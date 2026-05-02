package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/notification-service/internal/subscriber"
)

func main() {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = "nats://localhost:4222"
	}

	nc, err := subscriber.ConnectWithBackoff(url, 5)
	if err != nil {
		log.Fatalf("Failed to connect to broker: %v", err)
	}
	defer nc.Drain()

	subscriber.Subscribe(nc, []string{
		"doctors.created",
		"appointments.created",
		"appointments.status_updated",
	})

	log.Println("Notification Service is running — waiting for events...")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down Notification Service")
}
