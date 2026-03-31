package model

import "time"

type Appointment struct {
	ID          string
	Title       string
	Description string
	DoctorID    string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Status string

const (
	New         Status = "new"
	In_Progress Status = "in_progress"
	Done        Status = "done"
)
