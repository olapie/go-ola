package grpcutil

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.olapie.com/logs"
	"go.olapie.com/ola/activity"
	"go.olapie.com/ola/errorutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
)

type RetryOptions struct {
	Count              int
	Backoff            time.Duration
	RefreshAccessToken func(ctx context.Context) (string, error)
}

type Retry[IN proto.Message, OUT proto.Message] struct {
	options RetryOptions
	call    func(ctx context.Context, in IN, options ...grpc.CallOption) (OUT, error)
}

func NewRetry[IN proto.Message, OUT proto.Message](call func(ctx context.Context, in IN, options ...grpc.CallOption) (OUT, error), options ...func(options *RetryOptions)) *Retry[IN, OUT] {
	r := &Retry[IN, OUT]{
		call: call,
	}
	for _, opt := range options {
		opt(&r.options)
	}
	if r.options.Count <= 0 {
		r.options.Count = 3
	}
	if r.options.Backoff <= 0 {
		r.options.Backoff = time.Second // grpc conn initial backoff is one second, use the same value here
	}
	return r
}

func (r *Retry[IN, OUT]) Call(ctx context.Context, in IN, options ...grpc.CallOption) (OUT, error) {
	var out OUT
	var err error
	for i := 0; i < r.options.Count; i++ {
		if i > 0 {
			logs.FromContext(ctx).Info("retry", slog.Int("attempts", i))
		}
		out, err = r.call(ctx, in, options...)
		if err == nil {
			return out, err
		}

		if errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			return out, err
		}

		switch GetErrorCode(err) {
		// unrecoverable error codes, return immediately
		case codes.InvalidArgument,
			codes.Unimplemented,
			codes.PermissionDenied,
			codes.Internal,
			codes.AlreadyExists,
			codes.NotFound,
			codes.DeadlineExceeded:
			return out, err
		case codes.Unauthenticated:
			if r.options.RefreshAccessToken == nil {
				return out, err
			}
			accessToken, err := r.options.RefreshAccessToken(ctx)
			if errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
				return out, err
			}
			switch GetErrorCode(err) {
			case codes.InvalidArgument,
				codes.Unimplemented,
				codes.PermissionDenied,
				codes.Internal,
				codes.AlreadyExists,
				codes.NotFound,
				codes.DeadlineExceeded,
				codes.Unauthenticated:
				return out, err
			}

			act := activity.FromOutgoingContext(ctx)
			if act == nil {
				return out, errorutil.BadRequest("no outgoing context")
			}
			act.SetAuthorization(accessToken)
		default:
			time.Sleep(r.options.Backoff)
		}
	}
	return out, err
}
