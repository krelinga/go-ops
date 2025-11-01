package ops

import "reflect"

type Ord interface {
	Ord(Env, reflect.Value, reflect.Value) int
}

type ordTag struct {}

func Order[T any](env Env, a, b T) int {
	return OrderVals(env, reflect.ValueOf(a), reflect.ValueOf(b))
}

func OrderVals(env Env, a, b reflect.Value) int {
	if !a.IsValid() || !b.IsValid() {
		panic(ErrInvalid)
	}
	if a.Type() != b.Type() {
		panic(ErrWrongType)
	}
	t := a.Type()
	impl := func() Ord {
		if env == nil {
			return ordDefault{}
		}
		anyVal, ok := env.Get(t, ordTag{})
		if !ok {
			return ordDefault{}
		}
		impl := anyVal.(Ord)
		if impl == nil {
			return ordDefault{}
		}
		return impl
	}()
	return impl.Ord(env, a, b)
}

type ordDefault struct{}

func (o ordDefault) Ord(env Env, v1, v2 reflect.Value) int {
	return 0 // TODO
}

type OrdDeep struct{}

func (o OrdDeep) Ord(env Env, v1, v2 reflect.Value) int {
	return OrderVals(env, v1, v2)
}

func OrdOpt(t reflect.Type, ord Ord) Opt {
	return OptFunc(func(e Env) {
		e.Set(t, ordTag{}, ord)
	})
}

func OrdOptAll(ord Ord) Opt {
	return OptFunc(func(e Env) {
		e.SetAll(ordTag{}, ord)
	})
}