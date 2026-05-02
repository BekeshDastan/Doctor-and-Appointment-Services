package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor-service/internal/model"
	"github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor-service/internal/usecase"
	proto "github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor-service/proto"
)

type DoctorServer struct {
	proto.UnimplementedDoctorServiceServer
	createUC  usecase.CreateDoctorUseCase
	getByIdUC usecase.GetDoctorByIdUseCase
	getAllUC  usecase.GetAllDoctorsUseCase
}

func NewDoctorServer(
	createUC usecase.CreateDoctorUseCase,
	getByIdUC usecase.GetDoctorByIdUseCase,
	getAllUC usecase.GetAllDoctorsUseCase,
) *DoctorServer {
	return &DoctorServer{
		createUC:  createUC,
		getByIdUC: getByIdUC,
		getAllUC:  getAllUC,
	}
}

func (s *DoctorServer) CreateDoctor(ctx context.Context, req *proto.CreateDoctorRequest) (*proto.DoctorResponse, error) {
	doctor := &model.Doctor{
		ID:             uuid.New().String(),
		FullName:       req.FullName,
		Specialization: req.Specialization,
		Email:          req.Email,
	}

	err := s.createUC.Execute(ctx, doctor)
	if err != nil {
		if errors.Is(err, usecase.ErrEmptyFields) {
			return nil, status.Error(codes.InvalidArgument, "Full name, specialization, and email are required")
		}
		if errors.Is(err, usecase.ErrEmailAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "Email already in use")
		}
		return nil, status.Error(codes.Internal, "Failed to create doctor")
	}

	return &proto.DoctorResponse{
		Id:             doctor.ID,
		FullName:       doctor.FullName,
		Specialization: doctor.Specialization,
		Email:          doctor.Email,
	}, nil
}

func (s *DoctorServer) GetDoctor(ctx context.Context, req *proto.GetDoctorRequest) (*proto.DoctorResponse, error) {
	doctor, err := s.getByIdUC.Execute(ctx, req.Id)
	if err != nil {
		if errors.Is(err, usecase.ErrDoctorNotFound) {
			return nil, status.Error(codes.NotFound, "Doctor not found")
		}
		if errors.Is(err, usecase.ErrEmptyID) {
			return nil, status.Error(codes.InvalidArgument, "Doctor ID is required")
		}
		return nil, status.Error(codes.Internal, "Failed to get doctor")
	}

	return &proto.DoctorResponse{
		Id:             doctor.ID,
		FullName:       doctor.FullName,
		Specialization: doctor.Specialization,
		Email:          doctor.Email,
	}, nil
}

func (s *DoctorServer) ListDoctors(ctx context.Context, _ *proto.ListDoctorsRequest) (*proto.ListDoctorsResponse, error) {
	doctors, err := s.getAllUC.Execute(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to list doctors")
	}

	resp := &proto.ListDoctorsResponse{
		Doctors: make([]*proto.DoctorResponse, 0, len(doctors)),
	}

	for _, doc := range doctors {
		resp.Doctors = append(resp.Doctors, &proto.DoctorResponse{
			Id:             doc.ID,
			FullName:       doc.FullName,
			Specialization: doc.Specialization,
			Email:          doc.Email,
		})
	}

	return resp, nil
}
