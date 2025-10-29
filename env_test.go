package ops_test

import (
	"reflect"
	"testing"

	"github.com/krelinga/go-ops"
)

func TestNewEnv(t *testing.T) {
	env := ops.NewEnv()
	if env == nil {
		t.Fatal("NewEnv() returned nil")
	}
}

func TestMapEnv_Set(t *testing.T) {
	tests := []struct {
		name    string
		typ     reflect.Type
		tag     ops.Tag
		val     ops.Val
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid set",
			typ:  reflect.TypeOf(""),
			tag:  "test_tag",
			val:  "test_value",
		},
		{
			name:    "nil type",
			typ:     nil,
			tag:     "test_tag",
			val:     "test_value",
			wantErr: true,
			errMsg:  "type cannot be nil",
		},
		{
			name:    "nil tag",
			typ:     reflect.TypeOf(""),
			tag:     nil,
			val:     "test_value",
			wantErr: true,
			errMsg:  "tag cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := ops.NewEnv()

			if tt.wantErr {
				defer func() {
					if r := recover(); r != nil {
						if err, ok := r.(error); ok {
							if err.Error() != tt.errMsg {
								t.Errorf("Expected error message %q, got %q", tt.errMsg, err.Error())
							}
						} else {
							t.Errorf("Expected error, got %v", r)
						}
					} else {
						t.Error("Expected panic, but none occurred")
					}
				}()
			}

			env.Set(tt.typ, tt.tag, tt.val)

			if !tt.wantErr {
				// Verify the value was set correctly
				val, ok := env.Get(tt.typ, tt.tag)
				if !ok {
					t.Error("Expected value to be found")
				}
				if val != tt.val {
					t.Errorf("Expected value %v, got %v", tt.val, val)
				}
			}
		})
	}
}

func TestMapEnv_SetAll(t *testing.T) {
	tests := []struct {
		name    string
		tag     ops.Tag
		val     ops.Val
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid set all",
			tag:  "test_tag",
			val:  "test_value",
		},
		{
			name:    "nil tag",
			tag:     nil,
			val:     "test_value",
			wantErr: true,
			errMsg:  "tag cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := ops.NewEnv()

			if tt.wantErr {
				defer func() {
					if r := recover(); r != nil {
						if err, ok := r.(error); ok {
							if err.Error() != tt.errMsg {
								t.Errorf("Expected error message %q, got %q", tt.errMsg, err.Error())
							}
						} else {
							t.Errorf("Expected error, got %v", r)
						}
					} else {
						t.Error("Expected panic, but none occurred")
					}
				}()
			}

			env.SetAll(tt.tag, tt.val)

			if !tt.wantErr {
				// Verify the value is available for different types
				stringType := reflect.TypeOf("")
				intType := reflect.TypeOf(0)

				val1, ok1 := env.Get(stringType, tt.tag)
				if !ok1 {
					t.Error("Expected value to be found for string type")
				}
				if val1 != tt.val {
					t.Errorf("Expected value %v for string type, got %v", tt.val, val1)
				}

				val2, ok2 := env.Get(intType, tt.tag)
				if !ok2 {
					t.Error("Expected value to be found for int type")
				}
				if val2 != tt.val {
					t.Errorf("Expected value %v for int type, got %v", tt.val, val2)
				}
			}
		})
	}
}

func TestMapEnv_Get(t *testing.T) {
	tests := []struct {
		name    string
		typ     reflect.Type
		tag     ops.Tag
		wantErr bool
		errMsg  string
		setupFn func(ops.Env)
		wantVal ops.Val
		wantOk  bool
	}{
		{
			name: "get existing value",
			typ:  reflect.TypeOf(""),
			tag:  "test_tag",
			setupFn: func(env ops.Env) {
				env.Set(reflect.TypeOf(""), "test_tag", "test_value")
			},
			wantVal: "test_value",
			wantOk:  true,
		},
		{
			name: "get non-existing value",
			typ:  reflect.TypeOf(""),
			tag:  "non_existing_tag",
			setupFn: func(env ops.Env) {
				// Don't set anything
			},
			wantVal: nil,
			wantOk:  false,
		},
		{
			name: "get from SetAll",
			typ:  reflect.TypeOf(42),
			tag:  "all_tag",
			setupFn: func(env ops.Env) {
				env.SetAll("all_tag", "all_value")
			},
			wantVal: "all_value",
			wantOk:  true,
		},
		{
			name:    "nil type",
			typ:     nil,
			tag:     "test_tag",
			wantErr: true,
			errMsg:  "type cannot be nil",
		},
		{
			name:    "nil tag",
			typ:     reflect.TypeOf(""),
			tag:     nil,
			wantErr: true,
			errMsg:  "tag cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := ops.NewEnv()

			if tt.setupFn != nil {
				tt.setupFn(env)
			}

			if tt.wantErr {
				defer func() {
					if r := recover(); r != nil {
						if err, ok := r.(error); ok {
							if err.Error() != tt.errMsg {
								t.Errorf("Expected error message %q, got %q", tt.errMsg, err.Error())
							}
						} else {
							t.Errorf("Expected error, got %v", r)
						}
					} else {
						t.Error("Expected panic, but none occurred")
					}
				}()
			}

			val, ok := env.Get(tt.typ, tt.tag)

			if !tt.wantErr {
				if ok != tt.wantOk {
					t.Errorf("Expected ok=%v, got ok=%v", tt.wantOk, ok)
				}
				if val != tt.wantVal {
					t.Errorf("Expected value %v, got %v", tt.wantVal, val)
				}
			}
		})
	}
}

func TestWrapEnv(t *testing.T) {
	t.Run("wrap with no options", func(t *testing.T) {
		parent := ops.NewEnv()
		parent.Set(reflect.TypeOf(""), "parent_tag", "parent_value")

		wrapped := ops.WrapEnv(parent)

		// Should be able to get parent values
		val, ok := wrapped.Get(reflect.TypeOf(""), "parent_tag")
		if !ok {
			t.Error("Expected to find parent value")
		}
		if val != "parent_value" {
			t.Errorf("Expected parent_value, got %v", val)
		}
	})

	t.Run("wrap with options", func(t *testing.T) {
		parent := ops.NewEnv()

		opt := ops.OptFunc(func(env ops.Env) {
			env.Set(reflect.TypeOf(""), "opt_tag", "opt_value")
		})

		wrapped := ops.WrapEnv(parent, opt)

		// Should be able to get option-set values
		val, ok := wrapped.Get(reflect.TypeOf(""), "opt_tag")
		if !ok {
			t.Error("Expected to find option value")
		}
		if val != "opt_value" {
			t.Errorf("Expected opt_value, got %v", val)
		}
	})

	t.Run("child overrides parent", func(t *testing.T) {
		parent := ops.NewEnv()
		parent.Set(reflect.TypeOf(""), "tag", "parent_value")

		wrapped := ops.WrapEnv(parent)
		wrapped.Set(reflect.TypeOf(""), "tag", "child_value")

		// Child value should override parent
		val, ok := wrapped.Get(reflect.TypeOf(""), "tag")
		if !ok {
			t.Error("Expected to find value")
		}
		if val != "child_value" {
			t.Errorf("Expected child_value, got %v", val)
		}

		// Parent should still have original value
		parentVal, parentOk := parent.Get(reflect.TypeOf(""), "tag")
		if !parentOk {
			t.Error("Expected to find parent value")
		}
		if parentVal != "parent_value" {
			t.Errorf("Expected parent_value, got %v", parentVal)
		}
	})

	t.Run("fallback to parent", func(t *testing.T) {
		parent := ops.NewEnv()
		parent.Set(reflect.TypeOf(""), "parent_only_tag", "parent_value")

		wrapped := ops.WrapEnv(parent)
		wrapped.Set(reflect.TypeOf(""), "child_only_tag", "child_value")

		// Should find parent value when not in child
		val, ok := wrapped.Get(reflect.TypeOf(""), "parent_only_tag")
		if !ok {
			t.Error("Expected to find parent value")
		}
		if val != "parent_value" {
			t.Errorf("Expected parent_value, got %v", val)
		}

		// Should find child value
		childVal, childOk := wrapped.Get(reflect.TypeOf(""), "child_only_tag")
		if !childOk {
			t.Error("Expected to find child value")
		}
		if childVal != "child_value" {
			t.Errorf("Expected child_value, got %v", childVal)
		}
	})
}

func TestWrappedEnv_SetAll(t *testing.T) {
	parent := ops.NewEnv()
	wrapped := ops.WrapEnv(parent)

	wrapped.SetAll("all_tag", "all_value")

	// Should be available for any type
	val, ok := wrapped.Get(reflect.TypeOf(""), "all_tag")
	if !ok {
		t.Error("Expected to find SetAll value")
	}
	if val != "all_value" {
		t.Errorf("Expected all_value, got %v", val)
	}
}

func TestSetOverridesSetAll(t *testing.T) {
	env := ops.NewEnv()
	stringType := reflect.TypeOf("")

	// First SetAll
	env.SetAll("tag", "all_value")

	// Verify SetAll works initially
	val, ok := env.Get(stringType, "tag")
	if !ok {
		t.Error("Expected to find SetAll value initially")
	}
	if val != "all_value" {
		t.Errorf("Expected all_value initially, got %v", val)
	}

	// Then Set for specific type - this replaces the valForAllTypes with mapTypeToVal
	env.Set(stringType, "tag", "specific_value")

	// Specific value should be found
	val, ok = env.Get(stringType, "tag")
	if !ok {
		t.Error("Expected to find value")
	}
	if val != "specific_value" {
		t.Errorf("Expected specific_value, got %v", val)
	}

	// Other types should not find anything (mapTypeToVal only has string type)
	intType := reflect.TypeOf(0)
	intVal, intOk := env.Get(intType, "tag")
	if intOk {
		t.Errorf("Expected not to find value for int type, but got %v", intVal)
	}
}

func TestSetAllOverridesSet(t *testing.T) {
	env := ops.NewEnv()
	stringType := reflect.TypeOf("")

	// First Set for specific type
	env.Set(stringType, "tag", "specific_value")

	// Then SetAll
	env.SetAll("tag", "all_value")

	// SetAll should override specific Set
	val, ok := env.Get(stringType, "tag")
	if !ok {
		t.Error("Expected to find value")
	}
	if val != "all_value" {
		t.Errorf("Expected all_value, got %v", val)
	}
}
