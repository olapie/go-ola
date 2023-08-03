package types

type UserIDTypes interface {
	~int64 | ~string
}

type UserID[T UserIDTypes] struct {
	id T
}

func (u *UserID[T]) Value() any {
	return u.id
}

func (u *UserID[T]) Int() (int64, bool) {
	i, ok := any(u.id).(int64)
	return i, ok
}

func (u *UserID[T]) String() (string, bool) {
	s, ok := any(u.id).(string)
	return s, ok
}

func NewUserID[T ~int64 | ~string](id T) *UserID[T] {
	return &UserID[T]{id: id}
}
