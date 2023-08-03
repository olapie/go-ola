package httpkit

import (
	"compress/flate"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

var (
	_ http.Flusher = (*Compressor)(nil)
)

type Compressor struct {
	http.ResponseWriter
	compressWriter io.Writer
	err            error
	hasBody        bool
}

var _ http.ResponseWriter = (*Compressor)(nil)

func NewCompressor(w http.ResponseWriter, encodings ...string) (*Compressor, error) {
	if cw, ok := w.(*Compressor); ok {
		log.Println("cannot compress a compressed writer")
		return cw, nil
	}

	if len(encodings) == 0 {
		return nil, errors.New("missing encodings")
	}

	var err error
	for _, encoding := range encodings {
		cw, er := newCompressedWriter(w, encoding)
		if er == nil {
			return cw, nil
		}
		err = errors.Join(err, er)
	}

	return nil, err
}

func newCompressedWriter(w http.ResponseWriter, encoding string) (*Compressor, error) {
	switch encoding {
	case "gzip":
		cw := &Compressor{}
		cw.ResponseWriter = w
		cw.compressWriter = gzip.NewWriter(w)
		SetContentEncoding(w.Header(), encoding)
		return cw, nil
	case "deflate":
		fw, err := flate.NewWriter(w, flate.DefaultCompression)
		if err != nil {
			return nil, fmt.Errorf("new flate writer: %w", err)
		}
		cw := &Compressor{}
		cw.compressWriter = fw
		cw.ResponseWriter = w
		SetContentEncoding(w.Header(), encoding)
		return cw, nil
	default:
		return nil, errors.New("unsupported encoding")
	}
}

func (w *Compressor) Write(data []byte) (int, error) {
	if !w.hasBody {
		w.hasBody = len(data) > 0
	}
	return w.compressWriter.Write(data)
}

func (w *Compressor) Flush() {
	// Flush the compressed writer, then flush httpWriter
	if f, ok := w.compressWriter.(flusher); ok {
		if err := f.Flush(); err != nil {
			log.Println("flush", err)
			w.err = err
		}

		if f, ok = w.ResponseWriter.(flusher); ok {
			if err := f.Flush(); err != nil {
				log.Println("flush", err)
				w.err = err
			}
		}

	}
}

func (w *Compressor) Error() error {
	return w.err
}

func (w *Compressor) Close() error {
	if !w.hasBody {
		if f, ok := w.ResponseWriter.(flusher); ok {
			f.Flush()
		}
		return nil
	}
	if closer, ok := w.compressWriter.(io.Closer); ok {
		// Closing a writer without written data will cause an error if Response status is 204 NoContent
		return closer.Close()
	}
	return nil
}
