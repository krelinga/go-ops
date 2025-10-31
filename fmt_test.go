package ops_test

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/krelinga/go-ops"
)

type testInt int

func (ti testInt) String() string {
	return fmt.Sprintf("testInt of %d", int(ti))
}

func TestFormatFor(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		tests := []struct {
			name string
			f    func() string
			want string
		}{
			{
				name: "Integer",
				f: func() string {
					return ops.FormatFor(nil, 42)
				},
				want: "42",
			},
			{
				name: "String",
				f: func() string {
					return ops.FormatFor(nil, "hello")
				},
				want: `"hello"`,
			},
			{
				name: "Boolean",
				f: func() string {
					return ops.FormatFor(nil, true)
				},
				want: "true",
			},
			{
				name: "Float",
				f: func() string {
					return ops.FormatFor(nil, 3.14)
				},
				want: "3.14",
			},
			{
				name: "Complex",
				f: func() string {
					return ops.FormatFor(nil, 1+2i)
				},
				want: "(1+2i)",
			},
			{
				name: "Nil Pointer",
				f: func() string {
					var p *int
					return ops.FormatFor(nil, p)
				},
				want: "<nil>",
			},
			{
				name: "Nil Any Interface",
				f: func() string {
					var i any
					return ops.FormatFor(nil, i)
				},
				want: "any(nil)",
			},
			{
				name: "Nil Anonymous Interface",
				f: func() string {
					var i interface {
						Foo()
					}
					return ops.FormatFor(nil, i)
				},
				want: "interface { Foo() }(nil)",
			},
			{
				name: "Nil Stringer Interface",
				f: func() string {
					var s fmt.Stringer
					return ops.FormatFor(nil, s)
				},
				want: "fmt.Stringer(nil)",
			},
			{
				name: "Custom Stringer",
				f: func() string {
					var s fmt.Stringer = testInt(7)
					return ops.FormatFor(nil, s)
				},
				want: "fmt.Stringer(ops_test.testInt(7))",
			},
			{
				name: "Func",
				f: func() string {
					fn := func() {}
					return ops.FormatFor(nil, fn)
				},
				want: "func()(...)",
			},
			{
				name: "Channel",
				f: func() string {
					ch := make(chan int)
					return ops.FormatFor(nil, ch)
				},
				want: "chan int(...)",
			},
			{
				name: "Send-Only Channel",
				f: func() string {
					ch := make(chan<- int)
					return ops.FormatFor(nil, ch)
				},
				want: "chan<- int(...)",
			},
			{
				name: "Receive-Only Channel",
				f: func() string {
					ch := make(<-chan int)
					return ops.FormatFor(nil, ch)
				},
				want: "<-chan int(...)",
			},
			{
				name: "Unsafe Pointer",
				f: func() string {
					var up unsafe.Pointer
					return ops.FormatFor(nil, up)
				},
				want: "unsafe.Pointer(...)",
			},
			{
				name: "Slice",
				f: func() string {
					slice := []int{1, 2, 3}
					return ops.FormatFor(nil, slice)
				},
				want: `[]int{
  1,
  2,
  3,
}`,
			},
			{
				name: "Array",
				f: func() string {
					array := [3]string{"a", "b", "c"}
					return ops.FormatFor(nil, array)
				},
				want: `[3]string{
  "a",
  "b",
  "c",
}`,
			},
			{
				name: "Map",
				f: func() string {
					m := map[string]int{"one": 1}
					return ops.FormatFor(nil, m)
				},
				want: `map[string]int{
  "one": 1,
}`,
			},
			{
				name: "Struct",
				f: func() string {
					type Person struct {
						Name string
						Age  int
					}
					p := Person{Name: "Alice", Age: 30}
					return ops.FormatFor(nil, p)
				},
				want: `ops_test.Person{
  Name: "Alice",
  Age: 30,
}`,
			},
			{
				name: "Pointer to Struct",
				f: func() string {
					type Point struct {
						X, Y int
					}
					p := &Point{X: 10, Y: 20}
					return ops.FormatFor(nil, p)
				},
				want: `&ops_test.Point{
  X: 10,
  Y: 20,
}`,
			},
			{
				name: "Pointer to String",
				f: func() string {
					str := "hello"
					p := &str
					return ops.FormatFor(nil, p)
				},
				want: `&"hello"`, // TODO: I don't like the way this looks.
			},
			{
				name: "Uintptr",
				f: func() string {
					var up uintptr = 0x1234abcd
					return ops.FormatFor(nil, up)
				},
				want: "0x1234abcd",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.f()
				if got != tt.want {
					t.Errorf("got %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("With FmtStruct", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		p := Person{Name: "Bob", Age: 25}
		type Animal struct {
			Species string
			OwnedBy *Person
			isGood  bool
		}
		a := Animal{Species: "Dog", OwnedBy: &p, isGood: true}
		tests := []struct {
			name string
			opt  ops.Opt
			want string
		}{
			{
				name: "Elide A Field",
				opt: ops.FmtOpt(reflect.TypeFor[Person](), ops.FmtStruct{
					Fields: map[ops.Field]ops.Fmt{
						ops.NamedField("Age"): ops.FmtElide{},
					},
				}),
				want: `ops_test.Animal{
  Species: "Dog",
  OwnedBy: &ops_test.Person{
    Name: "Bob",
    Age: int(...),
  },
  ...
}`,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				env := ops.WrapEnv(ops.NewEnv(), tt.opt)
				got := ops.FormatFor(env, a)
				if got != tt.want {
					t.Errorf("got %q, want %q", got, tt.want)
				}
			})
		}
	})
}
