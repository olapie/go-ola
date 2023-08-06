package activity

import (
	"net/http"
	"time"

	"go.olapie.com/ola/session"
	"go.olapie.com/ola/types"
)

type Activity struct {
	Name        string
	TraceID     string
	UserID      types.UserID
	Session     *session.Session
	StartTime   time.Time
	HTTPRequest *http.Request
}
