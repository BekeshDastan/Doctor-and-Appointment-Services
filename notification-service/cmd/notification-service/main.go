package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/notification-service/internal/subscriber"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	url := os.Getenv("NATS_URL")
	if url == "" {
		url = "nats://localhost:4222"
	}

	nc, err := subscriber.ConnectWithBackoff(url, 5)
	if err != nil {
		slog.Error("failed to connect to broker", "error", err)
		os.Exit(1)
	}
	defer nc.Drain()

	subscriber.Subscribe(nc, []string{
		"doctors.created",
		"appointments.created",
		"appointments.status_updated",
	})

	slog.Info("notification service running", "status", "waiting for events")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	slog.Info("shutting down notification service")
}
