package errorutil

import (
	"fmt"
	"net/http"
	"reflect"

	"google.golang.org/grpc/status"

	"go.olapie.com/ola/internal/types"
	"go.olapie.com/utils"
)

func NewError(code int, format string, a ...any) error {
	msg := fmt.Sprintf(format, a...)
	if msg == "" {
		msg = http.StatusText(code)
	}
	return &types.Error{
		Code:    code,
		Message: msg,
	}
}

func NewSubError(code, subCode int, message string) error {
	if code <= 0 {
		panic("invalid code")
	}

	if subCode <= 0 {
		panic("invalid subCode")
	}

	if message == "" {
		message = http.StatusText(code)
	}
	return &types.Error{
		Code:    code,
		SubCode: subCode,
		Message: message,
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

func Wrap(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	a = append(a, err)
	return fmt.Errorf(format+":%w", a...)
}

// Cause returns the root cause error
func Cause(err error) error {
	for {
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	return err
}

func GetCode(err error) int {
	if st, ok := status.FromError(err); ok {
		return int(st.Code())
	}
	err = Cause(err)
	if err == nil {
		return 0
	}

	if s, ok := err.(interface{ Code() int }); ok {
		return s.Code()
	}

	if s, ok := err.(interface{ GetCode() int }); ok {
		return s.GetCode()
	}

	// ------------
	// int32

	if s, ok := err.(interface{ Code() int32 }); ok {
		return int(s.Code())
	}

	if s, ok := err.(interface{ GetCode() int32 }); ok {
		return int(s.GetCode())
	}

	if s, ok := err.(interface{ StatusCode() int }); ok {
		return s.StatusCode()
	}

	if s, ok := err.(interface{ GetStatusCode() int }); ok {
		return s.GetStatusCode()
	}

	if s, ok := err.(interface{ Status() int }); ok {
		return s.Status()
	}

	if s, ok := err.(interface{ GetStatus() int }); ok {
		return s.GetStatus()
	}

	// ------------
	// int32

	if s, ok := err.(interface{ StatusCode() int32 }); ok {
		return int(s.StatusCode())
	}

	if s, ok := err.(interface{ GetStatusCode() int32 }); ok {
		return int(s.GetStatusCode())
	}

	if s, ok := err.(interface{ Status() int32 }); ok {
		return int(s.Status())
	}

	if s, ok := err.(interface{ GetStatus() int32 }); ok {
		return int(s.GetStatus())
	}

	v := reflect.ValueOf(utils.Indirect(err))
	t := v.Type()
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			ft := t.Field(i)
			switch ft.Name {
			case "Code", "Status", "StatusCode", "ErrorCode":
				fv := v.Field(i)
				if fv.CanInt() {
					return int(fv.Int())
				}

				if fv.CanUint() {
					return int(fv.Uint())
				}

				return 0
			}
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			if k.Kind() != reflect.String {
				continue
			}
			switch k.String() {
			case "Code", "Status", "StatusCode", "ErrorCode":
				vv := v.MapIndex(k)
				if vv.CanInt() {
					return int(vv.Int())
				}

				if vv.CanUint() {
					return int(vv.Uint())
				}

				return 0
			}
		}
	}
	return 0
}
