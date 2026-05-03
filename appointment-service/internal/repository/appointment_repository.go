package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/model"
)

var ErrNotFound = errors.New("appointment not found")

type AppointmentRepository interface {
	Create(ctx context.Context, appt *model.Appointment) error
	GetByID(ctx context.Context, id string) (*model.Appointment, error)
	List(ctx context.Context) ([]*model.Appointment, error)
	Update(ctx context.Context, appt *model.Appointment) error
}

type postgresAppointmentRepository struct {
	db *sql.DB
}

func NewPostgresAppointmentRepository(db *sql.DB) AppointmentRepository {
	return &postgresAppointmentRepository{db: db}
}

func (r *postgresAppointmentRepository) Create(ctx context.Context, appt *model.Appointment) error {
	query := `
		INSERT INTO appointments (id, doctor_id, Description, title, status, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		appt.ID,
		appt.DoctorID,
		appt.Description,
		appt.Title,
		appt.Status,
		appt.CreatedAt,
		appt.UpdatedAt,
	)
	return err
}

func (r *postgresAppointmentRepository) GetByID(ctx context.Context, id string) (*model.Appointment, error) {
	query := `
		SELECT id, doctor_id, Description, title, status, created_at, updated_at 
		FROM appointments 
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var appt model.Appointment
	err := row.Scan(
		&appt.ID,
		&appt.DoctorID,
		&appt.Description,
		&appt.Title,
		&appt.Status,
		&appt.CreatedAt,
		&appt.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &appt, nil
}

func (r *postgresAppointmentRepository) List(ctx context.Context) ([]*model.Appointment, error) {
	query := `
		SELECT id, doctor_id, Description, title, status, created_at, updated_at 
		FROM appointments
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []*model.Appointment
	for rows.Next() {
		var appt model.Appointment
		err := rows.Scan(
			&appt.ID,
			&appt.DoctorID,
			&appt.Description,
			&appt.Title,
			&appt.Status,
			&appt.CreatedAt,
			&appt.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		appointments = append(appointments, &appt)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return appointments, nil
}

func (r *postgresAppointmentRepository) Update(ctx context.Context, appt *model.Appointment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`UPDATE appointments SET status = $1, updated_at = $2 WHERE id = $3`,
		appt.Status, appt.UpdatedAt, appt.ID)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return tx.Commit()
}
