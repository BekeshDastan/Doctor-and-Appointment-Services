package usecase

import (
	"context"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/repository"
)

type GetAppointmentUseCase struct {
	repo repository.AppointmentRepository
}

func NewGetAppointmentUseCase(repo repository.AppointmentRepository) *GetAppointmentUseCase {
	return &GetAppointmentUseCase{repo: repo}
}

func (uc *GetAppointmentUseCase) Execute(ctx context.Context, id string) (*model.Appointment, error) {
	return uc.repo.GetByID(ctx, id)
}
