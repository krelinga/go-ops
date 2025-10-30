package ops

import "errors"

var (
	ErrNilType   = errors.New("type cannot be nil")
	ErrNilTag    = errors.New("tag cannot be nil")
	ErrNilField  = errors.New("field cannot be nil")
	ErrWrongType = errors.New("value has wrong type")
	ErrInternal  = errors.New("internal error")
)
