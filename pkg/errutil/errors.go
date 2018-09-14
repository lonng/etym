package errutil

import "errors"

var (
	ErrEmptyQueryWord     = errors.New("empty query word")
	ErrWordNotFound       = errors.New("word not found")
	ErrInvalidImprovement = errors.New("invalid improvement")
	ErrDatabaseError      = errors.New("database error")
	ErrServerInternal     = errors.New("server internal error")
	ErrWordCantEmpty      = errors.New("word can not empty")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrIllegalParameter   = errors.New("illegal parameter")
	ErrWrongPassword      = errors.New("wrong password")
	ErrUserNotFound       = errors.New("uesr not found")
	ErrUserExists         = errors.New("uesr exists")
)
