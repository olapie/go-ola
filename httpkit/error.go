package httpkit

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"go.olapie.com/ola/types"
)

func ReadError(resp *http.Response) error {
	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	contentType := resp.Header.Get("Content-Type")
	if !isText(contentType) {
		return types.NewError(resp.StatusCode, resp.Status)
	}

	body, ioErr := io.ReadAll(resp.Body)
	resp.Body.Close()
	if ioErr != nil {
		log.Printf("failed reading response body: %v\n", ioErr)
		return nil
	}

	if strings.HasPrefix(contentType, "application/json") {
		var respError types.Error
		if err := json.Unmarshal(body, &respError); err != nil {
			log.Printf("unmarshal json body: %v\n", err)
		} else {
			return &respError
		}
	}

	code := resp.StatusCode
	message := string(body)
	if message == "" {
		message = resp.Status
	}

	return types.NewError(code, message)
}

var textTypes = []string{
	"text/plain", "text/html", "text/xml", "text/css", "application/xml", "application/xhtml+xml",
}

func isText(mimeType string) bool {
	for _, t := range textTypes {
		if strings.HasPrefix(mimeType, t) {
			return true
		}
	}
	return false
}
