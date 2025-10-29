package ops

import "reflect"

type Field interface {
	fieldIsAClosedType()
}

type namedField string

func (nf namedField) fieldIsAClosedType() {}

func NamedField(name string) Field {
	return namedField(name)
}

func EmbedField(typ reflect.Type) Field {
	return embedField{Typ: typ}
}

func EmbedFieldFor[T any]() Field {
	return EmbedField(reflect.TypeFor[T]())
}

type embedField struct {
	Typ reflect.Type
}

func (ef embedField) fieldIsAClosedType() {}
