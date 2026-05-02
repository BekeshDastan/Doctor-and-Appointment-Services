package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/client"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/internal/usecase"
	proto "github.com/BekeshDastan/Doctor-and-Appointment-Services/appointment-service/proto"
)

type AppointmentServer struct {
	proto.UnimplementedAppointmentServiceServer
	createUC  usecase.CreateAppointmentUseCase
	getUC     *usecase.GetAppointmentUseCase
	listUC    *usecase.ListAppointmentsUseCase
	updateUC  *usecase.UpdateStatusUseCase
	docClient client.DoctorClient
}

func NewAppointmentServer(
	createUC usecase.CreateAppointmentUseCase,
	getUC *usecase.GetAppointmentUseCase,
	listUC *usecase.ListAppointmentsUseCase,
	updateUC *usecase.UpdateStatusUseCase,
	docClient client.DoctorClient,
) *AppointmentServer {
	return &AppointmentServer{
		createUC:  createUC,
		getUC:     getUC,
		listUC:    listUC,
		updateUC:  updateUC,
		docClient: docClient,
	}
}

func (s *AppointmentServer) CreateAppointment(ctx context.Context, req *proto.CreateAppointmentRequest) (*proto.AppointmentResponse, error) {
	// Create domain model
	appointment := &model.Appointment{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		DoctorID:    req.DoctorId,
		Status:      model.New,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Call usecase (which will verify doctor exists)
	err := s.createUC.Execute(ctx, appointment)
	if err != nil {
		// Handle specific errors
		if errors.Is(err, usecase.ErrRequiredFields) {
			return nil, status.Error(codes.InvalidArgument, "Title and doctor_id are required")
		}
		if errors.Is(err, usecase.ErrDoctorServiceUnavailable) {
			return nil, status.Error(codes.Unavailable, "Doctor service is unavailable")
		}
		if errors.Is(err, usecase.ErrDoctorNotFound) {
			return nil, status.Error(codes.FailedPrecondition, "Doctor does not exist")
		}
		return nil, status.Error(codes.Internal, "Failed to create appointment")
	}

	// Return response
	return s.appointmentToProto(appointment), nil
}

func (s *AppointmentServer) GetAppointment(ctx context.Context, req *proto.GetAppointmentRequest) (*proto.AppointmentResponse, error) {
	appointment, err := s.getUC.Execute(ctx, req.Id)
	if err != nil {
		if errors.Is(err, usecase.ErrAptNotFound) {
			return nil, status.Error(codes.NotFound, "Appointment not found")
		}
		return nil, status.Error(codes.Internal, "Failed to get appointment")
	}

	return s.appointmentToProto(appointment), nil
}

func (s *AppointmentServer) ListAppointments(ctx context.Context, _ *proto.ListAppointmentsRequest) (*proto.ListAppointmentsResponse, error) {
	appointments, err := s.listUC.Execute(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to list appointments")
	}

	resp := &proto.ListAppointmentsResponse{
		Appointments: make([]*proto.AppointmentResponse, 0, len(appointments)),
	}

	for _, apt := range appointments {
		resp.Appointments = append(resp.Appointments, s.appointmentToProto(apt))
	}

	return resp, nil
}

func (s *AppointmentServer) UpdateAppointmentStatus(ctx context.Context, req *proto.UpdateStatusRequest) (*proto.AppointmentResponse, error) {
	// Validate status
	validStatuses := map[string]bool{
		"new":         true,
		"in_progress": true,
		"done":        true,
	}
	if !validStatuses[req.Status] {
		return nil, status.Error(codes.InvalidArgument, "Invalid status. Must be one of: new, in_progress, done")
	}

	// Get current appointment
	appointment, err := s.getUC.Execute(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Appointment not found")
	}

	// Update status
	newStatus := model.Status(req.Status)
	err = s.updateUC.Execute(ctx, req.Id, newStatus)
	if err != nil {
		// Check for invalid transition error
		if errors.Is(err, usecase.ErrStatusTransition) {
			return nil, status.Error(codes.InvalidArgument, "Cannot revert from done to new")
		}
		if errors.Is(err, usecase.ErrInvalidStatus) {
			return nil, status.Error(codes.InvalidArgument, "Invalid status")
		}
		return nil, status.Error(codes.Internal, "Failed to update appointment status")
	}

	// Get updated appointment
	appointment, err = s.getUC.Execute(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to retrieve updated appointment")
	}

	return s.appointmentToProto(appointment), nil
}

// appointmentToProto converts domain model to proto response
func (s *AppointmentServer) appointmentToProto(apt *model.Appointment) *proto.AppointmentResponse {
	return &proto.AppointmentResponse{
		Id:          apt.ID,
		Title:       apt.Title,
		Description: apt.Description,
		DoctorId:    apt.DoctorID,
		Status:      string(apt.Status),
		CreatedAt:   apt.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   apt.UpdatedAt.Format(time.RFC3339),
	}
}
