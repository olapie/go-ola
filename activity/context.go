package activity

import (
	"context"
	"log/slog"
	"reflect"
	"time"

	"go.olapie.com/logs"
	"go.olapie.com/ola/types"
)

type contextKey int

const contextKeyActivity contextKey = iota

func NewContext(ctx context.Context, a *Activity) context.Context {
	if v, _ := ctx.Value(contextKeyActivity).(*Activity); v != nil {
		logs.FromCtx(ctx).Warn("skipped new context with activity as it already exists")
		return ctx
	}
	return context.WithValue(ctx, contextKeyActivity, a)
}

func FromContext(ctx context.Context) *Activity {
	v := ctx.Value(contextKeyActivity)
	if v == nil {
		return &Activity{
			StartTime: time.Now(),
		}
	}
	return ctx.Value(contextKeyActivity).(*Activity)
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
