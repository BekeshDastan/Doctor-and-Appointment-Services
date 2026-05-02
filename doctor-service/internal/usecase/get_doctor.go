package usecase

import (
	"context"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor-service/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor-service/internal/repository"
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
		return nil, ErrEmptyID
	}
	doctor, err := uc.repo.GetDoctorById(ctx, id)
	if err != nil {
		if err.Error() == "doctor not found" {
			return nil, ErrDoctorNotFound
		}
		return nil, err
	}
	return doctor, nil
}
