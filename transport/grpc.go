package transport

import (
	"context"

	schemas "github.com/LewisJAllan/schemas/playgroundpb/playground"
	"google.golang.org/grpc"

	"github.com/LewisJAllan/greeter/service"
)

type Service interface {
	Respond(ctx context.Context, request service.RespondRequest) (service.RespondResponse, error)
}

type Client struct {
	service Service
	greeter schemas.UnimplementedGreeterServer
}

func NewClient(service Service) *Client {
	return &Client{service: service}
}

func (c *Client) Register(server *grpc.Server) {
	schemas.RegisterGreeterServer(server, c.greeter)
}
