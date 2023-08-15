package activity

import (
	"context"
	"log/slog"
	"reflect"

	"go.olapie.com/ola/types"
)

type activityIncomingContext struct{}
type activityOutgoingContext struct{}

func FromIncomingContext(ctx context.Context) *Activity {
	a, _ := ctx.Value(activityIncomingContext{}).(*Activity)
	return a
}

func FromOutgoingContext(ctx context.Context) *Activity {
	a, _ := ctx.Value(activityOutgoingContext{}).(*Activity)
	return a
}

func NewIncomingContext(ctx context.Context, a *Activity) context.Context {
	return context.WithValue(ctx, activityIncomingContext{}, a)
}

func NewOutgoingContext(ctx context.Context, a *Activity) context.Context {
	return context.WithValue(ctx, activityOutgoingContext{}, a)
}

func SetIncomingUserID[T types.UserIDTypes](ctx context.Context, id T) error {
	a := FromIncomingContext(ctx)
	if a == nil {
		return ErrNotExist
	}
	a.SetUserID(types.NewUserID(id))
	return nil
}

func GetIncomingUserID[T types.UserIDTypes](ctx context.Context) T {
	var id T
	a := FromIncomingContext(ctx)
	if a == nil {
		slog.Warn("no activityImpl")
		return id
	}

	v := a.UserID().Value()
	if id, ok := a.UserID().Value().(T); ok {
		return id
	}

	t := reflect.TypeOf(v)
	idType := reflect.TypeOf(id)
	if t.ConvertibleTo(reflect.TypeOf(id)) {
		id, _ = reflect.ValueOf(v).Convert(idType).Interface().(T)
	}

	return id
}
