package usecase

import (
	"context"
	"log"
	"time"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/repository"
)

type DoctorServiceClient interface {
	CheckDoctorExists(ctx context.Context, doctorID string) (bool, error)
}

type CreateAppointmentUseCase interface {
	Execute(ctx context.Context, appointment *model.Appointment) error
}

type CreateAppointmentInteractor struct {
	repo         repository.AppointmentRepository
	doctorClient DoctorServiceClient
	pub          Publisher
}

func NewCreateAppointmentUseCase(repo repository.AppointmentRepository, dc DoctorServiceClient, pub Publisher) *CreateAppointmentInteractor {
	return &CreateAppointmentInteractor{repo: repo, doctorClient: dc, pub: pub}
}

func (uc *CreateAppointmentInteractor) Execute(ctx context.Context, appointment *model.Appointment) error {
	if appointment.Title == "" || appointment.DoctorID == "" {
		return ErrRequiredFields
	}

	exists, err := uc.doctorClient.CheckDoctorExists(ctx, appointment.DoctorID)
	if err != nil {
		return ErrDoctorServiceUnavailable
	}
	if !exists {
		return ErrDoctorNotFound
	}

	appointment.Status = model.New
	appointment.CreatedAt = time.Now()
	appointment.UpdatedAt = time.Now()

	if err := uc.repo.Create(ctx, appointment); err != nil {
		return err
	}

	if err := uc.pub.Publish(ctx, "appointments.created", map[string]any{
		"event_type":  "appointments.created",
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
		"id":          appointment.ID,
		"title":       appointment.Title,
		"doctor_id":   appointment.DoctorID,
		"status":      string(appointment.Status),
	}); err != nil {
		log.Printf("publish appointments.created failed id=%s: %v", appointment.ID, err)
	}

	return nil
}
