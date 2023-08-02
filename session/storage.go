package session

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Storage interface {
	Set(ctx context.Context, sid, name string, value string) error
	Get(ctx context.Context, sid, name string) (string, error)
	Increase(ctx context.Context, sid, name string, incr int64) (int64, error)
	SetTTL(ctx context.Context, sid string, ttl time.Duration) error
	Destroy(ctx context.Context, sid string) error
}

var _ Storage = (*LocalStorage)(nil)

type entry struct {
	m         sync.Map
	expiresAt time.Time
}

type LocalStorage struct {
	m sync.Map // string:*sync.Map
}

func (l *LocalStorage) getOrCreate(ctx context.Context, sid string) *entry {
	e, ok := l.m.Load(sid)
	if ok {
		return e.(*entry)
	}

	e, _ = l.m.LoadOrStore(sid, new(entry))
	return e.(*entry)
}

func (l *LocalStorage) Set(ctx context.Context, sid, name string, value string) error {
	l.getOrCreate(ctx, sid).m.Store(name, value)
	return nil
}

func (l *LocalStorage) Get(ctx context.Context, sid, name string) (string, error) {
	e, ok := l.m.Load(sid)
	if !ok {
		return "", ErrNotExist
	}
	v, ok := e.(*entry).m.Load(name)
	if !ok {
		return "", ErrNoValue
	}
	if str, ok := v.(string); ok {
		return str, nil
	}
	return "", ErrNoValue
}

func (l *LocalStorage) Increase(ctx context.Context, sid, name string, incr int64) (int64, error) {
	e := l.getOrCreate(ctx, sid)
	for {
		old, ok := e.m.Load(name)
		var i int64
		if ok {
			var err error
			i, err = strconv.ParseInt(old.(string), 10, 64)
			if err != nil {
				return 0, fmt.Errorf("cannot parse %s to int64: %w", old, err)
			}
		}

		i += incr
		newValue := strconv.FormatInt(i, 10)
		if e.m.CompareAndSwap(name, old, newValue) {
			return i, nil
		}
	}
}

func (l *LocalStorage) SetTTL(ctx context.Context, sid string, ttl time.Duration) error {
	e, ok := l.m.Load(sid)
	if !ok {
		return ErrNotExist
	}
	e.(*entry).expiresAt = time.Now().Add(ttl)
	return nil
}

func (l *LocalStorage) Destroy(ctx context.Context, sid string) error {
	l.m.Delete(sid)
	return nil
}
