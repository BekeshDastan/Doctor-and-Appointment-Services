package usecase

import (
	"context"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/repository"
)

type GetDoctorByIdUseCase interface {
	Execute(ctx context.Context, id string) (*model.Doctor, error)
}

type getDoctorByIdUseCase struct {
	repo repository.DoctorRepository
}

func NewGetDoctorByIdUseCase(repo repository.DoctorRepository) GetDoctorByIdUseCase {
	return &getDoctorByIdUseCase{repo: repo}
}

func (uc *getDoctorByIdUseCase) Execute(ctx context.Context, id string) (*model.Doctor, error) {
	if id == "" {
		return nil, ErrEmptyFields
	}
	return uc.repo.GetDoctorById(ctx, id)
}
