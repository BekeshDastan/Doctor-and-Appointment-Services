package usecase

import (
	"context"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor-service/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor-service/internal/repository"
)

type GetAllDoctorsUseCase interface {
	Execute(ctx context.Context) ([]*model.Doctor, error)
}

type getAllDoctorsUseCase struct {
	repo repository.DoctorRepository
}

func NewGetAllDoctorsUseCase(repo repository.DoctorRepository) GetAllDoctorsUseCase {
	return &getAllDoctorsUseCase{repo: repo}
}

func (uc *getAllDoctorsUseCase) Execute(ctx context.Context) ([]*model.Doctor, error) {
	return uc.repo.GetAll(ctx)
}
