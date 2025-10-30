package ops

import (
	"reflect"
	"strings"
)

func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = "  " + lines[i]
	}
	return strings.Join(lines, "\n")
}

func ValueFor[T any](in T) reflect.Value {
	typ := reflect.TypeFor[T]()
	v := reflect.ValueOf(in)
	if typ.Kind() == reflect.Interface {
		newV := reflect.New(typ).Elem()
		if v.IsValid() {
			newV.Set(v)
		}
		v = newV
	}
	return v
}

func typeName(t reflect.Type) string {
	if t == reflect.TypeFor[any]() {
		return "any"
	}
	return t.String()
}