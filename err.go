package ops

import "errors"

var (
	ErrNilType = errors.New("type cannot be nil")
	ErrNilTag  = errors.New("tag cannot be nil")
)