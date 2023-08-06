package grpcutil

import (
	"strings"

	"google.golang.org/grpc/metadata"
)

const (
	KeyClientID      = "x-client-id"
	KeyAppID         = "x-app-id"
	KeyTraceID       = "x-trace-id"
	KeyAPIKey        = "x-api-key"
	keyAuthorization = "authorization"
)

func MatchMetadata(key string) (string, bool) {
	key = strings.ToLower(key)
	switch key {
	case KeyClientID, KeyAppID, KeyTraceID, KeyAPIKey:
		return key, true
	default:
		return "", false
	}
}

func GetTraceID(md metadata.MD) string {
	return GetMetadata(md, KeyTraceID)
}

func SetTraceID(md metadata.MD, id string) {
	md.Set(KeyTraceID, id)
}

func GetClientID(md metadata.MD) string {
	return GetMetadata(md, KeyClientID)
}

func SetClientID(md metadata.MD, id string) {
	md.Set(KeyClientID, id)
}

func GetAppID(md metadata.MD) string {
	return GetMetadata(md, KeyAppID)
}

func SetAppID(md metadata.MD, id string) {
	md.Set(KeyAppID, id)
}

func GetAuthorization(md metadata.MD) string {
	return GetMetadata(md, keyAuthorization)
}

func SetAuthorization(md metadata.MD, a string) {
	md.Set(keyAuthorization, a)
}

func GetMetadata(m metadata.MD, key string) string {
	v, ok := m[key]
	if !ok {
		v = m[strings.ToLower(key)]
	}
	if len(v) == 0 {
		return ""
	}
	return v[0]
}
