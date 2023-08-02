package session

import (
	"context"
	"fmt"
	"reflect"

	"go.olapie.com/ola/internal/delegate"
)

type ValueTypes interface {
	~int64 | ~string | ~[]byte
}

func SetUserID[T comparable](ctx context.Context, userID T) error {
	s := delegate.GetSession(ctx).(*Session)
	if s == nil {
		return ErrNotExist
	}
	return s.SetUserID(ctx, userID)
}

func GetUserID[T any](ctx context.Context) T {
	var uid T

	s := delegate.GetSession(ctx).(*Session)
	if s == nil {
		return uid
	}

	v, ok := s.userID.(T)
	if ok {
		return v
	}

	var resType = reflect.TypeOf(uid)
	if reflect.TypeOf(s.userID).ConvertibleTo(reflect.TypeOf(uid)) {
		uid, _ = reflect.ValueOf(s.userID).Convert(resType).Interface().(T)
	}

	return uid
}

func Set[T ValueTypes](ctx context.Context, name string, value T) error {
	s := delegate.GetSession(ctx).(*Session)
	if s == nil {
		return ErrNotExist
	}

	switch v := any(value).(type) {
	case int64:
		return s.SetInt64(ctx, name, v)
	case string:
		return s.SetString(ctx, name, v)
	case []byte:
		return s.SetBytes(ctx, name, v)
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Int64:
			return s.SetInt64(ctx, name, rv.Int())
		case reflect.String:
			return s.SetString(ctx, name, rv.String())
		default:
			if rv.Type().ConvertibleTo(reflect.TypeOf([]byte(nil))) {
				return s.SetBytes(ctx, name, rv.Bytes())
			}
			return fmt.Errorf("unsupported type %T", value)
		}
	}
}

func Get[T ValueTypes](ctx context.Context, name string) (value T, err error) {
	s := delegate.GetSession(ctx).(*Session)
	if s == nil {
		return value, ErrNotExist
	}

	switch any(value).(type) {
	case int64:
		i, err := s.GetInt64(ctx, name)
		if err != nil {
			return value, err
		}
		reflect.ValueOf(value).SetInt(i)
	case string:
		str, err := s.GetString(ctx, name)
		if err != nil {
			return value, err
		}
		reflect.ValueOf(value).SetString(str)
	case []byte:
		b, err := s.GetBytes(ctx, name)
		if err != nil {
			return value, err
		}
		reflect.ValueOf(value).SetBytes(b)
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Int64:
			i, err := s.GetInt64(ctx, name)
			if err != nil {
				return value, err
			}
			reflect.ValueOf(value).SetInt(i)
		case reflect.String:
			str, err := s.GetString(ctx, name)
			if err != nil {
				return value, err
			}
			reflect.ValueOf(value).SetString(str)
		default:
			if rv.Type().ConvertibleTo(reflect.TypeOf([]byte(nil))) {
				b, err := s.GetBytes(ctx, name)
				if err != nil {
					return value, err
				}
				reflect.ValueOf(value).SetBytes(b)
			} else {
				err = fmt.Errorf("unsupported type %T", value)
			}
		}
	}
	return
}
