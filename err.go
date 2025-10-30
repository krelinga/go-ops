package ops

import "errors"

var (
	ErrNilType   = errors.New("type cannot be nil")
	ErrNilTag    = errors.New("tag cannot be nil")
	ErrNilField  = errors.New("field cannot be nil")
	ErrWrongType = errors.New("value has wrong type")
	ErrInternal  = errors.New("internal error")
)

func try(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if recErr, ok := r.(error); ok {
				if errors.Is(recErr, ErrNilType) ||
					errors.Is(recErr, ErrNilTag) ||
					errors.Is(recErr, ErrNilField) ||
					errors.Is(recErr, ErrWrongType) ||
					errors.Is(recErr, ErrInternal) {
					err = recErr
					return
				}
			}
			panic(r)
		}
	}()
	f()
	return
}
