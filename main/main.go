package main

import (
	"context"

	grpclistener "github.com/LewisJAllan/application-helper/listeners/grpc"
	app "github.com/LewisJAllan/application-helper/runner"
	"github.com/LewisJAllan/application-helper/zaphelper"
	"go.uber.org/zap"

	"github.com/LewisJAllan/greeter/service"
	"github.com/LewisJAllan/greeter/transport"
)

const ServiceName = "Greeter"

func main() {
	if err := app.Run(ServiceName, setup); err != nil {
		zaphelper.FromContext(context.Background()).Fatal("failed to start service",
			zap.String("service_name", ServiceName),
			zap.Error(err))
	}
}

func setup(ctx context.Context, s *app.Service) ([]app.Runner, context.Context, error) {
	s.OnShutdown(func(ctx context.Context) {
		zaphelper.Info(ctx, "shutdown",
			zap.String("service_name", s.Name()))
	})

	zaphelper.Info(ctx, "starting service",
		zap.String("service_name", s.Name()),
	)

	svc := service.NewService()

	client := transport.NewClient(&svc)

	return []app.Runner{
		grpclistener.New(client),
	}, ctx, nil
}
