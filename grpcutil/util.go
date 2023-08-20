package grpcutil

import (
	"context"
	"crypto/tls"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"time"

	"go.olapie.com/logs"
	"go.olapie.com/ola/activity"
	"go.olapie.com/ola/errorutil"
	"go.olapie.com/ola/headers"
	"go.olapie.com/ola/types"
	"go.olapie.com/security/base62"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Refer to https://github.com/grpc/grpc/blob/master/doc/http-grpc-status-mapping.md

var statusErrorType = reflect.TypeOf(status.Error(codes.Unknown, ""))

func ServerStart(ctx context.Context,
	info *grpc.UnaryServerInfo,
	verifyAPIKey func(ctx context.Context, md metadata.MD) bool,
	authenticate func(ctx context.Context, md metadata.MD) types.UserID) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "failed reading request metadata")
	}

	a := activity.New(info.FullMethod, md)
	ctx = activity.NewIncomingContext(ctx, a)
	traceID := a.GetTraceID()
	if traceID == "" {
		traceID = base62.NewUUIDString()
		a.SetTraceID(traceID)
	}
	logger := logs.FromContext(ctx).With(slog.String("trace_id", traceID))
	ctx = logs.NewContext(ctx, logger)
	fields := make([]any, 0, len(md)+1)
	fields = append(fields, slog.String("full_method", info.FullMethod))
	for k, v := range md {
		if len(v) == 0 {
			continue
		}
		fields = append(fields, slog.String(k, v[0]))
	}
	logger.Info("START", fields...)

	if !verifyAPIKey(ctx, md) {
		logger.Error("invalid api key", md)
		return nil, status.Error(codes.InvalidArgument, "failed verifying")
	}

	uid := authenticate(ctx, md)
	if uid != nil {
		a.SetUserID(uid)
		logger.Info("authenticated", slog.Any("uid", uid.Value()))
	}
	return ctx, nil
}

func ServerFinish(resp any, err error, logger *slog.Logger, startAt time.Time) (any, error) {
	if err == nil {
		logger.Info("END", slog.Duration("cost", time.Now().Sub(startAt)))
		return resp, nil
	}

	if reflect.TypeOf(err) == statusErrorType {
		logger.Error("END", logs.Err(err))
		return nil, err
	}

	if s := errorutil.GetCode(err); s >= 100 && s < 600 {
		code := HTTPStatusToCode(s)
		logger.Info("END", slog.Int("status", s), slog.Int("code", int(code)), logs.Err(err))
		return nil, status.Error(code, err.Error())
	}
	logger.Error("END", logs.Err(err))

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

func WithSecure(cert []byte) grpc.DialOption {
	config := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{cert},
			},
		},
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	return grpc.WithTransportCredentials(credentials.NewTLS(config))
}

func WithInsecure() grpc.DialOption {
	return grpc.WithTransportCredentials(insecure.NewCredentials())
}

func WithSigner(createAPIKey func(md metadata.MD)) grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = signClientContext(ctx, createAPIKey)
		return invoker(ctx, method, req, reply, cc, opts...)
	})
}

func CodeToHTTPStatus(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.ResourceExhausted:
		return http.StatusServiceUnavailable
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.OutOfRange:
		return http.StatusRequestedRangeNotSatisfiable
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

func HTTPStatusToCode(s int) codes.Code {
	switch s {
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusNotImplemented:
		return codes.Unimplemented
	case http.StatusInternalServerError:
		return codes.Internal
	case http.StatusBadGateway:
		return codes.Unavailable
	case http.StatusServiceUnavailable:
		return codes.Unavailable
	default:
		return codes.Unknown
	}
}

func MatchMetadata(key string) (string, bool) {
	key = strings.ToLower(key)
	switch key {
	case headers.KeyClientID, headers.KeyAppID, headers.KeyTraceID, headers.KeyAPIKey:
		return key, true
	default:
		return "", false
	}
}
