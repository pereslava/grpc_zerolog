package grpc_zerolog

import (
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
)

var (
	// DefaultCodeToLevelFunc is the default implementation code to level logic.
	// returns Info on codes.OK and Error in all other cases
	DefaultCodeToLevelFunc CodeToLevel = func(code codes.Code) zerolog.Level {
		if code == codes.OK {
			return zerolog.InfoLevel
		}
		return zerolog.ErrorLevel
	}

	// DefaultDeciderFunc is the default implementation of decider
	// returns true always
	DefaultDeciderFunc Decider = func(fullMethodName string, err error) bool {
		return true
	}

	defaultOptions = &options{
		levelFunc: DefaultCodeToLevelFunc,
		shouldLog: DefaultDeciderFunc,
	}
)

// CodeToLevel function defines the mapping between gRPC return codes and interceptor log level
type CodeToLevel func(code codes.Code) zerolog.Level

// Decider function defines rules for suppressing any interceptor logs
type Decider func(fullMethodName string, err error) bool

// Option used to configure the interceptors
type Option func(*options)

// WithLevels customizes the function for mapping gRPC return codes and interceptor log level statements
func WithLevels(f CodeToLevel) Option {
	return func(o *options) {
		o.levelFunc = f
	}
}

// WithDecider customizes the function for deciding if the gRPC interceptor logs should log depends on fullMethodName and error from handler
func WithDecider(f Decider) Option {
	return func(o *options) {
		o.shouldLog = f
	}
}

type options struct {
	levelFunc CodeToLevel
	shouldLog Decider
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}
