package grpcutil

import (
	"context"
	"crypto/tls"
	"go.olapie.com/logs"
	"log/slog"
	"net/http"
	"reflect"
	"time"

	"go.olapie.com/security/base62"
	"go.olapie.com/utils"
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

	if !VerifyAPIKey(md, 10) {
		logs.FromCtx(ctx).Warn("invalid api key", md)
		return nil, status.Error(codes.InvalidArgument, "failed verifying")
	}

	traceID := GetTraceID(md)
	ctx = utils.NewRequestContextBuilder(ctx).WithAppID(GetAppID(md)).
		WithClientID(GetClientID(md)).
		WithTraceID(traceID).
		Build()
	logger := logs.FromCtx(ctx).With(slog.String("trace_id", traceID))
	fields := make([]any, 0, len(md)+1)
	fields = append(fields, slog.String("full_method", info.FullMethod))
	for k, v := range md {
		if len(v) == 0 {
			continue
		}
		fields = append(fields, slog.String(k, v[0]))
	}
	logger.Info("start", fields...)
	ctx = logs.NewCtx(ctx, logger)
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

	if s := utils.GetErrorCode(err); s >= 100 && s < 600 {
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

	if traceID := GetTraceID(md); traceID == "" {
		traceID = utils.GetTraceID(ctx)
		if traceID == "" {
			traceID = base62.NewUUIDString()
		}
		SetTraceID(md, traceID)
	}

	if clientID := GetClientID(md); clientID == "" {
		clientID = utils.GetClientID(ctx)
		if clientID != "" {
			SetClientID(md, clientID)
		} else {
			slog.Warn("missing ClientID in context")
		}
	}

	if appID := GetAppID(md); appID == "" {
		appID = utils.GetAppID(ctx)
		if appID != "" {
			SetAppID(md, appID)
		} else {
			slog.Warn("missing AppID in context")
		}
	}

	if auth := GetAuthorization(md); auth == "" {
		auth = utils.GetAuthorization(ctx)
		if auth != "" {
			SetAuthorization(md, auth)
		} else {
			slog.Warn("missing Authorization in context")
		}
	}

	SetAPIKey(md)
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
