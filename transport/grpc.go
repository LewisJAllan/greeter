package transport

import (
	"context"

	schemas "github.com/LewisJAllan/schemas/playgroundpb/playground"

	"github.com/LewisJAllan/greeter/service"
)

type Service interface {
	Respond(ctx context.Context, request service.RespondRequest) (service.RespondRequest, error)
}

type client struct {
	service Service
	greeter schemas.UnimplementedGreeterServer
}
