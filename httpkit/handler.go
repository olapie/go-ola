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

func InterceptHandler[T ~int64 | ~string](next http.Handler, timeout time.Duration, authenticator Authenticator[T]) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		startAt := time.Now()
		a := activity.New("", req.Header)
		ctx := activity.NewIncomingContext(req.Context(), a)
		traceID := a.Get(headers.KeyTraceID)
		if traceID == "" {
			traceID = base62.NewUUIDString()
		}
		logger := logs.FromContext(ctx).With(slog.String("trace_id", traceID))
		fields := make([]any, 0, 4+len(req.Header))
		fields = append(fields,
			slog.String("uri", req.RequestURI),
			slog.String("method", req.Method),
			slog.String("host", req.Host),
			slog.String("remote_addr", req.RemoteAddr))
		for key := range req.Header {
			fields = append(fields, slog.String(key, req.Header.Get(key)))
		}

		logger.Info("start", fields...)
		ctx = logs.NewCtx(ctx, logger)
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		req = req.WithContext(ctx)

		defer func() {
			if p := recover(); p != nil {
				slog.Error("panic", "error", p)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
		}()

		w := WrapWriter(rw)
		if headers.VerifyAPIKey(req.Header, 10) {
			if authenticator != nil {
				authResult, err := authenticator.Authenticate(req)
				if err != nil {
					logger.Error("auth failed", slog.String("error", err.Error()))
					Error(w, err)
				} else {
					logger.Info("auth succeeded", slog.Any("login", authResult))
					a.SetUserID(types.NewUserID(authResult))
					next.ServeHTTP(w, req)
				}
			} else {
				next.ServeHTTP(w, req)
			}
		} else {
			Error(w, errorutil.NewError(http.StatusBadRequest, "invalid api key"))
		}

		status := w.Status()
		fields = []any{slog.Int("status", status),
			slog.Duration("cost", time.Now().Sub(startAt))}
		if status >= 400 {
			fields = append(fields, slog.String("body", string(w.Body())))
			logger.Info("failed", fields...)
		} else {
			logger.Info("finished", fields...)
		}
	})
}
