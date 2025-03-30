package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	prometheusgrpc "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

type ListenConfig interface {
	Listen(ctx context.Context, net, addr string) (net.Listener, error)
}

type Registerer interface {
	Register(s *grpc.Server)
}

type register func(s *grpc.Server)

func (r register) Register(s *grpc.Server) {
	r(s)
}

func MultiListener(rs ...Registerer) Registerer {
	return register(func(s *grpc.Server) {
		for _, r := range rs {
			r.Register(s)
		}
	})
}

type options struct {
	serverOptions []grpc.ServerOption

	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor
}

func buckets(durations ...time.Duration) []float64 {
	s := make([]float64, len(durations))
	for i, duration := range durations {
		s[i] = duration.Seconds()
	}
	return s
}

func (o options) grpcServerOpts() []grpc.ServerOption {
	prometheusgrpc.EnableHandlingTimeHistogram(prometheusgrpc.WithHistogramBuckets(buckets(
		time.Millisecond,
		time.Millisecond*10,
		time.Millisecond*100,
		time.Millisecond*250,
		time.Millisecond*500,
		time.Second,
		time.Second*5,
		time.Second*10,
	)))

	// TODO: Make custom interceptors

	streamInterceptors := append([]grpc.StreamServerInterceptor{
		prometheusgrpc.StreamServerInterceptor,
	}, o.streamInterceptors...)

	unaryInterceptors := append([]grpc.UnaryServerInterceptor{
		prometheusgrpc.UnaryServerInterceptor,
	}, o.unaryInterceptors...)

	return append(o.serverOptions,
		grpc.StreamInterceptor(grpcmiddleware.ChainStreamServer(streamInterceptors...)),
		grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(unaryInterceptors...)),
	)
}

type Option func(o *options)

func WithGRPCOptions(opts ...grpc.ServerOption) Option {
	return func(o *options) {
		o.serverOptions = append(o.serverOptions, opts...)
	}
}

// WithStreamInterceptors allows setting interceptors that are not chained together.
// Default interceptors are already installed.
func WithStreamInterceptors(si ...grpc.StreamServerInterceptor) Option {
	return func(o *options) {
		o.streamInterceptors = append(o.streamInterceptors, si...)
	}
}

// WithUnaryInterceptors allows setting interceptors that are not chained together.
// Default interceptors are already installed.
func WithUnaryInterceptors(ui ...grpc.UnaryServerInterceptor) Option {
	return func(o *options) {
		o.unaryInterceptors = append(o.unaryInterceptors, ui...)
	}
}

type Handler struct {
	opts []grpc.ServerOption
	r    Registerer

	listenCfg ListenConfig
	addr      string
	server    *grpc.Server
}

func New(r Registerer, opts ...Option) *Handler {
	o := options{}

	for _, opt := range opts {
		opt(&o)
	}

	return &Handler{
		opts:      o.grpcServerOpts(),
		r:         r,
		listenCfg: &net.ListenConfig{},
		addr:      ":50051",
	}
}

func (h *Handler) Start(ctx context.Context) error {
	l, err := h.listenCfg.Listen(ctx, "tcp", h.addr)
	if err != nil {
		return fmt.Errorf("grpc: unable to create listener: %w", err)
	}

	s := grpc.NewServer(h.opts...)
	h.r.Register(s)
	prometheusgrpc.Register(s)

	h.server = s

	return s.Serve(l)
}

func (h *Handler) Stop(_ context.Context) error {
	h.server.GracefulStop()
	return nil
}

func (h *Handler) Name() string {
	return "grpc"
}
