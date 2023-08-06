package httpkit

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

type Upload struct {
	Filename string
	Size     int64
	Content  []byte
}

func NewUploadHandler(maxMemory int64, store func(*Upload) error) http.Handler {
	if maxMemory <= 0 {
		maxMemory = 10 * (1 << 20)
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		err := req.ParseMultipartForm(maxMemory)
		if err != nil {
			slog.Error("cannot parse multipart-form", slog.String("err", err.Error()))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		for _, fileHeaders := range req.MultipartForm.File {
			for _, fh := range fileHeaders {
				if upload, err := processMultipartFileHeader(fh); err != nil {
					slog.Error("cannot process multipart file header", slog.String("err", err.Error()))
					rw.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					if err = store(upload); err != nil {
						slog.Error("cannot store upload", slog.String("err", err.Error()))
						rw.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
			}
		}
	})
}

func processMultipartFileHeader(h *multipart.FileHeader) (*Upload, error) {
	f, err := h.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}
	upload := &Upload{
		Filename: filepath.Base(h.Filename),
		Size:     h.Size,
		Content:  content,
	}
	return upload, nil
}
