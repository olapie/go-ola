package grpcutil

import (
	"net/http"
	"reflect"
	"strings"

	"go.olapie.com/ola/headers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Refer to https://github.com/grpc/grpc/blob/master/doc/http-grpc-status-mapping.md

var statusErrorType = reflect.TypeOf(status.Error(codes.Unknown, ""))

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
	case headers.LowerKeyClientID, headers.LowerKeyAppID, headers.LowerKeyTraceID, headers.LowerKeyAPIKey, headers.LowerKeyUserAgent, headers.LowerKeyAuthorization, headers.LowerKeyLocation:
		return key, true
	default:
		return "", false
	}
}
