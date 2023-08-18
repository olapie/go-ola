package activity

import (
	"net/http"
	"reflect"
	"strings"
	"sync"

	"go.olapie.com/ola/headers"
	internalTypes "go.olapie.com/ola/internal/types"
	"go.olapie.com/ola/session"
	"go.olapie.com/ola/types"
)

const (
	ErrNotExist internalTypes.ErrorString = "Activity does not exist"
)

type Activity struct {
	name string
	// header can be http.Header or grpc metadata.MD
	header         map[string][]string
	initHeaderOnce sync.Once

	//Session is only available in incoming context, may be nil if session is not enabled
	session *session.Session
	userID  types.UserID
}

func New[T ~map[string][]string | ~map[string]string](name string, header T) *Activity {
	a := &Activity{
		name: name,
	}

	if header != nil {
		headerVal := reflect.ValueOf(header)
		if headerVal.Type().Elem().Kind() != reflect.String {
			a.header = headerVal.Convert(internalTypes.MapStringToStringSliceType).Interface().(map[string][]string)
		} else {
			m := headerVal.Convert(reflect.TypeOf(internalTypes.MapStringToStringType)).Interface().(map[string]string)
			a.header = make(map[string][]string, len(m))
			for k, v := range m {
				a.header[k] = []string{v}
			}
		}
	} else {
		a.header = make(map[string][]string)
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
	if a.header == nil {
		a.initHeaderOnce.Do(func() {
			a.header = make(map[string][]string)
		})
	}
	a.header[key] = []string{value}
}

func (a *Activity) Get(key string) string {
	if a == nil || a.header == nil {
		return ""
	}

	if l := a.header[key]; len(l) != 0 {
		return l[0]
	}

	if l := a.header[strings.ToLower(key)]; len(l) != 0 {
		return l[0]
	}

	if v := http.Header(a.header).Get(key); v != "" {
		return v
	}
	return ""
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

// Header returns values for http.Header or metadata.MD
// As http.Header and metadata.MD format header key in different ways, please copy values by http.Header.Set or metadata.MD.Set
func (a *Activity) Header() map[string][]string {
	if a.header == nil {
		a.initHeaderOnce.Do(func() {
			a.header = make(map[string][]string)
		})
	}
	return a.header
}
