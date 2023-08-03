package httpkit

import (
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"go.olapie.com/utils"
)

func BasicAuthenticate(w http.ResponseWriter, realm string) {
	a := "Basic realm=" + strconv.Quote(realm)
	w.Header().Set(KeyWWWAuthenticate, a)
	w.WriteHeader(http.StatusUnauthorized)
}

func Error(w http.ResponseWriter, err error) {
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	status := http.StatusInternalServerError
	if s := utils.GetErrorCode(err); s != 0 {
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
	SetContentType(w.Header(), MimeJsonUTF8)
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
	SetContentType(w.Header(), MimeOctetStream)
	_, err := w.Write(b)
	if err != nil {
		slog.Error("cannot write", "err", err.Error())
	}
}

func HTMLFile(w http.ResponseWriter, s string) {
	SetContentType(w.Header(), MimeHtmlUTF8)
	_, err := w.Write([]byte(s))
	if err != nil {
		slog.Error("cannot write", "err", err.Error())
	}
}

func CSSFile(w http.ResponseWriter, s string) {
	SetContentType(w.Header(), MimeCSS)
	_, err := w.Write([]byte(s))
	if err != nil {
		slog.Error("cannot write", "err", err.Error())
	}
}

func JSFile(w http.ResponseWriter, s string) {
	SetContentType(w.Header(), MimeJavascript)
	_, err := w.Write([]byte(s))
	if err != nil {
		slog.Error("cannot write", "err", err.Error())
	}
}

func StreamFile(w http.ResponseWriter, name string, f io.ReadCloser) {
	defer f.Close()
	SetContentType(w.Header(), MimeOctetStream)
	if name != "" {
		w.Header().Set(KeyContentDisposition, ToAttachment(name))
	}
	_, err := io.Copy(w, f)
	if err != nil {
		if err != io.EOF {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
