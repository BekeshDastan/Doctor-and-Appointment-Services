package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment/internal/repository"
)

type UpdateStatusUseCase struct {
	repo repository.AppointmentRepository
}

func NewUpdateStatusUseCase(repo repository.AppointmentRepository) *UpdateStatusUseCase {
	return &UpdateStatusUseCase{repo: repo}
}

func (uc *UpdateStatusUseCase) Execute(ctx context.Context, id string, newStatus model.Status) error {
	appt, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if appt.Status == model.Done && newStatus == model.New {
		return errors.New("cannot revert status from 'done' to 'new'")
	}

	if newStatus != model.New && newStatus != model.In_Progress && newStatus != model.Done {
		return errors.New("invalid status")
	}

	appt.Status = newStatus
	appt.UpdatedAt = time.Now()

	return uc.repo.Update(ctx, appt)
}
