package usecase

import (
	"context"
	"log"
	"time"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor-service/internal/event"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor-service/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor-service/internal/repository"
)

type CreateDoctorRequest struct {
	FullName       string `json:"full_name" binding:"required"`
	Specialization string `json:"specialization" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
}

type CreateDoctorUseCase interface {
	Execute(ctx context.Context, doctor *model.Doctor) error
}

type createDoctorInteractor struct {
	repo repository.DoctorRepository
	pub  event.Publisher
}

func NewCreateDoctorUseCase(repo repository.DoctorRepository, pub event.Publisher) CreateDoctorUseCase {
	return &createDoctorInteractor{repo: repo, pub: pub}
}

func (i *createDoctorInteractor) Execute(ctx context.Context, doctor *model.Doctor) error {
	if doctor.FullName == "" || doctor.Specialization == "" || doctor.Email == "" {
		return ErrEmptyFields
	}

	if err := i.repo.CreateWithEmailCheck(ctx, doctor); err != nil {
		if err == repository.ErrEmailAlreadyExists {
			return ErrEmailAlreadyExists
		}
		return err
	}

	if err := i.pub.Publish(ctx, "doctors.created", map[string]any{
		"event_type":     "doctors.created",
		"occurred_at":    time.Now().UTC().Format(time.RFC3339),
		"id":             doctor.ID,
		"full_name":      doctor.FullName,
		"specialization": doctor.Specialization,
		"email":          doctor.Email,
	}); err != nil {
		log.Printf("publish doctors.created failed id=%s: %v", doctor.ID, err)
	}

	return nil
}
