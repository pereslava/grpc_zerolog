package grpc_zerolog

import (
	"context"
	"path"
	"time"

	"github.com/pereslava/grpc_zerolog/ctxzerolog"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	msgUnary        message = "finished unary call"
	msgServerStream message = "finished stream call"
	msgClientStream message = "started stream call"
)

// NewUnaryServerInterceptor returns an unary server interceptor that adds zerolog to context and logs the gRPC calls
func NewUnaryServerInterceptor(logger zerolog.Logger, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateOptions(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		l := initLog(ctx, logger, info.FullMethod)

		res, err := handler(ctxzerolog.New(ctx, l.Logger()), req)
		if !o.shouldLog(info.FullMethod, err) {
			return res, err
		}
		doInterceptorLog(l, start, err, msgUnary, o.levelFunc)

		return res, err
	}
}

// NewUnaryClientInterceptor returns an unary client interceptor that logs the gRPC calls
func NewUnaryClientInterceptor(logger zerolog.Logger, opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateOptions(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()

		err := invoker(ctx, method, req, reply, cc, opts...)
		if !o.shouldLog(method, err) {
			return err
		}

		l := initLog(ctx, logger, method)
		doInterceptorLog(l, start, err, msgUnary, o.levelFunc)

		return err
	}
}

// NewStreamServerInterceptor returns a streaming server interceptor that adds zerolog to context and logs the gRPC calls
func NewStreamServerInterceptor(logger zerolog.Logger, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateOptions(opts)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		wrapped := wrapServerStream(stream)
		l := initLog(wrapped.wrappedContext, logger, info.FullMethod)
		wrapped.wrappedContext = ctxzerolog.New(wrapped.wrappedContext, l.Logger())

		err := handler(srv, wrapped)
		if !o.shouldLog(info.FullMethod, err) {
			return err
		}

		doInterceptorLog(l, start, err, msgServerStream, o.levelFunc)

		return err
	}
}

// NewStreamClientInterceptor returns a streaming client interceptor that logs the gRPC calls
func NewStreamClientInterceptor(logger zerolog.Logger, opts ...Option) grpc.StreamClientInterceptor {
	o := evaluateOptions(opts)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		start := time.Now()

		cs, err := streamer(ctx, desc, cc, method, opts...)
		if !o.shouldLog(method, err) {
			return cs, err
		}

		l := initLog(ctx, logger, method)
		doInterceptorLog(l, start, err, msgClientStream, o.levelFunc)

		return cs, err
	}
}

func initLog(ctx context.Context, logger zerolog.Logger, fullMethodString string) zerolog.Context {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)

	with := logger.With().
		Str("grpc.service", service).
		Str("grpc.method", method)

	if ctx == nil {
		return with
	}

	if d, ok := ctx.Deadline(); ok {
		with = with.Time("grpc.request.deadline", d)
	}

	return with
}

type message string

func doInterceptorLog(log zerolog.Context, start time.Time, callError error, msg message, ctl CodeToLevel) {
	code := status.Code(callError)
	with := log.Str("grpc.code", code.String()).Dur("grpc.time_ms", time.Since(start))
	if callError != nil {
		with = with.Err(callError)
	}
	l := with.Logger()
	l.WithLevel(ctl(code)).Msg(string(msg))
}

type wrappedServerStream struct {
	grpc.ServerStream
	wrappedContext context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.wrappedContext
}

func wrapServerStream(stream grpc.ServerStream) *wrappedServerStream {
	if existing, ok := stream.(*wrappedServerStream); ok {
		return existing
	}
	return &wrappedServerStream{ServerStream: stream, wrappedContext: stream.Context()}
}
