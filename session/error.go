package session

type errorString string

func (e errorString) Error() string {
	return string(e)
}

const (
	ErrNotExist         errorString = "session does not exist"
	ErrNoValue          errorString = "no value"
	ErrTooManyConflicts errorString = "too many conflict"
)
