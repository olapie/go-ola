package ids

type UserID interface {
	Int() (int64, bool)
	String() (string, bool)
	Value() any
}

type UserIDTypes interface {
	~int64 | ~string
}

func NewUserID[T UserIDTypes](id T) UserID {
	return &userIDImpl[T]{id: id}
}

type userIDImpl[T UserIDTypes] struct {
	id T
}

func (u *userIDImpl[T]) Value() any {
	return u.id
}

func (u *userIDImpl[T]) Int() (int64, bool) {
	i, ok := any(u.id).(int64)
	return i, ok
}

func (u *userIDImpl[T]) String() (string, bool) {
	s, ok := any(u.id).(string)
	return s, ok
}
