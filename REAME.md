# gRPC interceptors backed by zerolog logger.

Inspired by [github.com/grpc-ecosystem/go-grpc-middleware/logging](https://github.com/grpc-ecosystem/go-grpc-middleware/tree/master/logging) and motivated by [this issue](https://github.com/rs/zerolog/issues/211).

Instead of loggers of grpc-middleware it not depend on grpc-ecosystem and not need the grpc_widleware at all. It uses  the grpc.Chain* functions instead. 

[Documentation at pkg.go.dev](https://pkg.go.dev/github.com/pereslava/grpc_zerolog)