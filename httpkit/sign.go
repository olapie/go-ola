package httpkit

import (
	"net/http"

	"go.olapie.com/logs"
	"go.olapie.com/ola/activity"
	"go.olapie.com/ola/headers"
	"go.olapie.com/security/base62"
)

func SignRequest(req *http.Request, createAPIKey func(h http.Header)) {
	a := activity.FromOutgoingContext(req.Context())
	if a != nil {
		activity.CopyHeader(req.Header, a)
	} else {
		logs.FromContext(req.Context()).Warn("no outgoing context")
	}
	if traceID := headers.GetTraceID(req.Header); traceID == "" {
		if traceID == "" {
			traceID = base62.NewUUIDString()
			logs.FromContext(req.Context()).Info("generated trace id " + traceID)
		}
		headers.SetTraceID(req.Header, traceID)
	}
	createAPIKey(req.Header)
}
