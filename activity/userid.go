package activity

import (
	"context"
	"log/slog"
	"reflect"
)

type UserID interface {
	Int() (int64, bool)
	String() (string, bool)
}

type UserIDTypes interface {
	~int64 | ~string
}

func NewUserID[T UserIDTypes](id T) UserID {
	return &userIDHolder[T]{id: id}
}

func SetUserID[T UserIDTypes](ctx context.Context, id T) error {
	a := FromContext(ctx)
	if a == nil {
		return ErrNotExist
	}
	a.UserID = NewUserID(id)
	return nil
}

func GetUserID[T UserIDTypes](ctx context.Context) T {
	var id T
	a := FromContext(ctx)
	if a == nil {
		slog.Warn("no activity")
		return id
	}
	v, ok := a.UserID.(*userIDHolder[T])
	if ok {
		return v.id
	}

	var t = reflect.TypeOf(id)
	if reflect.TypeOf(v.id).ConvertibleTo(reflect.TypeOf(id)) {
		id, _ = reflect.ValueOf(v.id).Convert(t).Interface().(T)
	}

	return id
}

type userIDHolder[T UserIDTypes] struct {
	id T
}

func (u *userIDHolder[T]) Int() (int64, bool) {
	i, ok := any(u.id).(int64)
	return i, ok
}

func (u *userIDHolder[T]) String() (string, bool) {
	s, ok := any(u.id).(string)
	return s, ok
}
