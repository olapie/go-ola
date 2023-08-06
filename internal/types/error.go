package types

import (
	"fmt"
	"net/http"
)

type ErrorString string

func (s ErrorString) Error() string {
	return string(s)
}

type Error struct {
	Code    int    `json:"code,omitempty"`
	SubCode int    `json:"sub_code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e *Error) String() string {
	return e.Error()
}

func (e *Error) Error() string {
	if e.Message == "" {
		e.Message = http.StatusText(e.Code)
		if e.Message == "" {
			e.Message = fmt.Sprint(e.Code)
		} else if e.SubCode > 0 {
			e.Message = fmt.Sprintf("%s (%d)", e.Message, e.SubCode)
		}
	}
	return e.Message
}

func (e *Error) Is(target error) bool {
	if e == target {
		return true
	}

	if t, ok := target.(*Error); ok {
		return t.Code == e.Code && t.SubCode == e.SubCode && t.Message == e.Message
	}
	return false
}
