package usecase

import (
	"context"
	"log"
	"time"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/event"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/repository"
)

type UpdateStatusUseCase struct {
	repo repository.AppointmentRepository
	pub  event.Publisher
}

func NewUpdateStatusUseCase(repo repository.AppointmentRepository, pub event.Publisher) *UpdateStatusUseCase {
	return &UpdateStatusUseCase{repo: repo, pub: pub}
}

func (uc *UpdateStatusUseCase) Execute(ctx context.Context, id string, newStatus model.Status) error {
	appt, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if appt.Status == model.Done && newStatus == model.New {
		return ErrStatusTransition
	}

	if newStatus != model.New && newStatus != model.In_Progress && newStatus != model.Done {
		return ErrInvalidStatus
	}

	oldStatus := appt.Status
	appt.Status = newStatus
	appt.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, appt); err != nil {
		return err
	}

	if err := uc.pub.Publish(ctx, "appointments.status_updated", map[string]any{
		"event_type":  "appointments.status_updated",
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
		"id":          id,
		"old_status":  string(oldStatus),
		"new_status":  string(newStatus),
	}); err != nil {
		log.Printf("publish appointments.status_updated failed id=%s: %v", id, err)
	}

	return nil
}
