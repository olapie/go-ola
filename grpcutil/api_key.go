package grpcutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"time"

	"go.olapie.com/security"
	"go.olapie.com/security/base62"
	"google.golang.org/grpc/metadata"
)

func SetAPIKey(md metadata.MD) {
	t := time.Now().Unix()
	var b [41]byte
	b[0] = 1
	binary.BigEndian.PutUint64(b[1:], uint64(t))
	clientID := GetClientID(md)
	traceID := GetTraceID(md)
	hash := security.Hash32(fmt.Sprint(t) + traceID + clientID)
	copy(b[9:], hash[:])
	sign := base62.EncodeToString(b[:])
	md.Set(KeyAPIKey, sign)
}

func VerifyAPIKey(md metadata.MD, delaySeconds int) bool {
	sign := GetMetadata(md, KeyAPIKey)
	if sign == "" {
		slog.Warn("missing " + KeyAPIKey)
		return false
	}

	b, err := base62.DecodeString(sign)
	if err != nil {
		slog.Error("invalid "+KeyAPIKey, "err", err.Error())
		return false
	}
	t := int64(binary.BigEndian.Uint64(b[1:]))
	elapsed := time.Now().Unix() - t
	if elapsed < -3 || elapsed > int64(delaySeconds) {
		slog.Error("invalid timestamp", "value", t)
		return false
	}
	clientID := GetClientID(md)
	traceID := GetTraceID(md)
	hash := security.Hash32(fmt.Sprint(t) + traceID + clientID)
	return bytes.Equal(b[9:], hash[:])
}
