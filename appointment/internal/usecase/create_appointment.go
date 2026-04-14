package usecase

import (
	"context"
	"time"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/repository"
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
}

func NewCreateAppointmentUseCase(repo repository.AppointmentRepository, dc DoctorServiceClient) *CreateAppointmentInteractor {
	return &CreateAppointmentInteractor{repo: repo, doctorClient: dc}
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

	return uc.repo.Create(ctx, appointment)
}
