package ops

import (
	"fmt"
	"reflect"
	"strings"
)

type fmtTag struct{}

type Fmt interface {
	Fmt(Env, reflect.Value) string
}

type FmtFunc func(Env, reflect.Value) string

func (ff FmtFunc) Fmt(env Env, v reflect.Value) string {
	return ff(env, v)
}

type FmtElide struct{}

func (FmtElide) Fmt(_ Env, v reflect.Value) string {
	return fmt.Sprintf("%s(...)", typeName(v.Type()))
}

func FormatVal(env Env, v reflect.Value) string {
	if !v.IsValid() {
		return "<invalid>"
	}
	impl := func() Fmt {
		if env == nil {
			return fmtDefault{}
		}
		anyVal, ok := env.Get(v.Type(), fmtTag{})
		if !ok {
			return fmtDefault{}
		}
		impl := anyVal.(Fmt)
		if impl == nil {
			return fmtDefault{}
		}
		return impl
	}()
	return impl.Fmt(env, v)
}

func Format[T any](env Env, in T) string {
	v := ValueFor(in)
	return FormatVal(env, v)
}

func TryFormatVals(env Env, v reflect.Value) (string, error) {
	var str string
	err := try(func() {
		str = FormatVal(env, v)
	})
	if err != nil {
		return "", err
	}
	return str, nil
}

func TryFormat[T any](env Env, in T) (string, error) {
	var str string
	err := try(func() {
		str = Format(env, in)
	})
	if err != nil {
		return "", err
	}
	return str, nil
}

type FmtDeep struct{}

func (FmtDeep) Fmt(env Env, v reflect.Value) string {
	return FormatVal(env, v)
}

type fmtDefault struct{}

func literalStringCan[T any](v reflect.Value, can func(reflect.Value) bool, f func(reflect.Value) T) string {
	if !can(v) {
		return FmtElide{}.Fmt(nil, v)
	}
	return literalString(v, f)
}

var directTypes = map[reflect.Type]struct{}{
	reflect.TypeFor[bool]():       {},
	reflect.TypeFor[int]():        {},
	reflect.TypeFor[int8]():       {},
	reflect.TypeFor[int16]():      {},
	reflect.TypeFor[int32]():      {},
	reflect.TypeFor[int64]():      {},
	reflect.TypeFor[uint]():       {},
	reflect.TypeFor[uint8]():      {},
	reflect.TypeFor[uint16]():     {},
	reflect.TypeFor[uint32]():     {},
	reflect.TypeFor[uint64]():     {},
	reflect.TypeFor[uintptr]():    {},
	reflect.TypeFor[float32]():    {},
	reflect.TypeFor[float64]():    {},
	reflect.TypeFor[complex64]():  {},
	reflect.TypeFor[complex128](): {},
	reflect.TypeFor[string]():     {},
}

func literalString[T any](v reflect.Value, f func(reflect.Value) T) string {
	if _, ok := directTypes[v.Type()]; ok {
		return fmt.Sprintf("%#v", f(v))
	} else {
		return fmt.Sprintf("%s(%#v)", typeName(v.Type()), f(v))
	}
}

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
		return literalStringCan(v, reflect.Value.CanComplex, reflect.Value.Complex)
	case reflect.Float32, reflect.Float64:
		return literalStringCan(v, reflect.Value.CanFloat, reflect.Value.Float)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return literalStringCan(v, reflect.Value.CanInt, reflect.Value.Int)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return literalStringCan(v, reflect.Value.CanUint, reflect.Value.Uint)
	case reflect.Bool:
		return literalString(v, reflect.Value.Bool)
	case reflect.String:
		return literalString(v, reflect.Value.String)
	case reflect.Interface:
		return FmtInterface{}.Fmt(env, v)
	default:
		panic(fmt.Errorf("%w: unsupported kind %v for value %v", ErrInternal, v.Kind(), v))
	}
}

type FmtStruct struct {
	Fields map[Field]Fmt
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
			name = typeName(f.Type)
		} else {
			key = NamedField(f.Name)
			name = f.Name
		}
		impl, ok := sf.Fields[key]
		if !ok || impl == nil {
			impl = FmtDeep{}
		}
		val := v.Field(fNum)
		fieldStrings = append(fieldStrings, fmt.Sprintf("%s: %s,", name, impl.Fmt(env, val)))
	}
	if filtered {
		fieldStrings = append(fieldStrings, "...")
	}
	for i := range fieldStrings {
		fieldStrings[i] = indent(fieldStrings[i])
	}
	return fmt.Sprintf("%s{\n%s\n}", typeName(t), strings.Join(fieldStrings, "\n"))
}

type FmtMap struct {
	Keys Fmt
	Vals Fmt
	// TODO: possibly add an option to limit number of entries?  or to elide certain keys?
}

func (mf FmtMap) Fmt(env Env, v reflect.Value) string {
	t := v.Type()
	if t.Kind() != reflect.Map {
		panic(ErrWrongType)
	}
	keys := mf.Keys
	if keys == nil {
		keys = FmtDeep{}
	}
	vals := mf.Vals
	if vals == nil {
		vals = FmtDeep{}
	}
	entryStrings := make([]string, 0, v.Len())
	i := v.MapRange()
	for i.Next() {
		k := i.Key()
		val := i.Value()
		entryStrings = append(entryStrings, fmt.Sprintf("%s: %s,", keys.Fmt(env, k), vals.Fmt(env, val)))
	}
	for i := range entryStrings {
		entryStrings[i] = indent(entryStrings[i])
	}
	return fmt.Sprintf("%s{\n%s\n}", typeName(t), strings.Join(entryStrings, "\n"))
}

type FmtSlice struct {
	Elems Fmt
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
	elems := sf.Elems
	if elems == nil {
		elems = FmtDeep{}
	}
	elementStrings := make([]string, 0, v.Len())
	for i := range v.Len() {
		elem := v.Index(i)
		elementStrings = append(elementStrings, fmt.Sprintf("%s,", elems.Fmt(env, elem)))
	}
	for i := range elementStrings {
		elementStrings[i] = indent(elementStrings[i])
	}
	return fmt.Sprintf("%s{\n%s\n}", t, strings.Join(elementStrings, "\n"))
}

type FmtPointer struct {
	Elem Fmt
}

func (pf FmtPointer) Fmt(env Env, v reflect.Value) string {
	t := v.Type()
	if t.Kind() != reflect.Pointer {
		panic(ErrWrongType)
	}
	if v.IsNil() {
		return "<nil>"
	}
	impl := pf.Elem
	if impl == nil {
		impl = FmtDeep{}
	}
	elem := v.Elem()
	return fmt.Sprintf("&%s", impl.Fmt(env, elem))
}

type FmtInterface struct {
	Elem Fmt
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
		impl := fmtI.Elem
		if impl == nil {
			impl = FmtDeep{}
		}
		elem := v.Elem()
		elemStr = impl.Fmt(env, elem)
	}
	return fmt.Sprintf("%s(%s)", typeName(t), elemStr)
}

type FmtWrap struct {
	Opt  Opt
	Then Fmt
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

func FmtOpt(typ reflect.Type, fmt Fmt) Opt {
	return OptFunc(func(e Env) {
		e.Set(typ, fmtTag{}, fmt)
	})
}

func FmtOptAll(fmt Fmt) Opt {
	return OptFunc(func(e Env) {
		e.SetAll(fmtTag{}, fmt)
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

func FmtOptStringer[T fmt.Stringer]() Opt {
	return FmtOpt(reflect.TypeFor[T](), FmtStringer{})
}
