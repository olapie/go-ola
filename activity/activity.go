package activity

import (
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"

	"go.olapie.com/ola/headers"
	internalTypes "go.olapie.com/ola/internal/types"
	"go.olapie.com/ola/session"
	"go.olapie.com/ola/types"
	"google.golang.org/grpc/metadata"
)

const (
	ErrNotExist internalTypes.ErrorString = "Activity does not exist"
)

type HeaderTypes interface {
	http.Header | metadata.MD | map[string]string
}

type Activity struct {
	name string
	// http request
	header http.Header

	// grpc request
	md metadata.MD

	// aws lambda request
	properties map[string]string

	//Session is only available in incoming context, may be nil if session is not enabled
	session   *session.Session
	userID    types.UserID
	authAppID string
}

func New[H HeaderTypes](name string, header H) *Activity {
	a := &Activity{
		name: name,
	}

	if header == nil {
		panic("header is nil")
	}

	switch v := any(header).(type) {
	case http.Header:
		a.header = v
	case metadata.MD:
		a.md = v
	case map[string]string:
		a.properties = v
	default:
		panic(fmt.Sprintf("unsupported header type: %T", header))
	}
	return a
}

func (a *Activity) Name() string {
	return a.name
}

func (a *Activity) Session() *session.Session {
	return a.session
}

func (a *Activity) UserID() types.UserID {
	return a.userID
}

func (a *Activity) SetUserID(id types.UserID) {
	a.userID = id
}

func (a *Activity) Set(key string, value string) {
	if a.header != nil {
		a.header.Set(key, value)
	} else if a.md != nil {
		a.md.Set(key, value)
	} else {
		a.properties[key] = value
	}
}

func (a *Activity) Get(key string) string {
	if a.header != nil {
		return a.header.Get(key)
	}

	if a.md != nil {
		l := a.md.Get(key)
		if len(l) != 0 {
			return l[0]
		}
		return ""
	}

	if v, ok := a.properties[key]; ok {
		return v
	}

	if v, ok := a.properties[strings.ToLower(key)]; ok {
		return v
	}

	if v, ok := a.properties[textproto.CanonicalMIMEHeaderKey(key)]; ok {
		return v
	}

	return ""
}

func (a *Activity) GetAuthAppID() string {
	return a.authAppID
}

func (a *Activity) SetAuthAppID(id string) {
	a.authAppID = id
}

func (a *Activity) GetAppID() string {
	return a.Get(headers.KeyAppID)
}

func (a *Activity) SetAppID(id string) {
	a.Set(headers.KeyAppID, id)
}

func (a *Activity) GetTraceID() string {
	return a.Get(headers.KeyTraceID)
}

func (a *Activity) SetTraceID(id string) {
	a.Set(headers.KeyTraceID, id)
}

func (a *Activity) GetClientID() string {
	return a.Get(headers.KeyClientID)
}

func (a *Activity) SetClientID(id string) {
	a.Set(headers.KeyClientID, id)
}

func (a *Activity) GetAuthorization() string {
	return a.Get(headers.KeyAuthorization)
}

func (a *Activity) SetAuthorization(auth string) {
	a.Set(headers.KeyAuthorization, auth)
}

func (a *Activity) GetRequestTimeout() int {
	s := a.Get(headers.KeyRequestTimeout)
	if s == "" {
		return 0
	}
	t, err := strconv.Atoi(s)
	if err != nil || t < 0 {
		slog.Error("invalid Request-Timeout: " + s)
		return 0
	}
	return t
}

func (a *Activity) SetRequestTimeout(seconds int) {
	if seconds > 0 {
		a.Set(headers.KeyRequestTimeout, strconv.Itoa(seconds))
	}
}

func CopyHeader[H HeaderTypes](dest H, a *Activity) {
	switch h := any(dest).(type) {
	case http.Header:
		if a.header != nil {
			maps.Copy(h, a.header)
		} else if a.md != nil {
			for k, v := range a.md {
				h[textproto.CanonicalMIMEHeaderKey(k)] = v
			}
		} else {
			for k, v := range a.properties {
				h.Set(k, v)
			}
		}
	case metadata.MD:
		if a.header != nil {
			for k, v := range a.header {
				h[strings.ToLower(k)] = v
			}
		} else if a.md != nil {
			maps.Copy(h, a.md)
		} else {
			for k, v := range a.properties {
				h.Set(k, v)
			}
		}
	case map[string]string:
		if a.header != nil {
			for k, v := range a.header {
				if len(v) != 0 {
					h[k] = v[0]
				}
			}
		} else if a.md != nil {
			for k, v := range a.md {
				if len(v) != 0 {
					h[k] = v[0]
				}
			}
		} else {
			maps.Copy(h, a.properties)
		}
	}
}
