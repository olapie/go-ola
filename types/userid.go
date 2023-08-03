package types

import "go.olapie.com/ola/internal/types"

type UserID interface {
	Int() (int64, bool)
	String() (string, bool)
	Value() any
}

type UserIDTypes = types.UserIDTypes

func NewUserID[T ~int64 | ~string](id T) UserID {
	return types.NewUserID(id)
}
