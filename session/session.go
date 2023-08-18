package session

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.olapie.com/logs"
	"go.olapie.com/ola/types"
)

const (
	keyUserID     = "$uid"
	keyStartTime  = "$st"
	keyActiveTime = "$at"
)

type Session struct {
	id      string
	storage Storage
	userID  types.UserID
}

func NewSession(id string, storage Storage) *Session {
	if id == "" {
		id = uuid.NewString()
	}

	s := &Session{
		id:      id,
		storage: storage,
	}
	return s
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) UserID() types.UserID {
	return s.userID
}

func (s *Session) SetUserID(ctx context.Context, userID types.UserID) error {
	if userID == nil {
		s.userID = nil
		return s.storage.Set(ctx, s.id, keyUserID, "")
	}

	if s.userID != nil {
		ot := reflect.TypeOf(s.userID)
		nt := reflect.TypeOf(userID)
		if ot != nt {
			logs.FromContext(ctx).Warn("different userID type",
				slog.String("old", ot.String()),
				slog.String("new", nt.String()))
		} else if s.userID != userID {
			logs.FromContext(ctx).Warn("overwriting userID value",
				slog.Any("old", s.userID),
				slog.Any("new", userID))
		}
		s.userID = userID
	}

	if i, ok := userID.Int(); ok {
		return s.SetInt64(ctx, keyUserID, i)
	} else if str, ok := userID.String(); ok {
		return s.SetString(ctx, keyUserID, str)
	} else {
		return fmt.Errorf("unsupported userID type")
	}
}

func (s *Session) SetInt64(ctx context.Context, name string, value int64) error {
	return s.storage.Set(ctx, s.id, name, strconv.FormatInt(value, 10))
}

func (s *Session) GetInt64(ctx context.Context, name string) (int64, error) {
	str, err := s.storage.Get(ctx, s.id, name)
	if err != nil {
		return 0, fmt.Errorf("cannot get from storage: %w", err)
	}
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse string %s to int64", str)
	}
	return i, nil
}

func (s *Session) Increase(ctx context.Context, name string, incr int64) (int64, error) {
	return s.storage.Increase(ctx, s.id, name, incr)
}

func (s *Session) SetString(ctx context.Context, name string, value string) error {
	return s.storage.Set(ctx, s.id, name, value)
}

func (s *Session) GetString(ctx context.Context, name string) (string, error) {
	return s.storage.Get(ctx, s.id, name)
}

func (s *Session) SetBytes(ctx context.Context, name string, value []byte) error {
	return s.storage.Set(ctx, s.id, name, string(value))
}

func (s *Session) GetBytes(ctx context.Context, name string) ([]byte, error) {
	str, err := s.storage.Get(ctx, s.id, name)
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

func (s *Session) Start(ctx context.Context) error {
	_, err := s.GetInt64(ctx, keyStartTime)
	if err != nil {
		if !errors.Is(err, ErrNoValue) {
			return fmt.Errorf("failed to get %s: %w", keyStartTime, err)
		}
		err = s.SetInt64(ctx, keyStartTime, time.Now().Unix())
		if err != nil {
			return fmt.Errorf("failed to save %s: %w", keyStartTime, err)
		}
	}
	err = s.SetInt64(ctx, keyActiveTime, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("failed to save %s: %w", keyActiveTime, err)
	}
	return nil
}

func (s *Session) Destroy(ctx context.Context) error {
	return s.storage.Destroy(ctx, s.id)
}

func SetUserID[T types.UserIDTypes](ctx context.Context, s *Session, userID T) error {
	return s.SetUserID(ctx, types.NewUserID(userID))
}

func GetUserID[T types.UserIDTypes](ctx context.Context, s *Session) T {
	var uid T
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

type ValueTypes interface {
	~int64 | ~string | ~[]byte
}

func Set[T ValueTypes](ctx context.Context, s *Session, name string, value T) error {
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

func Get[T ValueTypes](ctx context.Context, s *Session, name string) (value T, err error) {
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
