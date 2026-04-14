package usecase

import "errors"

// Custom error types for Appointment Service
var (
	ErrAptNotFound              = errors.New("appointment not found")
	ErrInvalidStatus            = errors.New("invalid status")
	ErrStatusTransition         = errors.New("cannot revert status from 'done' to 'new'")
	ErrRequiredFields           = errors.New("appointment title and doctor id are required")
	ErrDoctorNotFound           = errors.New("doctor does not exist")
	ErrDoctorServiceUnavailable = errors.New("doctor service is unavailable")
)
