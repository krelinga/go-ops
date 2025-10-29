package ops

import (
	"fmt"
	"reflect"
	"strings"
)

type fmtTag struct{}

type Fmter interface {
	Fmt(Env, reflect.Value) string
}

type FmtFunc func(Env, reflect.Value) string

func (ff FmtFunc) Fmt(env Env, v reflect.Value) string {
	return ff(env, v)
}

type FmtElide struct{}

func (FmtElide) Fmt(_ Env, _ reflect.Value) string {
	return "..."
}

func Fmt(env Env, v reflect.Value) string {
	if !v.IsValid() {
		return "<invalid>"
	}
	fmter := func() Fmter {
		if env == nil {
			return fmtDefault{}
		}
		anyVal, ok := env.Get(v.Type(), fmtTag{})
		if !ok {
			return fmtDefault{}
		}
		fmter := anyVal.(Fmter)
		if fmter == nil {
			return fmtDefault{}
		}
		return fmter
	}()
	return fmter.Fmt(env, v)
}

func FmtFor[T any](env Env, in T) string {
	v := ValueFor(in)
	return Fmt(env, v)
}

type FmtDeep struct{}

func (FmtDeep) Fmt(env Env, v reflect.Value) string {
	return Fmt(env, v)
}

type fmtDefault struct{}

func (fmtDefault) Fmt(env Env, v reflect.Value) string {
	switch v.Kind() {
	case reflect.Struct:
		return FmtStruct{}.Fmt(env, v)
	case reflect.Map:
		return FmtMap{}.Fmt(env, v)
	case reflect.Slice, reflect.Array:
		return FmtSlice{}.Fmt(env, v)
	case reflect.Pointer:
		return FmtPointer{}.Fmt(env, v)
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return FmtElide{}.Fmt(env, v)
	case reflect.Complex128, reflect.Complex64:
		if !v.CanComplex() {
			return FmtElide{}.Fmt(env, v)
		}
		return fmt.Sprintf("%#v", v.Complex())
	case reflect.Float32, reflect.Float64:
		if !v.CanFloat() {
			return FmtElide{}.Fmt(env, v)
		}
		return fmt.Sprintf("%#v", v.Float())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if !v.CanInt() {
			return FmtElide{}.Fmt(env, v)
		}
		return fmt.Sprintf("%#v", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if !v.CanUint() {
			return FmtElide{}.Fmt(env, v)
		}
		return fmt.Sprintf("%#v", v.Uint())
	case reflect.Bool:
		return fmt.Sprintf("%#v", v.Bool())
	case reflect.String:
		return fmt.Sprintf("%#v", v.String())
	case reflect.Interface:
		return FmtInterface{}.Fmt(env, v)
	default:
		panic(fmt.Sprintf("unsupported kind %v for value %v", v.Kind(), v))
	}
}

type FmtStruct struct {
	Fields map[Field]Fmter
	// TODO: possibly add an option to include unexported fields?
}

func (sf FmtStruct) Fmt(env Env, v reflect.Value) string {
	t := v.Type()
	if t.Kind() != reflect.Struct {
		panic(ErrWrongType)
	}
	for f := range sf.Fields {
		if f == nil {
			panic(ErrNilField)
		}
	}
	// TODO: check for entries in sf.Fields that don't exist in t?
	var filtered bool
	fieldStrings := make([]string, 0, t.NumField())
	for fNum := range t.NumField() {
		f := t.Field(fNum)
		if !f.IsExported() {
			filtered = true
			continue
		}
		var key Field
		var name string
		if f.Anonymous {
			key = EmbedField(f.Type)
			name = f.Type.Name()
		} else {
			key = NamedField(f.Name)
			name = f.Name
		}
		fmter, ok := sf.Fields[key]
		if !ok || fmter == nil {
			fmter = FmtDeep{}
		}
		val := v.Field(fNum)
		fieldStrings = append(fieldStrings, fmt.Sprintf("%s: %s,", name, fmter.Fmt(env, val)))
	}
	if filtered {
		fieldStrings = append(fieldStrings, "...")
	}
	for i := range fieldStrings {
		fieldStrings[i] = indent(fieldStrings[i])
	}
	return fmt.Sprintf("%s {\n%s\n}", t.Name(), strings.Join(fieldStrings, "\n"))
}

type FmtMap struct {
	Keys Fmter
	Vals Fmter
	// TODO: possibly add an option to limit number of entries?  or to elide certain keys?
}

func (mf FmtMap) Fmt(env Env, v reflect.Value) string {
	t := v.Type()
	if t.Kind() != reflect.Map {
		panic(ErrWrongType)
	}
	keyFmter := mf.Keys
	if keyFmter == nil {
		keyFmter = FmtDeep{}
	}
	valueFmter := mf.Vals
	if valueFmter == nil {
		valueFmter = FmtDeep{}
	}
	entryStrings := make([]string, 0, v.Len())
	i := v.MapRange()
	for i.Next() {
		k := i.Key()
		val := i.Value()
		entryStrings = append(entryStrings, fmt.Sprintf("%s: %s,", keyFmter.Fmt(env, k), valueFmter.Fmt(env, val)))
	}
	for i := range entryStrings {
		entryStrings[i] = indent(entryStrings[i])
	}
	return fmt.Sprintf("%s{\n%s\n}", t.Name(), strings.Join(entryStrings, "\n"))
}

type FmtSlice struct {
	Elems Fmter
	// TODO: possibly add an option to limit number of elements?
}

func (sf FmtSlice) Fmt(env Env, v reflect.Value) string {
	t := v.Type()
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		// ok
	default:
		panic(ErrWrongType)
	}
	elementFmter := sf.Elems
	if elementFmter == nil {
		elementFmter = FmtDeep{}
	}
	elementStrings := make([]string, 0, v.Len())
	for i := range v.Len() {
		elem := v.Index(i)
		elementStrings = append(elementStrings, fmt.Sprintf("%s,", elementFmter.Fmt(env, elem)))
	}
	for i := range elementStrings {
		elementStrings[i] = indent(elementStrings[i])
	}
	return fmt.Sprintf("%s{\n%s\n}", t.Name(), strings.Join(elementStrings, "\n"))
}

type FmtPointer struct {
	Elem Fmter
}

func (pf FmtPointer) Fmt(env Env, v reflect.Value) string {
	t := v.Type()
	if t.Kind() != reflect.Pointer {
		panic(ErrWrongType)
	}
	if v.IsNil() {
		return "<nil>"
	}
	elementFmter := pf.Elem
	if elementFmter == nil {
		elementFmter = FmtDeep{}
	}
	elem := v.Elem()
	return fmt.Sprintf("&%s", elementFmter.Fmt(env, elem))
}

type FmtInterface struct {
	Elem Fmter
}

func (fmtI FmtInterface) Fmt(env Env, v reflect.Value) string {
	t := v.Type()
	if t.Kind() != reflect.Interface {
		panic(ErrWrongType)
	}
	var elemStr string
	if v.IsNil() {
		elemStr = "nil"
	} else {
		elementFmter := fmtI.Elem
		if elementFmter == nil {
			elementFmter = FmtDeep{}
		}
		elem := v.Elem()
		elemStr = elementFmter.Fmt(env, elem)
	}
	return fmt.Sprintf("%s(%s)", t.Name(), elemStr)
}

type FmtWrap struct {
	Opt  Opt
	Then Fmter
}

func (fw FmtWrap) Fmt(env Env, v reflect.Value) string {
	if fw.Opt != nil {
		env = WrapEnv(env, fw.Opt)
	}
	then := fw.Then
	if then == nil {
		then = FmtDeep{}
	}
	return then.Fmt(env, v)

}

type FmtStringer struct{}

func (FmtStringer) Fmt(_ Env, v reflect.Value) string {
	if !v.CanInterface() {
		return "<uninterfaceable>"
	}
	str, ok := v.Interface().(fmt.Stringer)
	if !ok {
		panic(ErrWrongType)
	}
	return str.String()
}

func FmtOpt(typ reflect.Type, fmter Fmter) Opt {
	return OptFunc(func(e Env) {
		e.Set(typ, fmtTag{}, fmter)
	})
}

func FmtOptAll(fmter Fmter) Opt {
	return OptFunc(func(e Env) {
		e.SetAll(fmtTag{}, fmter)
	})
}

func FmtOptFor[T any](typed func(Env, T) string) Opt {
	typedFunc := func(env Env, v reflect.Value) string {
		if !v.CanInterface() {
			return "<uninterfaceable>"
		}
		return typed(env, v.Interface().(T))
	}
	return FmtOpt(reflect.TypeFor[T](), FmtFunc(typedFunc))
}
