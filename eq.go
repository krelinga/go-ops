package ops

import (
	"fmt"
	"reflect"
)

type Eq interface {
	Eq(Env, reflect.Value, reflect.Value) bool
}

type eqTag struct{}

func EqualVals(env Env, v1, v2 reflect.Value) bool {
	if !v1.IsValid() || !v2.IsValid() {
		panic(ErrInvalid)
	}
	typ := v1.Type()
	if typ != v2.Type() {
		panic(ErrWrongType)
	}
	impl := func() Eq {
		if env == nil {
			return EqDefault{}
		}
		anyVal, ok := env.Get(typ, eqTag{})
		if !ok {
			return EqDefault{}
		}
		impl := anyVal.(Eq)
		if impl == nil {
			return EqDefault{}
		}
		return impl
	}()
	return impl.Eq(env, v1, v2)
}

func Equal[T any](env Env, in1, in2 T) bool {
	v1 := ValueFor(in1)
	v2 := ValueFor(in2)
	return EqualVals(env, v1, v2)
}

func TryEqualVals(env Env, v1, v2 reflect.Value) (bool, error) {
	var result bool
	err := try(func() {
		result = EqualVals(env, v1, v2)
	})
	if err != nil {
		return false, err
	}
	return result, nil
}

func TryEqual[T any](env Env, in1, in2 T) (bool, error) {
	var result bool
	err := try(func() {
		result = Equal(env, in1, in2)
	})
	if err != nil {
		return false, err
	}
	return result, nil
}

type EqDefault struct{}

func (EqDefault) Eq(env Env, v1, v2 reflect.Value) bool {
	t := v1.Type()
	switch t.Kind() {
	case reflect.Struct:
		return EqStruct{}.Eq(env, v1, v2)
	case reflect.Map:
		return EqMap{}.Eq(env, v1, v2)
	case reflect.Slice, reflect.Array:
		return EqSlice{}.Eq(env, v1, v2)
	case reflect.Ptr:
		return EqPointer{}.Eq(env, v1, v2)
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		// TODO: consider at least allowing both to be nil?
		panic(fmt.Errorf("%w: cannot compare type %s", ErrInvalid, typeName(t)))
	case reflect.Complex128, reflect.Complex64,
		reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.String, reflect.Bool:
		if !v1.CanInterface() || !v2.CanInterface() {
			panic(ErrInvalid) // TODO: better error?  Or handle via unsafe?
		}
		return v1.Interface() == v2.Interface()
	case reflect.Interface:
		return EqInterface{}.Eq(env, v1, v2)
	default:
		panic(fmt.Errorf("%w: unsupported kind %v for value %v", ErrInternal, t.Kind(), v1))
	}
}

type EqOptFunc func(Env, reflect.Value, reflect.Value) bool

func (f EqOptFunc) Eq(env Env, v1, v2 reflect.Value) bool {
	return f(env, v1, v2)
}

type EqTrue struct{}

func (EqTrue) Eq(_ Env, _, _ reflect.Value) bool {
	return true
}

type EqDeep struct{}

func (EqDeep) Eq(env Env, v1, v2 reflect.Value) bool {
	return EqualVals(env, v1, v2)
}

type EqPointer struct {
	Elem   Eq
	ByAddr bool
}

func (cp EqPointer) Eq(env Env, v1, v2 reflect.Value) bool {
	if v1.Kind() != reflect.Ptr {
		panic(ErrWrongType)
	}
	if cp.ByAddr {
		return v1.Pointer() == v2.Pointer()
	}
	if v1.IsNil() || v2.IsNil() {
		return v1.IsNil() && v2.IsNil()
	}
	elem := cp.Elem
	if elem == nil {
		elem = EqDeep{}
	}
	return elem.Eq(env, v1.Elem(), v2.Elem())
}

type EqInterface struct {
	Elem Eq
}

func (ci EqInterface) Eq(env Env, v1, v2 reflect.Value) bool {
	if v1.Kind() != reflect.Interface {
		panic(ErrWrongType)
	}
	if v1.IsNil() || v2.IsNil() {
		return v1.IsNil() && v2.IsNil()
	}
	e1 := v1.Elem()
	e2 := v2.Elem()
	if e1.Type() != e2.Type() {
		return false
	}
	elem := ci.Elem
	if elem == nil {
		elem = EqDeep{}
	}
	return elem.Eq(env, e1, e2)
}

type EqStruct struct {
	Fields map[Field]Eq
}

func (cs EqStruct) Eq(env Env, v1, v2 reflect.Value) bool {
	t := v1.Type()
	if t.Kind() != reflect.Struct {
		panic(ErrWrongType)
	}
	for f := range cs.Fields {
		if f == nil {
			panic(ErrNilField)
		}
	}
	for fNum := range t.NumField() {
		f := t.Field(fNum)
		if !f.IsExported() {
			// TODO: handle unexported fields?
			continue
		}
		var key Field
		if f.Anonymous {
			key = EmbedField(f.Type)
		} else {
			key = NamedField(f.Name)
		}
		impl, ok := cs.Fields[key]
		if !ok || impl == nil {
			impl = EqDeep{}
		}
		val1 := v1.Field(fNum)
		val2 := v2.Field(fNum)
		if !impl.Eq(env, val1, val2) {
			return false
		}
	}
	return true
}

type EqSlice struct {
	Elems Eq
	// TODO: support unordered comparison?
}

func (cs EqSlice) Eq(env Env, v1, v2 reflect.Value) bool {
	switch v1.Kind() {
	case reflect.Slice, reflect.Array:
		// ok
	default:
		panic(ErrWrongType)
	}
	if v1.Len() != v2.Len() {
		return false
	}
	if v1.IsNil() != v2.IsNil() {
		return false
	}
	elems := cs.Elems
	if elems == nil {
		elems = EqDeep{}
	}
	for elemNum := range v1.Len() {
		elem1 := v1.Index(elemNum)
		elem2 := v2.Index(elemNum)
		if !elems.Eq(env, elem1, elem2) {
			return false
		}
	}
	return true
}

type EqMap struct {
	Keys Eq
	Vals Eq
}

func (cm EqMap) Eq(env Env, v1, v2 reflect.Value) bool {
	if v1.Kind() != reflect.Map {
		panic(ErrWrongType)
	}
	if v1.Len() != v2.Len() {
		return false
	}
	if v1.IsNil() != v2.IsNil() {
		return false
	}
	keys := cm.Keys
	if keys == nil {
		keys = EqDeep{}
	}
	vals := cm.Vals
	if vals == nil {
		vals = EqDeep{}
	}

	type kv struct {
		key reflect.Value
		val reflect.Value
	}
	kvs1 := make([]kv, 0, v1.Len())
	i := v1.MapRange()
	for i.Next() {
		k := i.Key()
		v := i.Value()
		kvs1 = append(kvs1, kv{key: k, val: v})
	}
	used := make([]bool, len(kvs1))
	i = v2.MapRange()
k2loop:
	for i.Next() {
		k2 := i.Key()
		v2 := i.Value()
		for idx, kv1 := range kvs1 {
			if used[idx] {
				continue
			}
			if !keys.Eq(env, kv1.key, k2) || !vals.Eq(env, kv1.val, v2) {
				continue
			}
			used[idx] = true
			continue k2loop
		}
		return false
	}

	return true
}

func EqOpt(t reflect.Type, eq Eq) Opt {
	return OptFunc(func(env Env) {
		env.Set(t, eqTag{}, eq)
	})
}

func EqOptAll(eq Eq) Opt {
	return OptFunc(func(env Env) {
		env.SetAll(eqTag{}, eq)
	})
}
