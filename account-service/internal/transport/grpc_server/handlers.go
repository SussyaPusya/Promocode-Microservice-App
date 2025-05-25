package grpc_server

import (
	"context"
	"fmt"

	"gitlab.com/pisya-dev/account-service/internal/domain"
	pb "gitlab.com/pisya-dev/account-service/pkg/api/account_service"
	"gitlab.com/pisya-dev/account-service/pkg/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	CreateUser(context.Context, *domain.User) error
	UpdateUser(context.Context, *domain.User) error
	DeleteUser(context.Context, *domain.User) error

	CreateBuisness(context.Context, *domain.Business) error
	GetBuisness(context.Context, *domain.Business) (string, error)
	GetUser(context.Context, *domain.User) (*domain.User, error)
}

type Server struct {
	pb.UnimplementedAccount_ServiceServer
	JWTservice *jwt.ServiceJWT
	service    Service
}

func NewServer(service Service, jwt *jwt.ServiceJWT) *Server {
	return &Server{service: service, JWTservice: jwt}
}

func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{PingMessage: "PONG!!!!!"}, nil
}

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user := &domain.User{
		Guid:       req.GetId(),
		Name:       req.GetName(),
		Surname:    req.GetSurname(),
		Avatar_url: req.GetAvatarUrl(),
		Age:        req.GetAge(),
		Country:    req.GetCountry(),
	}

	err := s.service.CreateUser(ctx, user)
	if err != nil {
		return &pb.CreateUserResponse{}, status.Errorf(codes.Internal, "filed create user")
	}

	return &pb.CreateUserResponse{}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {

	user := &domain.User{
		Guid:       req.GetUuid(),
		Name:       req.GetName(),
		Surname:    req.GetSurname(),
		Avatar_url: req.GetAvatarUrl(),
		Age:        req.GetAge(),
		Country:    req.GetCountry(),
	}

	err := s.service.UpdateUser(ctx, user)
	if err != nil {
		return &pb.UpdateUserResponse{}, status.Errorf(codes.Internal, "filed update user")
	}

	return &pb.UpdateUserResponse{}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	user := &domain.User{
		Guid: req.GetId()}

	err := s.service.DeleteUser(ctx, user)
	if err != nil {
		return &pb.DeleteUserResponse{}, status.Errorf(codes.Internal, "filed delete user")
	}

	return &pb.DeleteUserResponse{}, nil
}

func (s *Server) CreateBuisness(ctx context.Context, req *pb.CreateBuisnessRequest) (*pb.CreateBuisnessResponse, error) {

	fmt.Println(req)
	id, ok := ctx.Value(domain.Uuid).(string)

	if !ok {
		if req.GetUuid() == "" {
			return &pb.CreateBuisnessResponse{}, status.Errorf(codes.Internal, "nothing uuid or jwt")
		}
		id = req.Uuid

	}

	bis := &domain.Business{
		Guid: id,
		Name: req.GetName(),
	}
	err := s.service.CreateBuisness(ctx, bis)
	if err != nil {
		return &pb.CreateBuisnessResponse{}, status.Errorf(codes.Internal, "filed create business")
	}

	return &pb.CreateBuisnessResponse{}, nil
}

func (s *Server) GetBuisness(ctx context.Context, req *pb.GetBuisnessRequest) (*pb.GetBuisnessResponse, error) {
	bis := &domain.Business{
		Guid: req.GetId(),
	}

	name, err := s.service.GetBuisness(ctx, bis)
	if err != nil {
		return &pb.GetBuisnessResponse{}, status.Errorf(codes.NotFound, "filed found business")
	}

	return &pb.GetBuisnessResponse{Name: name}, nil
}

func (s *Server) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {

	user := &domain.User{
		Guid: req.GetUuid(),
	}

	userProfile, err := s.service.GetUser(ctx, user)
	if err != nil {
		return &pb.GetUserProfileResponse{}, status.Errorf(codes.NotFound, "filed found profile")
	}

	return &pb.GetUserProfileResponse{Name: userProfile.Name,
		Surname:   userProfile.Surname,
		Age:       user.Age,
		AvatarUrl: userProfile.Avatar_url,
		Country:   userProfile.Country,
	}, nil
}
