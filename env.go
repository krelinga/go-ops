package ops

import "reflect"

type Tag any
type Val any

type Env interface {
	Set(reflect.Type, Tag, Val)
	SetAll(Tag, Val)
	Get(reflect.Type, Tag) (Val, bool)
}

func NewEnv() Env {
	return &mapEnv{}
}

type typeToVal interface {
	Get(reflect.Type) (Val, bool)
}

type mapTypeToVal map[reflect.Type]Val

func (m mapTypeToVal) Get(typ reflect.Type) (Val, bool) {
	val, ok := m[typ]
	return val, ok
}

type valForAllTypes struct {
	val Val
}

func (v valForAllTypes) Get(_ reflect.Type) (Val, bool) {
	return v.val, true
}

type mapEnv map[Tag]typeToVal

func (e *mapEnv) Set(typ reflect.Type, tag Tag, val Val) {
	if typ == nil {
		panic(ErrNilType)
	}
	if tag == nil {
		panic(ErrNilTag)
	}
	if *e == nil {
		*e = make(map[Tag]typeToVal)
	}
	ttvMap := func() mapTypeToVal {
		if ttv, ok := (*e)[tag]; ok {
			if ttvMap, ok := ttv.(mapTypeToVal); ok {
				return ttvMap
			}
		}
		ttvMap := make(mapTypeToVal)
		(*e)[tag] = ttvMap
		return ttvMap
	}()
	ttvMap[typ] = val
}

func (e *mapEnv) SetAll(tag Tag, val Val) {
	if tag == nil {
		panic(ErrNilTag)
	}
	if *e == nil {
		*e = make(map[Tag]typeToVal)
	}
	(*e)[tag] = valForAllTypes{val: val}
}

func (e *mapEnv) Get(typ reflect.Type, tag Tag) (Val, bool) {
	if typ == nil {
		panic(ErrNilType)
	}
	if tag == nil {
		panic(ErrNilTag)
	}
	if typeMap, ok := (*e)[tag]; ok {
		return typeMap.Get(typ)
	}
	return nil, false
}

type wrappedEnv struct {
	parent Env
	data *mapEnv
}

func (e *wrappedEnv) Set(typ reflect.Type, tag Tag, val Val) {
	e.data.Set(typ, tag, val)
}

func (e *wrappedEnv) SetAll(tag Tag, val Val) {
	e.data.SetAll(tag, val)
}

func (e *wrappedEnv) Get(typ reflect.Type, tag Tag) (Val, bool) {
	if val, ok := e.data.Get(typ, tag); ok {
		return val, true
	}
	return e.parent.Get(typ, tag)
}

func WrapEnv(parent Env, opts ...Opt) Env {
	we := &wrappedEnv{
		parent: parent,
		data:   &mapEnv{},
	}
	for _, opt := range opts {
		opt.Update(we)
	}
	return we
}