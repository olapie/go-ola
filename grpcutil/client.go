package grpcutil

import (
	"context"
	"crypto/tls"

	"go.olapie.com/logs"
	"go.olapie.com/ola/activity"
	"go.olapie.com/ola/headers"
	"go.olapie.com/security/base62"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func DialTLS(ctx context.Context, server string, options ...grpc.DialOption) (cc *grpc.ClientConn, err error) {
	options = append(options, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	return grpc.DialContext(ctx, server, options...)
}

func DialWithClientCert(ctx context.Context, server string, cert []byte, options ...grpc.DialOption) (cc *grpc.ClientConn, err error) {
	config := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{cert},
			},
		},
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	options = append(options, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	return grpc.DialContext(ctx, server, options...)
}

func Dial(ctx context.Context, server string, options ...grpc.DialOption) (cc *grpc.ClientConn, err error) {
	options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return grpc.DialContext(ctx, server, options...)
}

// WithSigner set trace id, api key and other properties in metadata
func WithSigner(createAPIKey func(md metadata.MD)) grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = signClientContext(ctx, createAPIKey)
		return invoker(ctx, method, req, reply, cc, opts...)
	})
}

func signClientContext(ctx context.Context, createAPIKey func(md metadata.MD)) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = make(metadata.MD)
	}

	a := activity.FromOutgoingContext(ctx)
	if a != nil {
		activity.CopyHeader(md, a)
	} else {
		logs.FromContext(ctx).Warn("no outgoing context")
	}
	if traceID := headers.GetTraceID(md); traceID == "" {
		if traceID == "" {
			traceID = base62.NewUUIDString()
			logs.FromContext(ctx).Info("generated trace id " + traceID)
		}
		headers.SetTraceID(md, traceID)
	}
	createAPIKey(md)
	return metadata.NewOutgoingContext(ctx, md)
}

func GetErrorCode(err error) codes.Code {
	if s, ok := status.FromError(err); ok {
		return s.Code()
	}
	return codes.Unknown
}
