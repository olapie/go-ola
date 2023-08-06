package headers

import (
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
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

const (
	MimePlain      = "text/plain"
	MimeHTML       = "text/html"
	MimeXML2       = "text/xml"
	MimeCSS        = "text/css"
	MimeJavascript = "text/javascript" // application/javascript is obsolete

	MimeXML      = "application/xml"
	MimeXHTML    = "application/xhtml+xml"
	MimeProtobuf = "application/x-protobuf"

	MimeFormData = "multipart/form-data"
	MimeGIF      = "image/gif"
	MimeJPEG     = "image/jpeg"
	MimePNG      = "image/png"
	MimeWEBP     = "image/webp"
	MimeICON     = "image/x-icon"

	MimeMPEG = "video/mpeg"

	FormURLEncoded  = "application/x-www-form-urlencoded"
	MimeOctetStream = "application/octet-stream"
	MimeJSON        = "application/json"
	MimePDF         = "application/pdf"
	MimeMSWord      = "application/msword"
	MimeGZIP        = "application/x-gzip"
	MimeWASM        = "application/wasm"
)

const (
	MimeCharsetUTF8 = "charset=utf-8"

	charsetSuffix = "; " + MimeCharsetUTF8

	MimePlainUTF8 = MimePlain + charsetSuffix

	// MimeHtmlUTF8 is better than HTMLUTF8, etc.
	MimeHtmlUTF8 = MimeHTML + charsetSuffix
	MimeJsonUTF8 = MimeJSON + charsetSuffix
	MimeXmlUTF8  = MimeXML + charsetSuffix
)

func Get[H ~map[string][]string](h H, key string) string {
	if l := h[key]; len(l) != 0 {
		return l[0]
	}

	if l := h[strings.ToLower(key)]; len(l) != 0 {
		return l[0]
	}

	if v := http.Header(h).Get(key); v != "" {
		return v
	}
	return ""
}

func Set[H ~map[string][]string](h H, key, value string) {
	switch m := any(h).(type) {
	case map[string]string:
		m[key] = value
	case map[string][]string:
		hh := http.Header(m)
		hh.Set(key, value)
	case http.Header:
		m.Set(key, value)
	default:
		panic(fmt.Sprintf("unsupported type %T", h))
	}
}

func SetNX[H ~map[string][]string](h H, key, value string) {
	if Get(h, key) != "" {
		return
	}
	Set(h, key, value)
}

func GetAcceptEncodings[H ~map[string][]string](h H) []string {
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

func GetContentType[H ~map[string][]string](h H) string {
	t, _, _ := mime.ParseMediaType(Get(h, KeyContentType))
	return t
}

func SetContentType[H ~map[string][]string](h H, contentType string) {
	Set(h, KeyContentType, contentType)
}

func SetContentTypeNX[H ~map[string][]string](h H, contentType string) {
	SetNX(h, KeyContentType, contentType)
}

func GetAuthorization[H ~map[string][]string](h H) string {
	return Get(h, KeyAuthorization)
}

func SetAuthorization[H ~map[string][]string](h H, contentType string) {
	Set(h, KeyContentType, contentType)
}

func SetAuthorizationNX[H ~map[string][]string](h H, contentType string) {
	SetNX(h, KeyContentType, contentType)
}

func GetBasicAccount[H ~map[string][]string](h H) (user string, password string) {
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
func GetBearer[H ~map[string][]string](h H) string {
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

func SetBearer[H ~map[string][]string](h H, bearer string) {
	authorization := Bearer + " " + bearer
	Set(h, KeyAuthorization, authorization)
}

func GetContentEncoding[H ~map[string][]string](h H, encoding string) string {
	return Get(h, KeyContentEncoding)
}

func SetContentEncoding[H ~map[string][]string](h H, encoding string) {
	Set(h, KeyContentEncoding, encoding)
}

func GetTraceID[H ~map[string][]string](h H) string {
	return Get(h, KeyTraceID)
}

func SetTraceID[H ~map[string][]string](h H, id string) {
	Set(h, KeyTraceID, id)
}

func GetClientID[H ~map[string][]string](h H) string {
	return Get(h, KeyClientID)
}

func SetClientID[H ~map[string][]string](h H, id string) {
	Set(h, KeyClientID, id)
}

func GetAppID[H ~map[string][]string](h H) string {
	return Get(h, KeyAppID)
}

func SetAppID[H ~map[string][]string](h H, id string) {
	Set(h, KeyAppID, id)
}

/**
ETag is enclosed in quotes https://www.rfc-editor.org/rfc/rfc7232#section-2.3
   Examples:

     ETag: "xyzzy"
     ETag: W/"xyzzy"
     ETag: ""
*/

func GetETag[H ~map[string][]string](h H) string {
	etag := Get(h, KeyETag)
	if etag == "" {
		etag = Get(h, "Etag")
	}
	return etag
}

func SetETag[H ~map[string][]string](h H, etag string) {
	Set(h, KeyETag, etag)
}

func IsWebsocket(h http.Header) bool {
	conn := strings.ToLower(h.Get("Connection"))
	if conn != "upgrade" {
		return false
	}
	return strings.EqualFold(h.Get("Upgrade"), "websocket")
}

func IsMimeText[T string | http.Header](typeOrHeader T) bool {
	switch v := any(typeOrHeader).(type) {
	case string:
		switch v {
		case MimePlain, MimeHTML, MimeCSS, MimeXML, MimeXML2, MimeXHTML, MimeJSON, MimePlainUTF8, MimeHtmlUTF8,
			MimeJsonUTF8, MimeXmlUTF8:
			return true
		default:
			return false
		}
	case http.Header:
		return IsMimeText(Get(v, KeyContentType))
	default:
		return false
	}
}

func IsMimeXML[T string | http.Header](typeOrHeader T) bool {
	switch v := any(typeOrHeader).(type) {
	case string:
		switch v {
		case MimeXML, MimeXML2, MimeXmlUTF8:
			return true
		default:
			return false
		}
	case http.Header:
		return IsMimeXML(Get(v, KeyContentType))
	default:
		return false
	}
}

func IsMimeJSON[T string | http.Header](typeOrHeader T) bool {
	switch v := any(typeOrHeader).(type) {
	case string:
		switch v {
		case MimeJSON, MimeJsonUTF8:
			return true
		default:
			return false
		}
	case http.Header:
		return IsMimeXML(Get(v, KeyContentType))
	default:
		return false
	}
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
