package activity

import (
	"time"

	"go.olapie.com/ola/session"
)

type Activity struct {
	Name      string
	TraceID   string
	UserID    UserID
	Session   *session.Session
	StartTime time.Time
	Request   any
}
