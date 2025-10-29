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

type mapEnv map[Tag]map[reflect.Type]Val

func (e *mapEnv) Set(typ reflect.Type, tag Tag, val Val) {
	if typ == nil {
		panic(ErrNilType)
	}
	if tag == nil {
		panic(ErrNilTag)
	}
	e.setImpl(typ, tag, val)
}

func (e *mapEnv) SetAll(tag Tag, val Val) {
	if tag == nil {
		panic(ErrNilTag)
	}
	e.setImpl(nil, tag, val)
}

func (e *mapEnv) setImpl(typ reflect.Type, tag Tag, val Val) {
	if *e == nil {
		*e = make(mapEnv)
	}
	if (*e)[tag] == nil {
		(*e)[tag] = make(map[reflect.Type]Val)
	}
	(*e)[tag][typ] = val
}

func (e *mapEnv) Get(typ reflect.Type, tag Tag) (Val, bool) {
	if typ == nil {
		panic(ErrNilType)
	}
	if tag == nil {
		panic(ErrNilTag)
	}
	if typeMap, ok := (*e)[tag]; ok {
		if allVal, ok := typeMap[nil]; ok {
			return allVal, true
		}
		if val, ok := typeMap[typ]; ok {
			return val, true
		}
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