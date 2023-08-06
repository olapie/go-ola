package activity

import (
	"go.olapie.com/ola/session"
	"google.golang.org/grpc/metadata"
	"net/http"
	"time"

	"go.olapie.com/ola/types"
)

type Activity struct {
	Name         string
	TraceID      string
	UserID       types.UserID
	StartTime    time.Time
	HTTPHeader   http.Header
	GRPCMetadata metadata.MD

	Session *session.Session
}
