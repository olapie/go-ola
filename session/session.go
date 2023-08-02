package session

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.olapie.com/logs"
)

const (
	keyUserID     = "$uid"
	keyStartTime  = "$start_time"
	keyActiveTime = "$active_time"
)

type Session struct {
	id         string
	storage    Storage
	userID     any
	createTime time.Time
	activeTime time.Time
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

func (s *Session) UserID() any {
	return s.userID
}

func (s *Session) SetUserID(ctx context.Context, userID any) error {
	if userID == nil {
		s.userID = nil
		return s.storage.Set(ctx, s.id, keyUserID, "")
	}

	if s.userID != nil {
		ot := reflect.TypeOf(s.userID)
		nt := reflect.TypeOf(userID)
		if ot != nt {
			logs.FromCtx(ctx).Warn("different userID type",
				slog.String("old", ot.String()),
				slog.String("new", nt.String()))
		} else if s.userID != userID {
			logs.FromCtx(ctx).Warn("overwriting userID value",
				slog.Any("old", s.userID),
				slog.Any("new", userID))
		}
		s.userID = userID
	}

	if s.storage == nil {
		return nil
	}

	switch v := userID.(type) {
	case int64:
		return s.SetInt64(ctx, keyUserID, v)
	case string:
		return s.SetString(ctx, keyUserID, v)
	default:
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
