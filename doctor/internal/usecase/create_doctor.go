package usecase

import (
	"context"
	"errors"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/repository"
)

var (
	ErrEmptyFields        = errors.New("full name, specialization, and email are required")
	ErrEmailAlreadyExists = errors.New("a doctor with this email already exists")
)

type CreateDoctorUseCase interface {
	Execute(ctx context.Context, doctor *model.Doctor) error
}

type createDoctorInteractor struct {
	repo repository.DoctorRepository
}

func NewCreateDoctorUseCase(repo repository.DoctorRepository) CreateDoctorUseCase {
	return &createDoctorInteractor{
		repo: repo,
	}
}

func (i *createDoctorInteractor) Execute(ctx context.Context, doctor *model.Doctor) error {
	if doctor.FullName == "" || doctor.Specialization == "" || doctor.Email == "" {
		return ErrEmptyFields
	}

	existingDoctor, _ := i.repo.GetByEmail(ctx, doctor.Email)
	if existingDoctor != nil {
		return ErrEmailAlreadyExists
	}

	return i.repo.CreateDoctor(ctx, doctor)

}
