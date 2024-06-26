package activity

import (
	"context"
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
		return id
	}

	if a.userID == nil {
		return id
	}

	v := a.userID.Value()
	if id, ok := v.(T); ok {
		return id
	}

	t := reflect.TypeOf(v)
	idType := reflect.TypeOf(id)
	if t.ConvertibleTo(reflect.TypeOf(id)) {
		id, _ = reflect.ValueOf(v).Convert(idType).Interface().(T)
	}

	return id
}

var systemUserID = types.NewUserID[string]("ola-system-user-id")

func SetSystemUser(ctx context.Context) {
	a := FromIncomingContext(ctx)
	if a == nil {
		a = new(Activity)
		NewIncomingContext(ctx, a)
	}
	a.userID = systemUserID
}

func IsSystemUser(ctx context.Context) bool {
	a := FromIncomingContext(ctx)
	if a == nil {
		return false
	}
	return a.userID == systemUserID
}
