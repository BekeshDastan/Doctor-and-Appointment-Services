package usecase

import (
	"context"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/repository"
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
