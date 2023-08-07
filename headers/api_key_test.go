package headers

import (
	"net/http"
	"testing"

	"go.olapie.com/security/base62"
	"go.olapie.com/utils"
)

func TestSetAPIKey(t *testing.T) {
	header := http.Header{}
	header.Set(KeyTraceID, base62.NewUUIDString())
	header.Set(KeyAppID, "test")
	header.Set(KeyClientID, base62.NewUUIDString())

	SetAPIKey(header)
	t.Log(header)

	res := VerifyAPIKey(header, 10)
	utils.MustTrueT(t, res)
}
