package spelunk_test

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"hhj.me/go/spelunk"
)

func ExampleSpelunker_Spelunk_zeroer() {
	example := struct {
		String string            `spelunk:"zero"`
		Int    int               `spelunk:"zero"`
		Float  float64           `spelunk:"zero"`
		Map    map[string]string `spelunk:"zero"`
		Slice  []string          `spelunk:"zero"`
		Array  [2]string         `spelunk:"zero"`
		Func   func() error      `spelunk:"zero"`
	}{
		String: "string",
		Int:    1,
		Float:  1.2,
		Map:    map[string]string{"key": "value"},
		Slice:  []string{"foo", "bar"},
		Array:  [2]string{"baz", "qux"},
		Func:   func() error { return nil },
	}

	fmt.Printf("%+v\n", example)

	s1 := spelunk.New()

	s1.SetHandler("zero", spelunk.Zeroer)
	s1.Spelunk(&example)

	fmt.Printf("%+v\n", example)
	// Output:
	// {String:string Int:1 Float:1.2 Map:map[key:value] Slice:[foo bar] Array:[baz qux] Func:0x4ffbb0}
	// {String: Int:0 Float:0 Map:map[] Slice:[] Array:[ ] Func:<nil>}
}

func ExampleSpelunker_Spelunk() {
	type Pet struct {
		Name string `spelunk:"capitalize"`
	}
	person := struct {
		Name    string `spelunk:"trim,capitalize"`
		Age     int    `spelunk2:"min:18"`
		Pets    []Pet
		Secrets map[string]interface{} `spelunk:"secret"`
	}{
		Name: "   gohpher  ",
		Age:  3,
		Pets: []Pet{
			{
				Name: "geomyidae",
			},
		},
		Secrets: map[string]interface{}{
			"foo": "bar",
		},
	}

	fmt.Printf("%+v\n", person)

	s1 := spelunk.New()
	s1.SetHandler("trim", func(name, path, tagValue string, value reflect.Value) error {
		if !value.CanSet() || value.Kind() != reflect.String {
			return errors.New("expected a settable string")
		}
		value.SetString(strings.TrimSpace(value.String()))
		return nil
	})
	s1.SetHandler("capitalize", func(name, path, tagValue string, value reflect.Value) error {
		if !value.CanSet() || value.Kind() != reflect.String {
			return errors.New("expected a settable string")
		}
		fmt.Printf("Capitalizing %s (%s)\n", name, path)
		value.SetString(strings.ToUpper(value.String()[:1]) + value.String()[1:])
		return nil
	})
	s1.SetHandler("secret", spelunk.Zeroer)
	s2 := spelunk.New()
	s2.SetTag("spelunk2")
	s2.SetHandler("min", func(name, path, tagValue string, value reflect.Value) error {
		if !value.CanSet() || value.Kind() != reflect.Int {
			return errors.New("expected a settable int")
		}
		min, err := strconv.Atoi(strings.Split(tagValue, ":")[1])
		if err != nil {
			return errors.New("expected min value to be a int")
		}
		if v := value.Int(); v < int64(min) {
			value.SetInt(int64(min))
		}
		return nil
	})
	s1.Spelunk(&person)
	s2.Spelunk(&person)

	fmt.Printf("%+v\n", person)
	// Output:
	// {Name:   gohpher   Age:3 Pets:[{Name:geomyidae}] Secrets:map[foo:bar]}
	// Capitalizing Name (Name)
	// Capitalizing Name (Pets[0].Name)
	// {Name:Gohpher Age:18 Pets:[{Name:Geomyidae}] Secrets:map[]}
}

func TestSpelunker_Spelunk(t *testing.T) {
	type StructWithString struct {
		String string `tag:"string"`
	}

	tests := map[string]struct {
		input                    interface{}
		want                     interface{}
		err                      error
		stringHandlerInvokes     int
		everyFieldHandlerInvokes int
	}{
		// Special cases.
		"nil": {
			input:                nil,
			want:                 nil,
			err:                  nil,
			stringHandlerInvokes: 0,
		},
		"struct-interface-nil": {
			input: &struct {
				Interface interface{}
			}{},
			want: &struct {
				Interface interface{}
			}{},
			stringHandlerInvokes:     0,
			everyFieldHandlerInvokes: 1,
		},
		"struct-every-field-handler-error": {
			input: struct {
				EveryFieldHandlerError interface{}
			}{},
			want: struct {
				EveryFieldHandlerError interface{}
			}{},
			err:                      errors.New("EveryFieldHandlerError"),
			stringHandlerInvokes:     0,
			everyFieldHandlerInvokes: 1,
		},
		// Input struct is a pointer.
		"struct-ptr": {
			input:                    &StructWithString{String: "string"},
			want:                     &StructWithString{String: "STRING"},
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 1,
		},
		"struct-ptr-struct": {
			input: &struct {
				Struct StructWithString
			}{Struct: StructWithString{String: "string"}},
			want: &struct {
				Struct StructWithString
			}{Struct: StructWithString{String: "STRING"}},
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 2,
		},
		"struct-ptr-slice-struct": {
			input: &struct {
				Strings []StructWithString
			}{Strings: []StructWithString{{String: "string"}}},
			want: &struct {
				Strings []StructWithString
			}{Strings: []StructWithString{{String: "STRING"}}},
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 2,
		},
		"struct-ptr-array-struct": {
			input: &struct {
				Strings [1]StructWithString
			}{Strings: [1]StructWithString{{String: "string"}}},
			want: &struct {
				Strings [1]StructWithString
			}{Strings: [1]StructWithString{{String: "STRING"}}},
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 2,
		},
		"struct-ptr-map-struct": {
			input: &struct {
				Strings map[string]StructWithString
			}{Strings: map[string]StructWithString{"key": {String: "string"}}},
			want: &struct {
				Strings map[string]StructWithString
			}{Strings: map[string]StructWithString{"key": {String: "string"}}},
			err:                      errors.New("cannot set"), // Map value still can't be modified.
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 1,
		},
		"struct-ptr-map-struct-ptr": {
			input: &struct {
				Strings map[string]*StructWithString
			}{Strings: map[string]*StructWithString{"key": {String: "string"}}},
			want: &struct {
				Strings map[string]*StructWithString
			}{Strings: map[string]*StructWithString{"key": {String: "STRING"}}},
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 2,
		},
		"struct-ptr-interface-struct": {
			input: &struct {
				Interface interface{}
			}{Interface: StructWithString{String: "string"}},
			want: &struct {
				Interface interface{}
			}{Interface: StructWithString{String: "string"}},
			err:                      errors.New("cannot set"), // Interface value still can't be modified.
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 1,
		},
		"struct-ptr-interface-struct-ptr": {
			input: &struct {
				Interface interface{}
			}{Interface: &StructWithString{String: "string"}},
			want: &struct {
				Interface interface{}
			}{Interface: &StructWithString{String: "STRING"}},
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 2,
		},
		// Input struct is not a pointer.
		"struct": {
			input:                    StructWithString{String: "string"},
			want:                     StructWithString{String: "string"},
			err:                      errors.New("cannot set"),
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 1,
		},
		"struct-struct": {
			input: struct {
				Struct StructWithString
			}{Struct: StructWithString{String: "string"}},
			want: struct {
				Struct StructWithString
			}{Struct: StructWithString{String: "string"}},
			err:                      errors.New("cannot set"),
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 1,
		},
		"struct-slice-struct": {
			input: struct {
				Strings []StructWithString
			}{Strings: []StructWithString{{String: "string"}}},
			want: struct {
				Strings []StructWithString
			}{Strings: []StructWithString{{String: "STRING"}}}, // Slice is indeed modified.
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 2,
		},
		"struct-slice-struct-ptr": {
			input: struct {
				Strings []*StructWithString
			}{Strings: []*StructWithString{{String: "string"}}},
			want: struct {
				Strings []*StructWithString
			}{Strings: []*StructWithString{{String: "STRING"}}},
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 2,
		},
		"struct-array-struct": {
			input: struct {
				Strings [1]StructWithString
			}{Strings: [1]StructWithString{{String: "string"}}},
			want: struct {
				Strings [1]StructWithString
			}{Strings: [1]StructWithString{{String: "string"}}},
			err:                      errors.New("cannot set"),
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 1,
		},
		"struct-array-struct-ptr": {
			input: struct {
				Strings [1]*StructWithString
			}{Strings: [1]*StructWithString{{String: "string"}}},
			want: struct {
				Strings [1]*StructWithString
			}{Strings: [1]*StructWithString{{String: "STRING"}}},
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 2,
		},
		"struct-map-struct": {
			input: struct {
				Strings map[string]StructWithString
			}{Strings: map[string]StructWithString{"key": {String: "string"}}},
			want: struct {
				Strings map[string]StructWithString
			}{Strings: map[string]StructWithString{"key": {String: "string"}}},
			err:                      errors.New("cannot set"),
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 1,
		},
		"struct-map-struct-ptr": {
			input: struct {
				Strings map[string]*StructWithString
			}{Strings: map[string]*StructWithString{"key": {String: "string"}}},
			want: struct {
				Strings map[string]*StructWithString
			}{Strings: map[string]*StructWithString{"key": {String: "STRING"}}},
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 2,
		},
		"struct-interface-struct": {
			input: struct {
				Interface interface{}
			}{Interface: StructWithString{String: "string"}},
			want: struct {
				Interface interface{}
			}{Interface: StructWithString{String: "string"}},
			err:                      errors.New("cannot set"),
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 1,
		},
		"struct-interface-struct-ptr": {
			input: struct {
				Interface interface{}
			}{Interface: &StructWithString{String: "string"}},
			want: struct {
				Interface interface{}
			}{Interface: &StructWithString{String: "STRING"}},
			stringHandlerInvokes:     1,
			everyFieldHandlerInvokes: 2,
		},
	}

	for key, test := range tests {
		key, test := key, test
		t.Run(key, func(t *testing.T) {
			t.Parallel()

			// OnlyRun(t, key, "struct-every-field-handler-error")

			p := spelunk.New()
			p.SetTag("tag")
			stringHandlerInvokes := 0
			p.SetHandler("string", func(name, path, tagVal string, f reflect.Value) error {
				fmt.Println(path)
				stringHandlerInvokes++
				if f.Kind() == reflect.Ptr && !f.IsNil() {
					f = f.Elem()
				}
				if f.CanSet() == false {
					return errors.New("cannot set")
				}
				f.Set(reflect.ValueOf(strings.ToUpper(f.String())))
				return nil
			})
			everyFieldHandlerInvokes := 0
			p.SetEveryFieldHandler(func(name, path, tagVal string, f reflect.Value) error {
				everyFieldHandlerInvokes++
				if name == "EveryFieldHandlerError" {
					return errors.New("EveryFieldHandlerError")
				}
				return nil
			})

			if err := ErrorEqual(p.Spelunk(test.input), test.err); err != nil {
				t.Errorf("error mismatch: %v", err)
			}

			if !reflect.DeepEqual(test.input, test.want) {
				t.Errorf("transformed input mismatch: got: %#v, want: %#v", test.input, test.want)
			}

			if stringHandlerInvokes != test.stringHandlerInvokes {
				t.Errorf("string handler invoke count mismatch: got: %d, want: %d", stringHandlerInvokes, test.stringHandlerInvokes)
			}

			if everyFieldHandlerInvokes != test.everyFieldHandlerInvokes {
				t.Errorf("every field handler invoke count mismatch: got: %d, want: %d", everyFieldHandlerInvokes, test.everyFieldHandlerInvokes)
			}
		})
	}
}

func TestZeroer(t *testing.T) {
	tests := map[string]struct {
		input interface{}
		want  interface{}
		err   error
	}{
		"nil-ptr": {
			input: func() *string { return nil }(),
			want:  func() *string { return nil }(),
			err:   errors.New("cannot set"),
		},
		"string": {
			input: "string",
			want:  "string",
			err:   errors.New("cannot set"),
		},
		"string-ptr": {
			input: func() *string { s := "string"; return &s }(),
			want:  func() *string { s := ""; return &s }(),
			err:   nil,
		},
		"slice": {
			input: []string{"a"},
			want:  []string{"a"},
			err:   errors.New("cannot set"),
		},
		"slice-ptr": {
			input: func() *[]string { s := []string{"a"}; return &s }(),
			want:  func() *[]string { var s []string; return &s }(),
		},
		"array": {
			input: [1]string{"a"},
			want:  [1]string{"a"},
			err:   errors.New("cannot set"),
		},
		"array-ptr": {
			input: func() *[1]string { s := [1]string{"a"}; return &s }(),
			want:  func() *[1]string { var s [1]string; return &s }(),
		},
		"interface": {
			input: 3,
			want:  3,
			err:   errors.New("cannot set"),
		},
		"interface-to-ptr": {
			input: func() interface{} { s := "string"; return &s }(),
			want:  func() interface{} { s := ""; return &s }(),
		},
		"map": {
			input: map[string]string{"key": "value"},
			want:  map[string]string{"key": "value"},
			err:   errors.New("cannot set"),
		},
		"map-ptr": {
			input: func() *map[string]string { m := map[string]string{"key": "value"}; return &m }(),
			want:  func() *map[string]string { var m map[string]string; return &m }(),
		},
	}

	for key, test := range tests {
		key, test := key, test
		t.Run(key, func(t *testing.T) {
			t.Parallel()

			if err := ErrorEqual(spelunk.Zeroer("", "", "", reflect.ValueOf(test.input)), test.err); err != nil {
				t.Errorf("error mismatch: %v", err)
			}

			if !reflect.DeepEqual(test.input, test.want) {
				t.Errorf("zeroed input mismatch: got: %#v, want: %#v", test.input, test.want)
			}

		})
	}
}

func ErrorEqual(got, want error) error {
	if got == nil {
		if want != nil {
			return errors.New("got error is nil; want error is '" + want.Error() + "'")
		}
		return nil
	}
	if want == nil {
		return errors.New("got error is '" + got.Error() + "'; want error is nil")
	}
	if got.Error() != want.Error() {
		return errors.New("got error is '" + got.Error() + "'; want error is '" + want.Error() + "'")
	}
	return nil
}
