package grpc_zerolog_test

import (
	"net"
	"os"
	"path"

	"github.com/pereslava/grpc_zerolog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func ExampleWithDecider() {
	addr := ":9000"
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Err(err).Str("addr", addr).Msg("cannont listen")
	}
	defer conn.Close()

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_zerolog.NewUnaryServerInterceptor(
				log.Logger,
				grpc_zerolog.WithDecider(func(fullMethodName string, err error) bool {
					// will not log Hello method if no error
					method := path.Base(fullMethodName)
					if err == nil && method == "Hello" {
						return false
					}
					return true

				}),
			),
		),
	)

	// pb.Register .... your service (s, my_service.New(opts...))

	log.Info().Str("addr", addr).Msg("serving")
	if err := s.Serve(conn); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}

func ExampleWithLevels() {
	addr := ":9090"
	log.Info().Str("addr", addr).Msg("connecting")

	conn, err := grpc.Dial(
		addr,
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(
			grpc_zerolog.NewUnaryClientInterceptor(
				log.Logger,
				// set custom log to level logic
				grpc_zerolog.WithLevels(func(code codes.Code) zerolog.Level {
					switch code {
					case codes.OK:
						return zerolog.DebugLevel
					case codes.Canceled:
						return zerolog.DebugLevel
					case codes.Unknown:
						return zerolog.InfoLevel
					case codes.InvalidArgument:
						return zerolog.DebugLevel
					case codes.DeadlineExceeded:
						return zerolog.InfoLevel
					case codes.NotFound:
						return zerolog.DebugLevel
					case codes.AlreadyExists:
						return zerolog.DebugLevel
					case codes.PermissionDenied:
						return zerolog.InfoLevel
					case codes.Unauthenticated:
						return zerolog.InfoLevel
					case codes.ResourceExhausted:
						return zerolog.DebugLevel
					case codes.FailedPrecondition:
						return zerolog.DebugLevel
					case codes.Aborted:
						return zerolog.DebugLevel
					case codes.OutOfRange:
						return zerolog.DebugLevel
					case codes.Unimplemented:
						return zerolog.WarnLevel
					case codes.Internal:
						return zerolog.WarnLevel
					case codes.Unavailable:
						return zerolog.WarnLevel
					case codes.DataLoss:
						return zerolog.WarnLevel
					default:
						return zerolog.InfoLevel
					}
				}),
			),
		),
	)

	if err != nil {
		log.Fatal().Err(err).Str("addr", addr).Send()
	}
	defer conn.Close()
}

var (
	customDecider             grpc_zerolog.Decider        = grpc_zerolog.DefaultDeciderFunc
	customPayloadDecider      grpc_zerolog.PayloadDecider = grpc_zerolog.DefaultPayloadDecider
	customCodeToLevelFunction grpc_zerolog.CodeToLevel    = grpc_zerolog.DefaultCodeToLevelFunc
)

func Example_serverInitializationSimple() {
	// Make sure that log statements internal to gRPC library are logged using the zerolog Logger as well
	grpc_zerolog.ReplaceGrpcLogger(zerolog.New(os.Stderr).Level(zerolog.ErrorLevel))

	// Create and customize the zerolog logger instance
	serverLogger := log.Level(zerolog.TraceLevel)
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_zerolog.NewPayloadUnaryServerInterceptor(serverLogger),
			grpc_zerolog.NewUnaryServerInterceptor(serverLogger),
		),
		grpc.ChainStreamInterceptor(
			grpc_zerolog.NewPayloadStreamServerInterceptor(serverLogger),
			grpc_zerolog.NewStreamServerInterceptor(serverLogger),
		),
	)
}

func Example_serverInitializationWithOptions() {
	_ = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_zerolog.NewPayloadUnaryServerInterceptor(
				log.Logger,
				grpc_zerolog.WithPayloadDecider(customPayloadDecider),
				grpc_zerolog.WithPayloadLevel(zerolog.Disabled),
			),
			grpc_zerolog.NewUnaryServerInterceptor(
				log.Logger,
				grpc_zerolog.WithDecider(customDecider),
				grpc_zerolog.WithLevels(customCodeToLevelFunction),
			),
		),
		grpc.ChainStreamInterceptor(
			grpc_zerolog.NewPayloadStreamServerInterceptor(
				log.Logger,
				grpc_zerolog.WithPayloadDecider(customPayloadDecider),
				grpc_zerolog.WithPayloadLevel(zerolog.InfoLevel),
			),
			grpc_zerolog.NewStreamServerInterceptor(
				log.Logger,
				grpc_zerolog.WithDecider(customDecider),
				grpc_zerolog.WithLevels(customCodeToLevelFunction),
			),
		),
	)
}

func Example_clientInitializationSimple() {
	addr := "localhost:9000"
	log.Print("Connecting to ", addr)
	// Make sure that log statements internal to gRPC library are logged using the zerolog Logger as well
	grpc_zerolog.ReplaceGrpcLogger(zerolog.New(os.Stderr).Level(zerolog.ErrorLevel))

	// Create and customize the zerolog logger instance
	clientLogger := log.Level(zerolog.TraceLevel)
	conn, err := grpc.Dial(
		addr,
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(
			grpc_zerolog.NewPayloadUnaryClientInterceptor(clientLogger),
			grpc_zerolog.NewUnaryClientInterceptor(clientLogger),
		),
		grpc.WithChainStreamInterceptor(
			grpc_zerolog.NewPayloadStreamClientInterceptor(clientLogger),
			grpc_zerolog.NewStreamClientInterceptor(clientLogger),
		),
	)
	if err != nil {
		log.Fatal().Err(err).Str("addr", addr).Send()
	}
	defer conn.Close()

}
