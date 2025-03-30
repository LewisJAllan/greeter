package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"

	"github.com/LewisJAllan/application-helper/zaphelper"
)

// Service is the representation of the application we are running
// TODO: expand, build options and create default option behaviour
type Service struct {
	name    string
	options options

	onShutdown []func(context.Context)

	stopSignalTime int64
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) OnShutdown(fn func(context.Context)) *Service {
	s.onShutdown = append(s.onShutdown, fn)
	return s
}

// TODO: expand on options, timeouts, health checker, readiness and liveness
type options struct {
	shutdownTimeout time.Duration
	timeout         time.Duration
}

type Option func(o *options)

func defaultOpts() options {
	return options{
		shutdownTimeout: time.Second * 15,
		timeout:         time.Second * 60,
	}
}

func (s *Service) run(setupApplication SetupApplication) error {
	// logger with service information attached from runtime
	logger := zaphelper.ZapLogger.With(
		zap.String("service", s.name),
	)

	// inject logger into context
	ctx := context.WithValue(context.Background(), struct{}{}, logger)

	// TODO: Continue from this point
	defer func() {
		loggerWithField := logger
		signalTime := atomic.LoadInt64(&s.stopSignalTime)

		if signalTime != 0 {
			now := time.Now().UnixNano()
			shutdownDuration := time.Duration(now - signalTime)
			loggerWithField = logger.With(zap.Int64("shutdown_duration_ms", shutdownDuration.Milliseconds()))
		}
		loggerWithField.Info("service stopped")
	}()

	logger.Info("service starting")

	// set number of CPUs to be used by the app according to the container quota
	_, err := maxprocs.Set(maxprocs.Logger(zap.NewStdLog(logger).Printf))
	if err != nil {
		return fmt.Errorf("failed to set GOMAXPROCS: %w", err)
	}

	logger.Info("set GOMAXPROCS",
		zap.Int("num_cpu", runtime.NumCPU()),
		zap.Int("GOMAXPROCS", runtime.GOMAXPROCS(0)),
	)

	errs := make(chan error, 1)

	runners, ctx, err := s.setupRunners(ctx, setupApplication)
	if err != nil {
		return fmt.Errorf("application: unable to initialise runners: %w", err)
	}
	defer s.stop(ctx, runners)

	runnersCount := 0

	// start the runners one at a time
	for _, r := range runners {
		if r != nil {
			runnersCount++
			s.begin(ctx, r, errs)
		}
	}

	// check if the user passed any runners
	if runnersCount == 0 {
		return nil
	}

	return s.wait(ctx, errs)
}

// setupRunners creates the listeners passed by the user but does not launch them.  This is a wrapper function around
// the setupApplication in order to catch any panics during initialisation.  It also executes the setupApplication
// that the user passed.
func (s *Service) setupRunners(ctx context.Context, setupApplication SetupApplication) ([]Runner, context.Context, error) {
	type runnerSetup struct {
		runners []Runner
		ctx     context.Context
		err     error
	}

	setup := make(chan runnerSetup, 1)
	go func() {
		runners, ctx, err := func() (_ []Runner, _ context.Context, err error) {
			defer func() {
				r := recover()
				if r == nil {
					return
				}

				if v, ok := r.(error); ok {
					err = fmt.Errorf("application: panic during runner setup: %w", v)
				} else {
					err = fmt.Errorf("application: panic during runner setup: %v", r)
				}
				// allow the system to exit normally and not re-panicing
				zaphelper.FromContext(ctx).Error("panic during runner setup", zap.Stack("stack"), zap.Any("panic", r))
			}()

			return setupApplication(ctx, s)
		}()

		setup <- runnerSetup{runners: runners, ctx: ctx, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx, fmt.Errorf("application: runner setup exceeded deadline: %w", ctx.Err())
	case resp := <-setup:
		if resp.ctx != nil {
			ctx = resp.ctx
		}
		return resp.runners, ctx, resp.err
	}
}

// stop terminates all runners in parallel.  Each runner is terminated in their own go routine
func (s *Service) stop(ctx context.Context, runners []Runner) {
	ctx, cancel := context.WithTimeout(ctx, s.options.shutdownTimeout)
	defer cancel()

	var wg sync.WaitGroup
	for _, r := range runners {
		if r == nil {
			continue
		}

		wg.Add(1)

		// the for loop reuses the same var when assigning new values.  This ensures we get a new variable to use in the
		// below go routine
		r := r

		go func() {
			defer wg.Done()
			err := r.Stop(ctx)
			switch {
			case errors.Is(err, context.DeadlineExceeded):
				err = fmt.Errorf("runner did not stop within the allocated time [%.1fs]", s.options.shutdownTimeout.Seconds())
				zaphelper.FromContext(ctx).Error("error stopping runner",
					zap.String("runner", r.Name()),
					zap.Error(err),
				)
			case err != nil:
				zaphelper.FromContext(ctx).Error("error stopping runner",
					zap.String("runner", r.Name()),
					zap.Error(err),
				)
			}
		}()
	}

	wg.Wait()
}

// wait holds the runner until an event has told the runner to do something else
func (s *Service) wait(ctx context.Context, errs <-chan error) error {
	sigs := make(chan os.Signal, 1)
	// kubernetes sends SIGTERM to signal to stop and will then use SIGKILL to kill the application
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigs)

	select {
	case err := <-errs: // the first runner to exit will trigger this, even when err is nil
		return err
	case sig := <-sigs:
		zaphelper.FromContext(ctx).Info("received signal, stopping runner", zap.Stringer("signal", sig))
		// record the time of the signal
		atomic.StoreInt64(&s.stopSignalTime, time.Now().UnixNano())
		return nil
	case <-ctx.Done():
		select {
		case err := <-errs:
			return err
		default:
			return ctx.Err()
		}
	}
}

// begin starts the runner in a new go routine
func (s *Service) begin(ctx context.Context, runner Runner, errs chan<- error) {
	// do not use an err group as then we only return once all functions return.  Once one listener returns an error
	// then exit and handle the shutdown gracefully of all runners.
	go func() {
		err := runner.Start(ctx)
		zaphelper.FromContext(ctx).Info("runner terminated", zap.String("runner_name", runner.Name()), zap.Error(err))

		select {
		case errs <- err:
		default:
		}
	}()
}
