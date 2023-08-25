package httpkit

import (
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"go.olapie.com/ola/mimetypes"

	"go.olapie.com/ola/headers"

	"go.olapie.com/ola/errorutil"
)

func BasicAuthenticate(w http.ResponseWriter, realm string) {
	a := "Basic realm=" + strconv.Quote(realm)
	w.Header().Set(headers.KeyWWWAuthenticate, a)
	w.WriteHeader(http.StatusUnauthorized)
}

func Error(w http.ResponseWriter, err error) {
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	status := http.StatusInternalServerError
	if s := errorutil.GetCode(err); s != 0 {
		status = s
	}

	if status < 100 || status > 599 {
		log.Println("invalid status:", status)
		status = http.StatusInternalServerError
	}

	w.WriteHeader(status)
	_, err = w.Write([]byte(err.Error()))
	if err != nil {
		log.Println(err)
	}
}

func JSON(w http.ResponseWriter, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	headers.SetContentType(w.Header(), mimetypes.JsonUTF8)
	_, err = w.Write(data)
	if err != nil {
		log.Println(err)
	}
}

func JSONOrError(w http.ResponseWriter, v any, err error) {
	if err != nil {
		Error(w, err)
	} else {
		JSON(w, v)
	}
}

func OctetStream(w http.ResponseWriter, b []byte) {
	headers.SetContentType(w.Header(), mimetypes.OctetStream)
	_, err := w.Write(b)
	if err != nil {
		slog.Error("cannot write", "err", err.Error())
	}
}

func HTMLFile(w http.ResponseWriter, s string) {
	headers.SetContentType(w.Header(), mimetypes.HtmlUTF8)
	_, err := w.Write([]byte(s))
	if err != nil {
		slog.Error("cannot write", "err", err.Error())
	}
}

func CSSFile(w http.ResponseWriter, s string) {
	headers.SetContentType(w.Header(), mimetypes.CSS)
	_, err := w.Write([]byte(s))
	if err != nil {
		slog.Error("cannot write", "err", err.Error())
	}
}

func JSFile(w http.ResponseWriter, s string) {
	headers.SetContentType(w.Header(), mimetypes.Javascript)
	_, err := w.Write([]byte(s))
	if err != nil {
		slog.Error("cannot write", "err", err.Error())
	}
}

func StreamFile(w http.ResponseWriter, name string, f io.ReadCloser) {
	defer f.Close()
	headers.SetContentType(w.Header(), mimetypes.OctetStream)
	if name != "" {
		w.Header().Set(headers.KeyContentDisposition, headers.ToAttachment(name))
	}
	_, err := io.Copy(w, f)
	if err != nil {
		if err != io.EOF {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
