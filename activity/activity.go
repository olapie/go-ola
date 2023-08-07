package activity

import (
	"net/http"
	"strings"

	internalTypes "go.olapie.com/ola/internal/types"
	"go.olapie.com/ola/session"
	"go.olapie.com/ola/types"
	"google.golang.org/grpc/metadata"
)

const (
	ErrNotExist internalTypes.ErrorString = "activity does not exist"
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

func New[T http.Header | metadata.MD](name string, header T) Activity {
	a := &activity{
		name:   name,
		header: header,
	}
	if a.header == nil {
		a.header = make(map[string][]string)
	}
	return a
}

type activity struct {
	name string
	// header can be http.Header or grpc metadata.MD
	header map[string][]string

	//Session is only available in incoming context, may be nil if session is not enabled
	session *session.Session
	userID  types.UserID
}

func (a *activity) Name() string {
	return a.name
}

func (a *activity) Session() *session.Session {
	return a.session
}

func (a *activity) UserID() types.UserID {
	return a.userID
}

func (a *activity) SetUserID(id types.UserID) {
	a.userID = id
}

func (a *activity) Set(key string, value string) {
	a.header[key] = []string{value}
}

func (a *activity) Get(key string) string {
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

func (a *activity) Header() map[string][]string {
	return a.header
}
