package httpkit

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetETag(t *testing.T) {
	const jsonStr = `
{"Content-Length":["51"],"Content-Type":["application/json; charset=utf-8"],"Date":["Tue, 28 Feb 2023 04:24:13 GMT"],"Etag":["\"4aa2596a8f744c1a660d82c87c0a40c4\""],"X-Trace-Id":["3NZ4qMbYt37e6Gubow0y2J"]}
`
	var header http.Header
	err := json.Unmarshal([]byte(jsonStr), &header)
	if err != nil {
		t.Fatal(err)
	}

	etag := GetETag(header)
	if etag == "" {
		t.FailNow()
	}
}
