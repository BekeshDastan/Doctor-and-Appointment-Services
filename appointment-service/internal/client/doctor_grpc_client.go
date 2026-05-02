package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	proto "github.com/BekeshDastan/Doctor-and-Appointment-Services/doctor-service/proto"
)

type DoctorClient interface {
	CheckDoctorExists(ctx context.Context, doctorID string) (bool, error)
}

type GRPCDoctorClient struct {
	client proto.DoctorServiceClient
}

func NewGRPCDoctorClient(conn *grpc.ClientConn) *GRPCDoctorClient {
	return &GRPCDoctorClient{
		client: proto.NewDoctorServiceClient(conn),
	}
}

func (c *GRPCDoctorClient) CheckDoctorExists(ctx context.Context, doctorID string) (bool, error) {
	resp, err := c.client.GetDoctor(ctx, &proto.GetDoctorRequest{
		Id: doctorID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return resp != nil, nil
}
