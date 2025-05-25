package grpc_client

import (
	"context"
	"fmt"

	"gitlab.com/pisya-dev/auth-service/internal/config"
	pb "gitlab.com/pisya-dev/auth-service/pkg/api/account_service" // замените на актуальный путь к protobuf
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AccountServiceClient struct {
	conn   *grpc.ClientConn
	client pb.Account_ServiceClient
}

func NewAccountServiceClient(config config.GrpcConfig) (*AccountServiceClient, error) {
	conn, err := grpc.NewClient(config.AccountClientAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to account service: %w", err)
	}
	client := pb.NewAccount_ServiceClient(conn)
	return &AccountServiceClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *AccountServiceClient) Close() error {
	return c.conn.Close()
}

func (c *AccountServiceClient) GetUserAccount(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {
	return c.client.GetUserProfile(ctx, req)
}

func (c *AccountServiceClient) CreateUserAccount(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return c.client.CreateUser(ctx, req)
}

func (c *AccountServiceClient) CreateBuisnessAccount(ctx context.Context, req *pb.CreateBuisnessRequest) (*pb.CreateBuisnessResponse, error) {
	return c.client.CreateBuisness(ctx, req)
}
func (c *AccountServiceClient) GetBuisnessAccount(ctx context.Context, req *pb.GetBuisnessRequest) (*pb.GetBuisnessResponse, error) {
	return c.client.GetBuisness(ctx, req)
}
