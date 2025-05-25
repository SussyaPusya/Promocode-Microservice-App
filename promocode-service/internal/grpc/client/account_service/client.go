package account_service

import (
	"context"
	"fmt"

	"gitlab.com/pisya-dev/account-service/pkg/api/account_service"
)

// Client is a grpc Client for account service
type Client struct {
	Client account_service.Account_ServiceClient
}

// NewClient constructs a new mock Client
func NewClient(accountServiceClient account_service.Account_ServiceClient) *Client {
	return &Client{Client: accountServiceClient}
}

// GetCompanyNameByCompanyID gets company name by company id
func (c *Client) GetCompanyNameByCompanyID(ctx context.Context, companyID string) (string, error) {

	resp, err := c.Client.GetBuisness(ctx, &account_service.GetBuisnessRequest{
		Id: companyID,
	})
	if err != nil {
		return "", fmt.Errorf("c.Client.GetBuisness: %w", err)
	}

	return resp.Name, nil

}
