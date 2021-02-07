package grpc_zerolog

import (
	"bytes"
	"context"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

const (
	msgPayloadRequest  payloadMessage = "grpc.request.payload"
	msgPayloadResponse payloadMessage = "grpc.response.payload"
)

// NewPayloadUnaryServerInterceptor return an unary server interceptor that logs the payloads of requests and responses
func NewPayloadUnaryServerInterceptor(logger zerolog.Logger, opts ...PayloadOption) grpc.UnaryServerInterceptor {
	o := evaluatePayloadOptions(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !o.shouldLog(info.FullMethod) {
			return handler(ctx, req)
		}

		l := initLog(nil, logger, info.FullMethod).Logger()
		logProtoMessageAsJson(l, o.level, req, msgPayloadRequest)
		res, err := handler(ctx, req)
		if err == nil {
			logProtoMessageAsJson(l, o.level, res, msgPayloadResponse)
		}
		return res, err
	}
}

// NewPayloadUnaryClientInterceptor returns an unary client interceptor that logs the payloads of requests and responses
func NewPayloadUnaryClientInterceptor(logger zerolog.Logger, opts ...PayloadOption) grpc.UnaryClientInterceptor {
	o := evaluatePayloadOptions(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !o.shouldLog(method) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		l := initLog(nil, logger, method).Logger()
		logProtoMessageAsJson(l, o.level, req, msgPayloadRequest)
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			logProtoMessageAsJson(l, o.level, reply, msgPayloadResponse)
		}
		return err
	}
}

// NewPayloadStreamServerInterceptor returns a streaming server interceptor that logs the payloads of requests and responses
func NewPayloadStreamServerInterceptor(logger zerolog.Logger, opts ...PayloadOption) grpc.StreamServerInterceptor {
	o := evaluatePayloadOptions(opts)
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !o.shouldLog(info.FullMethod) {
			return handler(srv, ss)
		}

		l := initLog(nil, logger, info.FullMethod).Logger()
		newStream := &loggingServerStream{ServerStream: ss, l: l, level: o.level}
		return handler(srv, newStream)
	}
}

// NewPayloadUnaryClientInterceptor returns a streaming client interceptor that logs the payloads of requests and responses
func NewPayloadStreamClientInterceptor(logger zerolog.Logger, opts ...PayloadOption) grpc.StreamClientInterceptor {
	o := evaluatePayloadOptions(opts)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if !o.shouldLog(method) {
			return streamer(ctx, desc, cc, method, opts...)
		}

		l := initLog(nil, logger, method).Logger()
		cs, err := streamer(ctx, desc, cc, method, opts...)
		newStream := &loggingClientStream{ClientStream: cs, l: l, level: o.level}
		return newStream, err
	}
}

type payloadMessage string

func logProtoMessageAsJson(logger zerolog.Logger, level zerolog.Level, pbMsg interface{}, key payloadMessage) {
	if p, ok := pbMsg.(proto.Message); ok {
		m := &jsonpbMarshalleble{p}
		json, err := m.MarshalJSON()
		if err != nil {
			logger.WithLevel(level).Err(err).Msg("Failed to marshal message")
		}
		logger.WithLevel(level).RawJSON(string(key), json).Send()
	}
}

type loggingServerStream struct {
	grpc.ServerStream
	l     zerolog.Logger
	level zerolog.Level
}

func (s *loggingServerStream) SendMsg(m interface{}) error {
	err := s.ServerStream.SendMsg(m)
	if err == nil {
		logProtoMessageAsJson(s.l, s.level, m, msgPayloadResponse)
	}
	return err
}

func (s *loggingServerStream) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)
	if err == nil {
		logProtoMessageAsJson(s.l, s.level, m, msgPayloadRequest)
	}
	return err
}

type loggingClientStream struct {
	grpc.ClientStream
	l     zerolog.Logger
	level zerolog.Level
}

func (s *loggingClientStream) SendMsg(m interface{}) error {
	err := s.ClientStream.SendMsg(m)
	if err == nil {
		logProtoMessageAsJson(s.l, s.level, m, msgPayloadRequest)
	}
	return err
}

func (s *loggingClientStream) RecvMsg(m interface{}) error {
	err := s.ClientStream.RecvMsg(m)
	if err == nil {
		logProtoMessageAsJson(s.l, s.level, m, msgPayloadResponse)
	}
	return err
}

type jsonpbMarshalleble struct {
	proto.Message
}

func (j *jsonpbMarshalleble) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}

	m := &jsonpb.Marshaler{}

	m.MarshalToString(j.Message)

	if err := m.Marshal(b, j.Message); err != nil {
		return nil, fmt.Errorf("jsonpb serializer failed: %v", err)
	}
	return b.Bytes(), nil
}
