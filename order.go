package ops

import (
	"cmp"
	"reflect"
)

type Ord interface {
	Ord(Env, reflect.Value, reflect.Value) int
}

type ordTag struct{}

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

func orderLiteralCan[T cmp.Ordered](v1, v2 reflect.Value, can func(reflect.Value) bool, f func(reflect.Value) T) int {
	if !can(v1) || !can(v2) {
		panic(ErrInvalid)
	}
	return orderLiteral(v1, v2, f)
}

func orderLiteral[T cmp.Ordered](v1, v2 reflect.Value, f func(reflect.Value) T) int {
	val1 := f(v1)
	val2 := f(v2)
	return cmp.Compare(val1, val2)
}

func (o ordDefault) Ord(env Env, v1, v2 reflect.Value) int {
	t := v1.Type()
	switch t.Kind() {
	case reflect.Struct, // TODO: implement.
		reflect.Map,                  // TODO: maybe implement.
		reflect.Slice, reflect.Array, // TODO: implement.
		reflect.Chan, reflect.Func, reflect.UnsafePointer,
		reflect.Bool, // TODO: maybe implement.
		reflect.Uintptr,
		reflect.Complex128, reflect.Complex64:
		panic(ErrWrongType) // TODO: better error?
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return orderLiteralCan(v1, v2, reflect.Value.CanInt, reflect.Value.Int)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return orderLiteralCan(v1, v2, reflect.Value.CanUint, reflect.Value.Uint)
	case reflect.Float32, reflect.Float64:
		return orderLiteralCan(v1, v2, reflect.Value.CanFloat, reflect.Value.Float)
	case reflect.String:
		return orderLiteral(v1, v2, reflect.Value.String)
	case reflect.Pointer:
		return 0 // TODO: implement
	case reflect.Interface:
		return 0 // TODO: implement
	default:
		panic(ErrWrongType) // TODO: better error?
	}
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
