package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor/internal/model"
)

type postgresDoctorRepository struct {
	db *sql.DB
}

func NewPostgresDoctorRepository(db *sql.DB) DoctorRepository {
	return &postgresDoctorRepository{db: db}
}

func (r *postgresDoctorRepository) CreateDoctor(ctx context.Context, doctor *model.Doctor) error {
	query := `
		INSERT INTO doctors (id, full_name, specialization, email) 
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, doctor.ID, doctor.FullName, doctor.Specialization, doctor.Email)
	return err
}

func (r *postgresDoctorRepository) GetDoctorById(ctx context.Context, id string) (*model.Doctor, error) {
	query := `
		SELECT id, full_name, specialization, email 
		FROM doctors 
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var doc model.Doctor
	err := row.Scan(&doc.ID, &doc.FullName, &doc.Specialization, &doc.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("doctor not found") // Здесь лучше возвращать кастомную ошибку домена
		}
		return nil, err
	}

	return &doc, nil
}

func (r *postgresDoctorRepository) GetAll(ctx context.Context) ([]*model.Doctor, error) {
	query := `
		SELECT id, full_name, specialization, email 
		FROM doctors
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var doctors []*model.Doctor
	for rows.Next() {
		var doc model.Doctor
		if err := rows.Scan(&doc.ID, &doc.FullName, &doc.Specialization, &doc.Email); err != nil {
			return nil, err
		}
		doctors = append(doctors, &doc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return doctors, nil
}

func (r *postgresDoctorRepository) GetByEmail(ctx context.Context, email string) (*model.Doctor, error) {
	query := `
		SELECT id, full_name, specialization, email 
		FROM doctors 
		WHERE email = $1
	`

	row := r.db.QueryRowContext(ctx, query, email)

	var doc model.Doctor
	err := row.Scan(&doc.ID, &doc.FullName, &doc.Specialization, &doc.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("doctor not found")
		}
		return nil, err
	}

	return &doc, nil
}
