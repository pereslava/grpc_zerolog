package grpc_zerolog

import "github.com/rs/zerolog"

var (
	// DefaultPayloadDecider is the default implementation of payload decider
	// returns always true
	DefaultPayloadDecider PayloadDecider = func(fullMethodName string) bool {
		return true
	}

	// DefaultPayloadLogLevel is the default log level of payload interceptors
	DefaultPayloadLogLevel zerolog.Level = zerolog.DebugLevel

	defaultPayloadOptions = &payloadOptions{
		shouldLog: DefaultPayloadDecider,
	}
)

// WithPayloadDecider customizes the function for deciding if the gRPC interceptor logs should log depends on fullMethodName
func WithPayloadDecider(f PayloadDecider) PayloadOption {
	return func(o *payloadOptions) {
		o.shouldLog = f
	}
}

// WithPayloadLevel overrides the log level of payload interceptors
func WithPayloadLevel(l zerolog.Level) PayloadOption {
	return func(o *payloadOptions) {
		o.level = l
	}
}

// PayloadOption used to configure the payload interceptors
type PayloadOption func(*payloadOptions)

// PayloadDecider defines rules for suppressing payload interceptor logs
type PayloadDecider func(fullMethodName string) bool

type payloadOptions struct {
	shouldLog PayloadDecider
	level     zerolog.Level
}

func evaluatePayloadOptions(opts []PayloadOption) *payloadOptions {
	optCopy := &payloadOptions{}
	*optCopy = *defaultPayloadOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}
