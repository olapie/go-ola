package types

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorString string

func (s ErrorString) Error() string {
	return string(s)
}

type Error struct {
	code    int
	subCode int
	message string
}

type jsonError struct {
	Code    int    `json:"code,omitempty"`
	SubCode int    `json:"sub_code,omitempty"`
	Message string `json:"message,omitempty"`
}

var _ json.Marshaler = (*Error)(nil)
var _ json.Unmarshaler = (*Error)(nil)

func (e *Error) Code() int {
	return e.code
}

func (e *Error) SubCode() int {
	return e.subCode
}

func (e *Error) Message() string {
	return e.message
}

func (e *Error) String() string {
	return e.Error()
}

func (e *Error) Error() string {
	if e.message == "" {
		e.message = http.StatusText(e.code)
		if e.message == "" {
			e.message = fmt.Sprint(e.code)
		} else if e.subCode > 0 {
			e.message = fmt.Sprintf("%s (%d)", e.message, e.subCode)
		}
	}
	return e.message
}

func (e *Error) Is(target error) bool {
	if e == target {
		return true
	}

	if t, ok := target.(*Error); ok {
		return t.code == e.code && t.subCode == e.subCode && t.message == e.message
	}
	return false
}

func (e *Error) MarshalJSON() (text []byte, err error) {
	je := &jsonError{
		Code:    e.code,
		SubCode: e.subCode,
		Message: e.message,
	}
	return json.Marshal(je)
}

func (e *Error) UnmarshalJSON(text []byte) error {
	var je jsonError
	err := json.Unmarshal(text, &je)
	if err != nil {
		return err
	}
	e.code = je.Code
	e.subCode = je.SubCode
	e.message = je.Message
	return nil
}

func NewError(code int, format string, a ...any) *Error {
	msg := fmt.Sprintf(format, a...)
	if msg == "" {
		msg = http.StatusText(code)
	}
	return &Error{
		code:    code,
		message: msg,
	}
}

func NewSubError(code, subCode int, message string) *Error {
	if code <= 0 {
		panic("invalid code")
	}

	if subCode <= 0 {
		panic("invalid subCode")
	}

	if message == "" {
		message = http.StatusText(code)
	}
	return &Error{
		code:    code,
		subCode: subCode,
		message: message,
	}
}

func BadRequest(format string, a ...any) error {
	return NewError(http.StatusBadRequest, format, a...)
}

func Unauthorized(format string, a ...any) error {
	return NewError(http.StatusUnauthorized, format, a...)
}

func PaymentRequired(format string, a ...any) error {
	return NewError(http.StatusPaymentRequired, format, a...)
}

func Forbidden(format string, a ...any) error {
	return NewError(http.StatusForbidden, format, a...)
}

func NotFound(format string, a ...any) error {
	return NewError(http.StatusNotFound, format, a...)
}

func MethodNotAllowed(format string, a ...any) error {
	return NewError(http.StatusMethodNotAllowed, format, a...)
}

func NotAcceptable(format string, a ...any) error {
	return NewError(http.StatusNotAcceptable, format, a...)
}

func ProxyAuthRequired(format string, a ...any) error {
	return NewError(http.StatusProxyAuthRequired, format, a...)
}

func RequestTimeout(format string, a ...any) error {
	return NewError(http.StatusRequestTimeout, format, a...)
}

func Conflict(format string, a ...any) error {
	return NewError(http.StatusConflict, format, a...)
}

func LengthRequired(format string, a ...any) error {
	return NewError(http.StatusLengthRequired, format, a...)
}

func PreconditionFailed(format string, a ...any) error {
	return NewError(http.StatusPreconditionFailed, format, a...)
}

func RequestEntityTooLarge(format string, a ...any) error {
	return NewError(http.StatusRequestEntityTooLarge, format, a...)
}

func RequestURITooLong(format string, a ...any) error {
	return NewError(http.StatusRequestURITooLong, format, a...)
}

func ExpectationFailed(format string, a ...any) error {
	return NewError(http.StatusExpectationFailed, format, a...)
}

func Teapot(format string, a ...any) error {
	return NewError(http.StatusTeapot, format, a...)
}

func MisdirectedRequest(format string, a ...any) error {
	return NewError(http.StatusMisdirectedRequest, format, a...)
}

func UnprocessableEntity(format string, a ...any) error {
	return NewError(http.StatusUnprocessableEntity, format, a...)
}

func Locked(format string, a ...any) error {
	return NewError(http.StatusLocked, format, a...)
}

func TooEarly(format string, a ...any) error {
	return NewError(http.StatusTooEarly, format, a...)
}

func UpgradeRequired(format string, a ...any) error {
	return NewError(http.StatusUpgradeRequired, format, a...)
}

func PreconditionRequired(format string, a ...any) error {
	return NewError(http.StatusPreconditionRequired, format, a...)
}

func TooManyRequests(format string, a ...any) error {
	return NewError(http.StatusTooManyRequests, format, a...)
}

func RequestHeaderFieldsTooLarge(format string, a ...any) error {
	return NewError(http.StatusRequestHeaderFieldsTooLarge, format, a...)
}

func UnavailableForLegalReasons(format string, a ...any) error {
	return NewError(http.StatusUnavailableForLegalReasons, format, a...)
}

func InternalServerError(format string, a ...any) error {
	return NewError(http.StatusInternalServerError, format, a...)
}

func NotImplemented(format string, a ...any) error {
	return NewError(http.StatusNotImplemented, format, a...)
}

func BadGateway(format string, a ...any) error {
	return NewError(http.StatusBadGateway, format, a...)
}

func ServiceUnavailable(format string, a ...any) error {
	return NewError(http.StatusServiceUnavailable, format, a...)
}

func GatewayTimeout(format string, a ...any) error {
	return NewError(http.StatusGatewayTimeout, format, a...)
}

func HTTPVersionNotSupported(format string, a ...any) error {
	return NewError(http.StatusHTTPVersionNotSupported, format, a...)
}

func VariantAlsoNegotiates(format string, a ...any) error {
	return NewError(http.StatusVariantAlsoNegotiates, format, a...)
}

func InsufficientStorage(format string, a ...any) error {
	return NewError(http.StatusInsufficientStorage, format, a...)
}

func LoopDetected(format string, a ...any) error {
	return NewError(http.StatusLoopDetected, format, a...)
}

func NotExtended(format string, a ...any) error {
	return NewError(http.StatusNotExtended, format, a...)
}

func NetworkAuthenticationRequired(format string, a ...any) error {
	return NewError(http.StatusNetworkAuthenticationRequired, format, a...)
}
