package transport

import (
	"context"
	"fmt"

	schemas "github.com/LewisJAllan/schemas/playgroundpb/playground"
	"google.golang.org/grpc"

	"github.com/LewisJAllan/greeter/service"
)

func (c *Client) SayHello(ctx context.Context, request *schemas.HelloRequest, _ ...grpc.CallOption) (*schemas.HelloReply, error) {
	resp, err := c.service.Respond(ctx, service.RespondRequest{
		OriginalMessage: request.GetName(),
	})
	if err != nil {
		return nil, fmt.Errorf("error occurred: %w", err)
	}

	return &schemas.HelloReply{
		Message: resp.ResponseMessage,
	}, nil
}
