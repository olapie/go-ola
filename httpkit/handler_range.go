package httpkit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"

	"go.olapie.com/ola/internal/types"
)

type Content interface {
	io.ReadSeeker
	Type() string
	Length() int64
}

type contentImpl struct {
	io.ReadSeeker
	typ  string
	size int64
}

func (c *contentImpl) Type() string {
	return c.typ
}

func (c *contentImpl) Length() int64 {
	return c.size
}

func NewContent(data []byte, typ string) Content {
	return &contentImpl{
		ReadSeeker: bytes.NewReader(data),
		typ:        typ,
		size:       int64(len(data)),
	}
}

func NewRangeHandler(f func(req *http.Request) (Content, error)) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		content, err := f(req)
		if err != nil {
			Error(rw, err)
			return
		}
		HandleRangeRequest(rw, req, content)
	})
}

func HandleRangeRequest(rw http.ResponseWriter, req *http.Request, content Content) {
	contentType, size := content.Type(), content.Length()
	sendSize := content.Length()
	var sendContent io.Reader = content
	code := http.StatusOK
	ranges, err := parseByteRanges(req.Header.Get("Range"), size)
	switch err {
	case nil:
	case errNoOverlap:
		if size == 0 {
			ranges = nil
			break
		}
		rw.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
		fallthrough
	default:
		http.Error(rw, err.Error(), http.StatusRequestedRangeNotSatisfiable)
		return
	}

	if sumRangesSize(ranges) > size {
		ranges = nil
	}
	switch {
	case len(ranges) == 1:
		ra := ranges[0]
		if _, err := content.Seek(ra.start, io.SeekStart); err != nil {
			http.Error(rw, err.Error(), http.StatusRequestedRangeNotSatisfiable)
			return
		}
		sendSize = ra.length
		code = http.StatusPartialContent
		rw.Header().Set("Content-Type", contentType)
		rw.Header().Set("Content-Range", ra.contentRange(size))
	case len(ranges) > 1:
		sendSize = rangesMIMESize(ranges, contentType, size)
		code = http.StatusPartialContent

		pr, pw := io.Pipe()
		mw := multipart.NewWriter(pw)
		rw.Header().Set("Content-Type", "multipart/byteranges; boundary="+mw.Boundary())
		sendContent = pr
		defer pr.Close() // cause writing goroutine to fail and exit if CopyN doesn't finish.
		go func() {
			for _, ra := range ranges {
				part, err := mw.CreatePart(ra.mimeHeader(contentType, size))
				if err != nil {
					pw.CloseWithError(err)
					return
				}
				if _, err := content.Seek(ra.start, io.SeekStart); err != nil {
					pw.CloseWithError(err)
					return
				}
				if _, err := io.CopyN(part, content, ra.length); err != nil {
					pw.CloseWithError(err)
					return
				}
			}
			mw.Close()
			pw.Close()
		}()
	}

	rw.Header().Set("Accept-Ranges", "bytes")
	if rw.Header().Get("Content-Encoding") == "" {
		rw.Header().Set("Content-Length", strconv.FormatInt(sendSize, 10))
	}

	rw.WriteHeader(code)
	if req.Method != "HEAD" {
		io.CopyN(rw, sendContent, sendSize)
	}
}

type Range struct {
	Start int64
	End   int64
}

// ParseRanges parses a Range header string as per RFC 7233.
// errNoOverlap is returned if none of the ranges overlap.
func ParseRanges(req *http.Request) ([]Range, error) {
	s := req.Header.Get("Range")
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}
	var ranges []Range
	for _, ra := range strings.Split(s[len(b):], ",") {
		ra = textproto.TrimString(ra)
		if ra == "" {
			continue
		}
		start, end, ok := strings.Cut(ra, "-")
		if !ok {
			return nil, errors.New("invalid range")
		}
		start, end = textproto.TrimString(start), textproto.TrimString(end)
		var r Range
		if start == "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file,
			// and we are dealing with <suffix-length>
			// which has to be a non-negative integer as per
			// RFC 7233 Section 2.1 "Byte-Ranges".
			if end == "" || end[0] == '-' {
				return nil, errors.New("invalid range")
			}
			i, err := strconv.ParseInt(end, 10, 64)
			if i < 0 || err != nil {
				return nil, errors.New("invalid range")
			}
			r.End = i
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i < 0 {
				return nil, errors.New("invalid range")
			}

			r.Start = i
			if end != "" {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.Start > i {
					return nil, errors.New("invalid range")
				}
				r.End = i
			}
		}
		ranges = append(ranges, r)
	}
	return ranges, nil
}

// errNoOverlap is returned by serveContent's parseByteRanges if first-byte-pos of
// all the byte-range-spec values is greater than the content size.
const errNoOverlap = types.ErrorString("invalid range: failed to overlap")

// byteRange specifies the byte range to be sent to the client.
type byteRange struct {
	start, length int64
}

func (r *byteRange) Start() int64 {
	return r.start
}

func (r *byteRange) Length() int64 {
	return r.length
}

func (r *byteRange) contentRange(size int64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.start, r.start+r.length-1, size)
}

func (r *byteRange) mimeHeader(contentType string, size int64) textproto.MIMEHeader {
	return textproto.MIMEHeader{
		"Content-Range": {r.contentRange(size)},
		"Content-Type":  {contentType},
	}
}

// parseByteRanges parses a Range header string as per RFC 7233.
// errNoOverlap is returned if none of the ranges overlap.
func parseByteRanges(s string, size int64) ([]byteRange, error) {
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}
	var ranges []byteRange
	noOverlap := false
	for _, ra := range strings.Split(s[len(b):], ",") {
		ra = textproto.TrimString(ra)
		if ra == "" {
			continue
		}
		start, end, ok := strings.Cut(ra, "-")
		if !ok {
			return nil, errors.New("invalid range")
		}
		start, end = textproto.TrimString(start), textproto.TrimString(end)
		var r byteRange
		if start == "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file,
			// and we are dealing with <suffix-length>
			// which has to be a non-negative integer as per
			// RFC 7233 Section 2.1 "Byte-Ranges".
			if end == "" || end[0] == '-' {
				return nil, errors.New("invalid range")
			}
			i, err := strconv.ParseInt(end, 10, 64)
			if i < 0 || err != nil {
				return nil, errors.New("invalid range")
			}
			if i > size {
				i = size
			}
			r.start = size - i
			r.length = size - r.start
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i < 0 {
				return nil, errors.New("invalid range")
			}
			if i >= size {
				// If the range begins after the size of the content,
				// then it does not overlap.
				noOverlap = true
				continue
			}
			r.start = i
			if end == "" {
				// If no end is specified, range extends to end of the file.
				r.length = size - r.start
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.start > i {
					return nil, errors.New("invalid range")
				}
				if i >= size {
					i = size - 1
				}
				r.length = i - r.start + 1
			}
		}
		ranges = append(ranges, r)
	}
	if noOverlap && len(ranges) == 0 {
		// The specified ranges did not overlap with the content.
		return nil, errNoOverlap
	}
	return ranges, nil
}

func sumRangesSize(ranges []byteRange) (size int64) {
	for _, ra := range ranges {
		size += ra.length
	}
	return
}

// rangesMIMESize returns the number of bytes it takes to encode the
// provided ranges as a multipart response.
func rangesMIMESize(ranges []byteRange, contentType string, contentSize int64) (encSize int64) {
	var w countingWriter
	mw := multipart.NewWriter(&w)
	for _, ra := range ranges {
		mw.CreatePart(ra.mimeHeader(contentType, contentSize))
		encSize += ra.length
	}
	mw.Close()
	encSize += int64(w)
	return
}

// countingWriter counts how many bytes have been written to it.
type countingWriter int64

func (w *countingWriter) Write(p []byte) (n int, err error) {
	*w += countingWriter(len(p))
	return len(p), nil
}
