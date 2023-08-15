package activity

import (
	"net/http"
	"reflect"
	"strings"

	internalTypes "go.olapie.com/ola/internal/types"
	"go.olapie.com/ola/session"
	"go.olapie.com/ola/types"
)

const (
	ErrNotExist internalTypes.ErrorString = "activityImpl does not exist"
)

type Activity interface {
	Name() string
	Session() *session.Session
	UserID() types.UserID
	SetUserID(id types.UserID)
	Set(key string, value string)
	Get(key string) string

	// Header returns values for http.Header or metadata.MD
	// As http.Header and metadata.MD format header key in different ways, please copy values by http.Header.Set or metadata.MD.Set
	Header() map[string][]string
}

func New[T ~map[string][]string | ~map[string]string](name string, header T) Activity {
	a := &activityImpl{
		name: name,
	}

	if header != nil {
		headerVal := reflect.ValueOf(header)
		if headerVal.Type().Elem().Kind() == reflect.String {
			a.header = headerVal.Convert(reflect.TypeOf(map[string][]string(nil))).Interface().(map[string][]string)
		} else {
			m := headerVal.Convert(reflect.TypeOf(map[string]string(nil))).Interface().(map[string]string)
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

type activityImpl struct {
	name string
	// header can be http.Header or grpc metadata.MD
	header map[string][]string

	//Session is only available in incoming context, may be nil if session is not enabled
	session *session.Session
	userID  types.UserID
}

func (a *activityImpl) Name() string {
	return a.name
}

func (a *activityImpl) Session() *session.Session {
	return a.session
}

func (a *activityImpl) UserID() types.UserID {
	return a.userID
}

func (a *activityImpl) SetUserID(id types.UserID) {
	a.userID = id
}

func (a *activityImpl) Set(key string, value string) {
	a.header[key] = []string{value}
}

func (a *activityImpl) Get(key string) string {
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

func (a *activityImpl) Header() map[string][]string {
	return a.header
}
