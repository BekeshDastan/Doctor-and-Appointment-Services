package usecase

import "errors"

// Custom error types for Doctor Service
var (
	ErrEmptyFields        = errors.New("full name, specialization, and email are required")
	ErrEmailAlreadyExists = errors.New("a doctor with this email already exists")
	ErrDoctorNotFound     = errors.New("doctor not found")
	ErrEmptyID            = errors.New("doctor id cannot be empty")
)
