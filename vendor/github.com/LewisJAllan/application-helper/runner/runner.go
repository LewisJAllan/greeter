package runner

import (
	"context"

	"github.com/LewisJAllan/application-helper/zaphelper"
)

// Runner represents the code that can be run while exposing the Stop() method, which allows the Run() to gracefully
// shutdown the listener.  The Start() method is called in a new goroutine while the Stop() method is called from a
// management goroutine.  This means that the listener logic responsible for shutting down the active listener needs
// to be thread-safe
type Runner interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Name() string
}

func GetContextWithLogger() context.Context {
	return context.WithValue(context.Background(), struct{}{}, &zaphelper.ZapLogger)
}

// SetupApplication represents the function the user will pass to the Run() that will initialise the listeners.
type SetupApplication func(ctx context.Context, service *Service) ([]Runner, context.Context, error)

// Run launches the application, manages the lifecycle and common setup of a service.  The SetupApplication is executed
// to produce the listener and context.
// TODO: build health checks to add into the Service to ensure the application runs and is both ready and live.
// the SetupApplication is executed and continues to run until Run() returns. The provided context contains the root
// logger.  If the returned context is cancelled, Run() will return an error.  Once Run() returns, the application
// should handle the error if it exists and exit.  Listeners are ran during the lifetime of the service.
func Run(name string, setup SetupApplication, opts ...Option) error {
	o := defaultOpts()

	for _, opt := range opts {
		opt(&o)
	}

	s := &Service{
		name:    name,
		options: o,
	}

	return s.run(setup)
}
