package activity

type errorString string

func (e errorString) Error() string {
	return string(e)
}

const (
	ErrNotExist errorString = "activity does not exist"
)
