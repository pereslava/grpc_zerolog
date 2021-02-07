/*
'grpc_zerolog' is a gRPC logging interceptors backed by zerolog loggers.

It inspired by https://github.com/grpc-ecosystem/go-grpc-middleware/tree/master/logging packages but without dependency from https://github.com/grpc-ecosystem/go-grpc-middleware

It accepts a user-configured zerolog.Logger interface that will be used for logging compleeted gRPC calls, and be populated into the 'context.Context' passed
into gRPC handler code.
*/
package grpc_zerolog
