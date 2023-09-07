package grpcutil

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/golang/protobuf/proto"
	"time"

	"go.olapie.com/logs"
	"go.olapie.com/ola/activity"
	"go.olapie.com/ola/headers"
	"go.olapie.com/security/base62"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func DialTLS(ctx context.Context, server string, options ...grpc.DialOption) (cc *grpc.ClientConn, err error) {
	options = append(options, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	return dial(ctx, server, options...)
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
	return dial(ctx, server, options...)
}

func Dial(ctx context.Context, server string, options ...grpc.DialOption) (cc *grpc.ClientConn, err error) {
	options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return dial(ctx, server, options...)
}

// WithSigner set trace id, api key and other properties in metadata
func WithSigner(createAPIKey func(md metadata.MD)) grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = signClientContext(ctx, createAPIKey)
		return invoker(ctx, method, req, reply, cc, opts...)
	})
}

func dial(ctx context.Context, server string, options ...grpc.DialOption) (cc *grpc.ClientConn, err error) {
	for i := 0; i < 3; i++ {
		cc, err = grpc.DialContext(ctx, server, options...)
		if err == nil {
			return cc, nil
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		time.Sleep(time.Second)
	}
	return nil, err
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

func Retry[IN proto.Message, OUT proto.Message](ctx context.Context, retries int, backoff time.Duration, call func(ctx context.Context, in IN, options ...grpc.CallOption) (OUT, error), in IN, options ...grpc.CallOption) (OUT, error) {
	var out OUT
	var err error
	for i := 0; i < retries; i++ {
		out, err = call(ctx, in, options...)
		if err == nil {
			break
		}
		time.Sleep(backoff)
	}
	return out, err
}
