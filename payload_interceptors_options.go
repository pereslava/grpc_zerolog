package grpc_zerolog

import (
	"github.com/rs/zerolog"
)

var (
	// DefaultPayloadDecider is the default implementation of payload decider
	// returns always true
	DefaultPayloadDecider PayloadDecider = func(fullMethodName string) bool {
		return true
	}

	// DefaultPayloadLogLevel is the default log level of payload interceptors
	DefaultPayloadLogLevel zerolog.Level = zerolog.TraceLevel

	// DefaultLogErrorsDecider is the default decider for logging errors, it returns true if error exist with WarnLevel
	DefaultLogErrorsDecider LogErrorsDecider = func(fullMethodName string, err error) (bool, zerolog.Level) {
		if err == nil {
			return false, zerolog.NoLevel
		}
		return true, zerolog.WarnLevel
	}

	defaultPayloadOptions = &payloadOptions{
		decider:         DefaultPayloadDecider,
		shouldLogErrors: DefaultLogErrorsDecider,
		level:           DefaultPayloadLogLevel,
	}
)

// WithPayloadDecider customizes the function for deciding if the gRPC interceptor logs should log, depends on fullMethodName
func WithPayloadDecider(f PayloadDecider) PayloadOption {
	return func(o *payloadOptions) {
		o.decider = f
	}
}

// WithLogErrorsDecider customizes the function for deciding if the gRPC interceptor logs should log in error case
// this decider will be executed if logging payload disabled by log level or by PayloadDecider
func WithLogErrorsDecider(f LogErrorsDecider) PayloadOption {
	return func(o *payloadOptions) {
		o.shouldLogErrors = f
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

// LogErrorsDecider defines rules for suppressing log payload in error case, also returns the log level for zerolog
type LogErrorsDecider func(fullMethodName string, err error) (bool, zerolog.Level)

type payloadOptions struct {
	decider         PayloadDecider
	shouldLogErrors LogErrorsDecider
	level           zerolog.Level
}

func evaluatePayloadOptions(opts []PayloadOption) *payloadOptions {
	optCopy := &payloadOptions{}
	*optCopy = *defaultPayloadOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func (o *payloadOptions) shouldLog(method string) bool {
	gl := zerolog.GlobalLevel()
	switch {
	case !o.decider(method):
		return false
	case gl == zerolog.NoLevel, o.level == zerolog.NoLevel:
		return false
	case o.level < gl:
		return false
	default:
		return true
	}
}
