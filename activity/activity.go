package activity

import (
	"time"

	"go.olapie.com/ola/session"
	"go.olapie.com/ola/types"
)

type Activity struct {
	Name      string
	TraceID   string
	UserID    types.UserID
	Session   *session.Session
	StartTime time.Time
	Request   any
}
