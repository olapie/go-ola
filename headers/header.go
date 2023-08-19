package headers

import (
	"encoding/base64"
	"fmt"
	"go.olapie.com/ola/internal/types"
	"google.golang.org/grpc/metadata"
	"mime"
	"net/http"
	"reflect"
	"strings"
)

// http.Client will convert x-app-id to X-App-Id by default

const (
	KeyAuthorization       = "Authorization"
	KeyAcceptEncoding      = "Accept-Encoding"
	KeyACLAllowCredentials = "Access-Control-Allow-Credentials"
	KeyACLAllowHeaders     = "Access-Control-Allow-Headers"
	KeyACLAllowMethods     = "Access-Control-Allow-Methods"
	KeyACLAllowOrigin      = "Access-Control-Allow-Origin"
	KeyACLExposeHeaders    = "Access-Control-Expose-Headers"
	KeyContentType         = "Content-Type"
	KeyContentDisposition  = "Content-Disposition"
	KeyContentEncoding     = "Content-Encoding"
	KeyCookies             = "Cookies"
	KeyLocation            = "Location"
	KeyReferrer            = "Referer"
	KeyReferrerPolicy      = "Referrer-Policy"
	KeyUserAgent           = "User-Agent"
	KeyWWWAuthenticate     = "WWW-Authenticate"
	KeyAcceptLanguage      = "Accept-Language"
	KeyETag                = "ETag"

	KeyClientID  = "X-Client-Id"
	KeyAppID     = "X-App-Id"
	KeyTraceID   = "X-Trace-Id"
	KeyAPIKey    = "X-Api-Key"
	KeyServiceID = "X-Service-Id"
)

const (
	Bearer = "Bearer"
	Basic  = "Basic"
)

type HeaderTypes interface {
	~map[string][]string | ~map[string]string
}

func Get[H HeaderTypes](h H, key string) string {
	v := get(h, key)
	if v == "" {
		v = get(h, strings.ToLower(key))
	}
	return v
}

func get[H HeaderTypes](h H, key string) string {
	switch m := any(h).(type) {
	case map[string]string:
		return m[key]
	case map[string][]string:
		return http.Header(m).Get(key)
	case metadata.MD:
		if a := m.Get(key); len(a) != 0 {
			return a[0]
		}
		return ""
	case http.Header:
		return m.Get(key)
	default:
		v := reflect.ValueOf(h)
		if v.CanConvert(types.MapStringToStringType) {
			return v.Convert(types.MapStringToStringType).Interface().(map[string]string)[key]
		} else if v.CanConvert(types.MapStringToStringSliceType) {
			return http.Header(v.Convert(types.MapStringToStringSliceType).Interface().(map[string][]string)).Get(key)
		}
		panic(fmt.Sprintf("unsupported type %T", h))
	}
}

func Set[H HeaderTypes](h H, key, value string) {
	switch m := any(h).(type) {
	case map[string]string:
		m[key] = value
	case map[string][]string:
		hh := http.Header(m)
		hh.Set(key, value)
	case metadata.MD:
		m.Set(key, value)
	case http.Header:
		m.Set(key, value)
	default:
		v := reflect.ValueOf(h)
		if v.CanConvert(types.MapStringToStringType) {
			v.Convert(types.MapStringToStringType).Interface().(map[string]string)[key] = value
		} else if v.CanConvert(types.MapStringToStringSliceType) {
			http.Header(v.Convert(types.MapStringToStringSliceType).Interface().(map[string][]string)).Set(key, value)
		}
		panic(fmt.Sprintf("unsupported type %T", h))
	}
}

func SetNX[H HeaderTypes](h H, key, value string) {
	if Get(h, key) != "" {
		return
	}
	Set(h, key, value)
}

func GetAcceptEncodings[H HeaderTypes](h H) []string {
	a := strings.Split(Get(h, KeyAcceptEncoding), ",")
	for i, s := range a {
		a[i] = strings.TrimSpace(s)
	}

	// Remove empty strings
	for i := len(a) - 1; i >= 0; i-- {
		if a[i] == "" {
			a = append(a[:i], a[i+1:]...)
		}
	}
	return a
}

func GetContentType[H HeaderTypes](h H) string {
	t, _, _ := mime.ParseMediaType(Get(h, KeyContentType))
	return t
}

func SetContentType[H HeaderTypes](h H, contentType string) {
	Set(h, KeyContentType, contentType)
}

func SetContentTypeNX[H HeaderTypes](h H, contentType string) {
	SetNX(h, KeyContentType, contentType)
}

func GetAuthorization[H HeaderTypes](h H) string {
	return Get(h, KeyAuthorization)
}

func SetAuthorization[H HeaderTypes](h H, contentType string) {
	Set(h, KeyContentType, contentType)
}

func SetAuthorizationNX[H HeaderTypes](h H, contentType string) {
	SetNX(h, KeyContentType, contentType)
}

func GetBasicAccount[H HeaderTypes](h H) (user string, password string) {
	s := GetAuthorization(h)
	l := strings.Split(s, " ")
	if len(l) != 2 {
		return
	}

	if l[0] != Basic {
		return
	}

	b, err := base64.StdEncoding.DecodeString(l[1])
	if err != nil {
		return
	}

	userAndPass := strings.Split(string(b), ":")
	if len(userAndPass) != 2 {
		return
	}
	return userAndPass[0], userAndPass[1]
}

// GetBearer returns bearer token in header
func GetBearer[H HeaderTypes](h H) string {
	s := GetAuthorization(h)
	l := strings.Split(s, " ")
	if len(l) != 2 {
		return ""
	}
	if l[0] == Bearer {
		return l[1]
	}
	return ""
}

func SetBearer[H HeaderTypes](h H, bearer string) {
	authorization := Bearer + " " + bearer
	Set(h, KeyAuthorization, authorization)
}

func GetContentEncoding[H HeaderTypes](h H, encoding string) string {
	return Get(h, KeyContentEncoding)
}

func SetContentEncoding[H HeaderTypes](h H, encoding string) {
	Set(h, KeyContentEncoding, encoding)
}

func GetTraceID[H HeaderTypes](h H) string {
	return Get(h, KeyTraceID)
}

func SetTraceID[H HeaderTypes](h H, id string) {
	Set(h, KeyTraceID, id)
}

func GetClientID[H HeaderTypes](h H) string {
	return Get(h, KeyClientID)
}

func SetClientID[H HeaderTypes](h H, id string) {
	Set(h, KeyClientID, id)
}

func GetAppID[H HeaderTypes](h H) string {
	return Get(h, KeyAppID)
}

func SetAppID[H HeaderTypes](h H, id string) {
	Set(h, KeyAppID, id)
}

/**
ETag is enclosed in quotes https://www.rfc-editor.org/rfc/rfc7232#section-2.3
   Examples:

     ETag: "xyzzy"
     ETag: W/"xyzzy"
     ETag: ""
*/

func GetETag[H HeaderTypes](h H) string {
	etag := Get(h, KeyETag)
	if etag == "" {
		etag = Get(h, "Etag")
	}
	return etag
}

func SetETag[H HeaderTypes](h H, etag string) {
	Set(h, KeyETag, etag)
}

func IsWebsocket(h http.Header) bool {
	conn := strings.ToLower(h.Get("Connection"))
	if conn != "upgrade" {
		return false
	}
	return strings.EqualFold(h.Get("Upgrade"), "websocket")
}

// ToAttachment returns value for Content-Disposition
// e.g. Content-Disposition: attachment; filename=test.txt
func ToAttachment(filename string) string {
	return fmt.Sprintf(`attachment; filename="%s"`, filename)
}

func CreateUserAuthorizations(userToPassword map[string]string) map[string]string {
	userToAuthorization := make(map[string]string)
	for user, password := range userToPassword {
		if user == "" || password == "" {
			panic("empty user or password")
		}
		account := user + ":" + password
		userToAuthorization[user] = "Basic " + base64.StdEncoding.EncodeToString([]byte(account))
	}
	return userToAuthorization
}
