package usecase

import (
	"context"
	"errors"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/repository"
)

type GetAppointmentUseCase struct {
	repo repository.AppointmentRepository
}

func NewGetAppointmentUseCase(repo repository.AppointmentRepository) *GetAppointmentUseCase {
	return &GetAppointmentUseCase{repo: repo}
}

func (uc *GetAppointmentUseCase) Execute(ctx context.Context, id string) (*model.Appointment, error) {
	apt, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrAptNotFound
		}
		return nil, err
	}
	return apt, nil
}
