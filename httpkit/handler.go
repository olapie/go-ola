package httpkit

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.olapie.com/logs"
	"go.olapie.com/ola/activity"
	"go.olapie.com/ola/errorutil"
	"go.olapie.com/ola/headers"
	"go.olapie.com/ola/types"
	"go.olapie.com/security/base62"
)

type joinHandler struct {
	handlers []http.Handler
}

func (j *joinHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	w := WrapWriter(writer)
	for _, h := range j.handlers {
		h.ServeHTTP(w, request)
		if w.Status() != 0 {
			return
		}
	}
}

var _ http.Handler = (*joinHandler)(nil)

func JoinHandlers(handlers ...http.Handler) http.Handler {
	return &joinHandler{
		handlers: handlers,
	}
}

func JoinHandlerFuncs(funcs ...http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		w := WrapWriter(writer)
		for _, f := range funcs {
			f.ServeHTTP(w, request)
			if w.Status() != 0 {
				return
			}
		}
	}
}

type Authenticator[T types.UserIDTypes] interface {
	Authenticate(req *http.Request) (T, error)
}

func NewStartHandler(
	maybeNext http.Handler,
	verifyAPIKey func(ctx context.Context, header http.Header) bool,
	authenticate func(ctx context.Context, header http.Header) *types.Auth) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		startAt := time.Now()
		a := activity.New("", req.Header)
		ctx := activity.NewIncomingContext(req.Context(), a)
		traceID := a.Get(headers.KeyTraceID)
		if traceID == "" {
			traceID = base62.NewUUIDString()
		}

		logger := logs.FromContext(ctx).With(slog.String("traceId", traceID))

		fields := make([]any, 0, 4+len(req.Header))
		fields = append(fields,
			slog.String("uri", req.RequestURI),
			slog.String("method", req.Method),
			slog.String("host", req.Host),
			slog.String("remoteAddr", req.RemoteAddr))
		for key := range req.Header {
			fields = append(fields, slog.String(key, req.Header.Get(key)))
		}

		ctx = logs.NewContext(ctx, logger)
		logger = logger.With("module", "httpkit")
		logger.Info("START", fields...)
		var cancel context.CancelFunc
		if seconds := a.GetRequestTimeout(); seconds > 0 {
			ctx, cancel = context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
			defer cancel()
		}

		req = req.WithContext(ctx)
		w := WrapWriter(rw)

		defer func() {
			if p := recover(); p != nil {
				logger.Error("panic", "error", p)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}

			status := w.Status()
			fields = []any{slog.Int("status", status),
				slog.Duration("cost", time.Now().Sub(startAt))}
			if status >= 400 {
				fields = append(fields, slog.String("body", string(w.Body())))
				logger.Error("END", fields...)
			} else {
				logger.Info("END", fields...)
			}
		}()

		appID := a.GetAppID()
		if appID == "" {
			Error(w, errorutil.NewError(http.StatusBadRequest, "client appId does not match authenticated appId"))
			return
		}

		if verifyAPIKey(ctx, req.Header) {
			auth := authenticate(ctx, req.Header)
			if auth != nil {
				if auth.AppID != appID {
					Error(w, errorutil.NewError(http.StatusUnauthorized, "client appId does not match authenticated appId"))
					return
				} else {
					a.SetUserID(auth.UserID)
					logger.Info("authenticated", slog.Any("uid", auth.UserID.Value()), slog.String("appId", auth.AppID))
				}
			}
			maybeNext.ServeHTTP(w, req)
		} else {
			Error(w, errorutil.NewError(http.StatusBadRequest, "invalid api key"))
			return
		}
	})
}
