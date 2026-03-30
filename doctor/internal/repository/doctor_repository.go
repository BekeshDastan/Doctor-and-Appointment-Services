package repository

import (
	"context"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/model"
)

type DoctorRepository interface {
	CreateDoctor(ctx context.Context, doctor *model.Doctor) error
	GetDoctorById(ctx context.Context, id string) (*model.Doctor, error)
	GetAll(ctx context.Context) ([]*model.Doctor, error)
	GetByEmail(ctx context.Context, email string) (*model.Doctor, error)
}
