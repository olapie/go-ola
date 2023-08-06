package grpcutil

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"go.olapie.com/security/base62"
	"google.golang.org/grpc/metadata"
)

func TestSign(t *testing.T) {
	md := make(metadata.MD)
	SetAPIKey(md)
	t.Log(md)
	if !VerifyAPIKey(md, 1) {
		t.FailNow()
	}

	ts := time.Now().Unix() - 20
	var b [36]byte
	binary.BigEndian.PutUint64(b[:], uint64(ts))
	hash := sha256.New().Sum([]byte(fmt.Sprintf("%x", ts)))
	copy(b[4:], hash[:])
	sign := base62.EncodeToString(b[:])
	md.Set(KeyAPIKey, sign)
	if VerifyAPIKey(md, 1) {
		t.FailNow()
	}
}
