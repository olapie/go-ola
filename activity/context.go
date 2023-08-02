package activity

import (
	"context"
	"time"

	"go.olapie.com/logs"
	"go.olapie.com/ola/session"
)

type contextKey int

const contextKeyActivity contextKey = iota

func NewContext(ctx context.Context, a *Activity) context.Context {
	if FromContext(ctx) != nil {
		logs.FromCtx(ctx).Warn("skipped new context with activity as it already exists")
		return ctx
	}
	return context.WithValue(ctx, contextKeyActivity, a)
}

func FromContext(ctx context.Context) *Activity {
	s, _ := ctx.Value(contextKeyActivity).(*Activity)
	return s
}

func GetSession(ctx context.Context) *session.Session {
	a := FromContext(ctx)
	if a == nil {
		logs.FromCtx(ctx).Warn("no activity")
		return nil
	}
	return a.Session
}

func GetTraceID(ctx context.Context) string {
	a := FromContext(ctx)
	if a == nil {
		logs.FromCtx(ctx).Warn("no activity")
		return ""
	}
	return a.TraceID
}

func GetStartTime(ctx context.Context) time.Time {
	a := FromContext(ctx)
	if a == nil {
		logs.FromCtx(ctx).Warn("no activity")
		return time.Time{}
	}
	return a.StartTime
}

func GetName(ctx context.Context) string {
	a := FromContext(ctx)
	if a == nil {
		logs.FromCtx(ctx).Warn("no activity")
		return ""
	}
	return a.Name
}
