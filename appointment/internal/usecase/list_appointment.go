package usecase

import (
	"context"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/repository"
)

type ListAppointmentsUseCase struct {
	repo repository.AppointmentRepository
}

func NewListAppointmentsUseCase(repo repository.AppointmentRepository) *ListAppointmentsUseCase {
	return &ListAppointmentsUseCase{repo: repo}
}

func (uc *ListAppointmentsUseCase) Execute(ctx context.Context) ([]*model.Appointment, error) {
	return uc.repo.List(ctx)
}
