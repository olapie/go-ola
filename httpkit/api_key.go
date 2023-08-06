package httpkit

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.olapie.com/security"
	"go.olapie.com/security/base62"
)

func SetAPIKey[T http.Header | *http.Request](reqOrHeader T) {
	switch v := any(reqOrHeader).(type) {
	case http.Header:
		setAPIKey(v)
	case *http.Request:
		setAPIKey(v.Header)
	}
}

func setAPIKey(header http.Header) {
	t := time.Now().Unix()
	var b [41]byte
	b[0] = 1
	binary.BigEndian.PutUint64(b[1:], uint64(t))
	clientID := GetClientID(header)
	traceID := GetTraceID(header)
	hash := security.Hash32(fmt.Sprint(t) + traceID + clientID)
	copy(b[9:], hash[:])
	sign := base62.EncodeToString(b[:])
	header.Set(KeyAPIKey, sign)
}

func VerifyAPIKey[T http.Header | *http.Request](reqOrHeader T, delaySeconds int) bool {
	switch v := any(reqOrHeader).(type) {
	case http.Header:
		return verifyAPIKey(v, delaySeconds)
	case *http.Request:
		return verifyAPIKey(v.Header, delaySeconds)
	default:
		return false
	}
}

func verifyAPIKey(header http.Header, delaySeconds int) bool {
	sign := GetHeader(header, KeyAPIKey)
	if sign == "" {
		log.Println("missing", KeyAPIKey)
		return false
	}

	b, err := base62.DecodeString(sign)
	if err != nil {
		log.Println("invalid", KeyAPIKey, err)
		return false
	}

	t := int64(binary.BigEndian.Uint64(b[1:]))
	elapsed := time.Now().Unix() - t
	if elapsed < -3 || elapsed > int64(delaySeconds) {
		log.Println("invalid timestamp", t, elapsed)
		return false
	}
	clientID := GetClientID(header)
	traceID := GetTraceID(header)
	hash := security.Hash32(fmt.Sprint(t) + traceID + clientID)
	return bytes.Equal(b[9:], hash[:])
}
