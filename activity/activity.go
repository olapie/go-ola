package activity

import (
	"google.golang.org/grpc/metadata"
	"net/http"
	"time"

	"go.olapie.com/ola/session"
	"go.olapie.com/ola/types"
)

type Activity struct {
	Name         string
	TraceID      string
	UserID       types.UserID
	Session      *session.Session
	StartTime    time.Time
	HTTPRequest  *http.Request
	GRPCMetadata metadata.MD
}
