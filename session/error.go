package session

type errorString string

func (e errorString) Error() string {
	return string(e)
}

const (
	ErrNoValue          errorString = "no value"
	ErrTooManyConflicts errorString = "too many conflicts"
)
