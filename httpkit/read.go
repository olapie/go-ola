package httpkit

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"go.olapie.com/logs"
)

func ReadJSONBody(rw http.ResponseWriter, req *http.Request, ptrToModel any) bool {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		Error(rw, err)
		logs.FromCtx(req.Context()).Error("read request body", slog.String("err", err.Error()))
		return false
	}
	err = json.Unmarshal(body, ptrToModel)
	if err != nil {
		Error(rw, err)
		logs.FromCtx(req.Context()).Error("unmarshal json", slog.String("err", err.Error()))
		return false
	}
	return true
}
