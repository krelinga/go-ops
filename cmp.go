package ops

import "reflect"

type Cmper interface {
	Cmp(Env, reflect.Value, reflect.Value) bool
}

type cmpTag struct{}

func Cmp(env Env, v1, v2 reflect.Value) bool {
	if !v1.IsValid() || !v2.IsValid() {
		panic(ErrInvalid)
	}
	typ := v1.Type()
	if typ != v2.Type() {
		panic(ErrWrongType)
	}
	cmper := func() Cmper {
		if env == nil {
			return cmpDefault{}
		}
		anyVal, ok := env.Get(typ, cmpTag{})
		if !ok {
			return cmpDefault{}
		}
		cmper := anyVal.(Cmper)
		if cmper == nil {
			return cmpDefault{}
		}
		return cmper
	}()
	return cmper.Cmp(env, v1, v2)
}

func CmpFor[T any](env Env, in1, in2 T) bool {
	v1 := ValueFor(in1)
	v2 := ValueFor(in2)
	return Cmp(env, v1, v2)
}

func TryCmp(env Env, v1, v2 reflect.Value) (bool, error) {
	var result bool
	err := try(func() {
		result = Cmp(env, v1, v2)
	})
	if err != nil {
		return false, err
	}
	return result, nil
}

func TryCmpFor[T any](env Env, in1, in2 T) (bool, error) {
	var result bool
	err := try(func() {
		result = CmpFor(env, in1, in2)
	})
	if err != nil {
		return false, err
	}
	return result, nil
}

type cmpDefault struct{}

func (cmpDefault) Cmp(env Env, v1, v2 reflect.Value) bool {
	return false // TODO
}

type CmpOptFunc func(Env, reflect.Value, reflect.Value) bool

func (f CmpOptFunc) Cmp(env Env, v1, v2 reflect.Value) bool {
	return f(env, v1, v2)
}

type CmpTrue struct{}

func (CmpTrue) Cmp(_ Env, _, _ reflect.Value) bool {
	return true
}

type CmpDeep struct{}

func (CmpDeep) Cmp(env Env, v1, v2 reflect.Value) bool {
	return Cmp(env, v1, v2)
}

type CmpPointer struct {
	Elem   Cmper
	ByAddr bool
}

func (cp CmpPointer) Cmp(env Env, v1, v2 reflect.Value) bool {
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
		elem = CmpDeep{}
	}
	return elem.Cmp(env, v1.Elem(), v2.Elem())
}

type CmpInterface struct {
	Elem Cmper
}

func (ci CmpInterface) Cmp(env Env, v1, v2 reflect.Value) bool {
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
		elem = CmpDeep{}
	}
	return elem.Cmp(env, e1, e2)
}

type CmpStruct struct {
	Fields map[Field]Cmper
}

func (cs CmpStruct) Cmp(env Env, v1, v2 reflect.Value) bool {
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
		cmper, ok := cs.Fields[key]
		if !ok || cmper == nil {
			cmper = CmpDeep{}
		}
		val1 := v1.Field(fNum)
		val2 := v2.Field(fNum)
		if !cmper.Cmp(env, val1, val2) {
			return false
		}
	}
	return true
}

type CmpSlice struct {
	Elems Cmper
	// TODO: support unordered comparison?
}

func (cs CmpSlice) Cmp(env Env, v1, v2 reflect.Value) bool {
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
		elems = CmpDeep{}
	}
	for elemNum := range v1.Len() {
		elem1 := v1.Index(elemNum)
		elem2 := v2.Index(elemNum)
		if !elems.Cmp(env, elem1, elem2) {
			return false
		}
	}
	return true
}

type CmpMap struct {
	Keys Cmper
	Vals Cmper
}

func (cm CmpMap) Cmp(env Env, v1, v2 reflect.Value) bool {
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
		keys = CmpDeep{}
	}
	vals := cm.Vals
	if vals == nil {
		vals = CmpDeep{}
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
			if !keys.Cmp(env, kv1.key, k2) || !vals.Cmp(env, kv1.val, v2) {
				continue
			}
			used[idx] = true
			continue k2loop
		}
		return false
	}

	return true
}

func CmpOpt(t reflect.Type, cmper Cmper) Opt {
	return OptFunc(func(env Env) {
		env.Set(t, cmpTag{}, cmper)
	})
}

func CmpOptAll(cmper Cmper) Opt {
	return OptFunc(func(env Env) {
		env.SetAll(cmpTag{}, cmper)
	})
}
