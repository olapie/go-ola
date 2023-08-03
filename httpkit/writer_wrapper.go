package httpkit

import (
	"bufio"
	"errors"
	"log/slog"
	"net"
	"net/http"
)

type statusGetter interface {
	Status() int
}

// http.Flusher doesn't return error, however gzip.Writer/deflate.WriterWrapper only implement `Flush() error`
type flusher interface {
	Flush() error
}

var (
	_ statusGetter        = (*WriterWrapper)(nil)
	_ http.Hijacker       = (*WriterWrapper)(nil)
	_ http.Flusher        = (*WriterWrapper)(nil)
	_ http.ResponseWriter = (*WriterWrapper)(nil)
)

// WriterWrapper is a wrapper of http.ResponseWriter to make sure write status code only one time
type WriterWrapper struct {
	http.ResponseWriter
	status int
	body   []byte
}

func WrapWriter(rw http.ResponseWriter) *WriterWrapper {
	if w, ok := rw.(*WriterWrapper); ok {
		return w
	}
	return &WriterWrapper{
		ResponseWriter: rw,
	}
}

func (w *WriterWrapper) WriteHeader(statusCode int) {
	if statusCode < http.StatusContinue {
		slog.Error("cannot write invalid status code", "code", statusCode)
		statusCode = http.StatusInternalServerError
	}
	if w.status > 0 {
		slog.Warn("status code already written")
		return
	}
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *WriterWrapper) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.body = data
	return w.ResponseWriter.Write(data)
}

func (w *WriterWrapper) Status() int {
	return w.status
}

func (w *WriterWrapper) Body() []byte {
	return w.body
}

func (w *WriterWrapper) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *WriterWrapper) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *WriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("hijack not supported")
}

func (w *WriterWrapper) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
