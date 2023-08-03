package activity

import (
	"context"
	"log/slog"
	"reflect"
	"time"

	"go.olapie.com/logs"
	"go.olapie.com/ola/session"
	"go.olapie.com/ola/types"
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

func SetUserID[T types.UserIDTypes](ctx context.Context, id T) error {
	a := FromContext(ctx)
	if a == nil {
		return ErrNotExist
	}
	a.UserID = types.NewUserID(id)
	return nil
}

func GetUserID[T types.UserIDTypes](ctx context.Context) T {
	var id T
	a := FromContext(ctx)
	if a == nil {
		slog.Warn("no activity")
		return id
	}

	v := a.UserID.Value()
	if id, ok := a.UserID.Value().(T); ok {
		return id
	}

	t := reflect.TypeOf(v)
	idType := reflect.TypeOf(id)
	if t.ConvertibleTo(reflect.TypeOf(id)) {
		id, _ = reflect.ValueOf(v).Convert(idType).Interface().(T)
	}

	return id
}
