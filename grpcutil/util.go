package grpcutil

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"time"

	"go.olapie.com/logs"
	"go.olapie.com/ola/activity"
	"go.olapie.com/ola/errorutil"
	"go.olapie.com/ola/headers"
	"go.olapie.com/security/base62"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Refer to https://github.com/grpc/grpc/blob/master/doc/http-grpc-status-mapping.md

var statusErrorType = reflect.TypeOf(status.Error(codes.Unknown, ""))

func ServerStart(ctx context.Context, info *grpc.UnaryServerInfo) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "failed reading request metadata")
	}

	if !headers.VerifyAPIKey(md, 10) {
		logs.FromContext(ctx).Warn("invalid api key", md)
		return nil, status.Error(codes.InvalidArgument, "failed verifying")
	}

	a := activity.New(info.FullMethod, md)
	traceID := a.Get(headers.KeyTraceID)
	if traceID == "" {
		traceID = base62.NewUUIDString()
		a.Set(headers.KeyTraceID, traceID)
	}
	ctx = activity.NewIncomingContext(ctx, a)
	logger := logs.FromContext(ctx).With(slog.String("trace_id", traceID))
	fields := make([]any, 0, len(md)+1)
	fields = append(fields, slog.String("full_method", info.FullMethod))
	for k, v := range md {
		if len(v) == 0 {
			continue
		}
		fields = append(fields, slog.String(k, v[0]))
	}
	logger.Info("start", fields...)
	ctx = logs.NewContext(ctx, logger)
	return ctx, nil
}

func ServerFinish(resp any, err error, logger *slog.Logger, startAt time.Time) (any, error) {
	if err == nil {
		logger.Info("finished", slog.Duration("cost", time.Now().Sub(startAt)))
		return resp, nil
	}

	logger.Error("failed", slog.String("error", err.Error()))

	if reflect.TypeOf(err) == statusErrorType {
		return nil, err
	}

	if s := errorutil.GetCode(err); s >= 100 && s < 600 {
		code := HTTPStatusToCode(s)
		logger.Info("failed", slog.Int("status", s), slog.Int("code", int(code)))
		return nil, status.Error(code, err.Error())
	}

	return nil, err
}

func SignClientContext(ctx context.Context) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = make(metadata.MD)
	}

	a := activity.FromOutgoingContext(ctx)

	if traceID := headers.Get(md, headers.KeyTraceID); traceID == "" {
		traceID = a.Get(headers.KeyTraceID)
		if traceID == "" {
			traceID = base62.NewUUIDString()
			logs.FromContext(ctx).Info("generated trace id " + traceID)
		}
		md.Set(headers.KeyTraceID, traceID)
	}

	if clientID := headers.Get(md, headers.KeyClientID); clientID == "" {
		clientID = a.Get(headers.KeyClientID)
		if clientID != "" {
			md.Set(headers.KeyClientID, clientID)
		} else {
			logs.FromContext(ctx).Warn(fmt.Sprintf("missing %s in context", headers.KeyClientID))
		}
	}

	if appID := headers.Get(md, headers.KeyAppID); appID == "" {
		appID = a.Get(headers.KeyAppID)
		if appID != "" {
			md.Set(headers.KeyAppID, appID)
		} else {
			logs.FromContext(ctx).Warn(fmt.Sprintf("missing %s in context", headers.KeyAppID))
		}
	}

	if auth := headers.Get(md, headers.KeyAuthorization); auth == "" {
		auth = a.Get(headers.KeyAuthorization)
		if auth != "" {
			md.Set(headers.KeyAuthorization, auth)
		} else {
			logs.FromContext(ctx).Warn(fmt.Sprintf("missing %s in context", headers.KeyAuthorization))
		}
	}

	headers.SetAPIKey(md)
	return metadata.NewOutgoingContext(ctx, md)
}

func WithClientCert(cert []byte) grpc.DialOption {
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

func WithClientSign() grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = SignClientContext(ctx)
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
