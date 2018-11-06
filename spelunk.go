package spelunk

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// A Handler operates on struct field values and is called by the Spelunker for each appropriately-tagged field.
// The name is the field name, the path the field path, and the tagValue the value of the tag.
// Returning a non-nil error immediately causes the Spelunker's Spelunk method to stop and return the error.
type Handler func(name, path, tagValue string, value reflect.Value) error

// A Spelunker recursively explores into a struct looking at all found fields for a specific tag.
// Once a tag is found it's value is used to call registered handlers.
// Tag values should be comma-separated.
// To pass additional values to a handler the tag value is further split by ":".
// The left-hand side of the split value is handler key.
// For example: a tag value of "foo:3,bar" would cause calls to the "foo" and "bar" handlers.
// The EveryFieldHandler is called, if set, for each found field.
type Spelunker struct {
	Tag               string
	Handlers          map[string]Handler
	EveryFieldHandler Handler
}

// New returns a Spelunker with the Tag set to the default value of "spelunk".
func New() *Spelunker {
	return &Spelunker{
		Tag: "spelunk",
	}
}

// SetTag sets the tag for which the Spelunker looks.
func (p *Spelunker) SetTag(tag string) *Spelunker {
	p.Tag = tag
	return p
}

// SetHandler sets the handler to be called when a specific key is found in the tag value.
func (p *Spelunker) SetHandler(key string, h Handler) *Spelunker {
	if p.Handlers == nil {
		p.Handlers = make(map[string]Handler)
	}
	p.Handlers[key] = h
	return p
}

// SetEveryFieldHandler sets the handler to be called for each found field.
func (p *Spelunker) SetEveryFieldHandler(h Handler) *Spelunker {
	p.EveryFieldHandler = h
	return p
}

// Spelunk begins the process of recursively exploring the specified struct.
func (p *Spelunker) Spelunk(i interface{}) error {
	return p.spelunk("", reflect.ValueOf(i))
}

func (p *Spelunker) spelunk(name string, v reflect.Value) error {
	if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && !v.IsNil() {
		return p.spelunk(name, v.Elem())
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		path := fieldType.Name
		if name != "" {
			path = strings.Join([]string{name, path}, ".")
		}

		switch field.Kind() {
		case reflect.Struct:
			if err := p.spelunk(path, field); err != nil {
				return err
			}
		case reflect.Slice, reflect.Array:
			for i := 0; i < field.Len(); i++ {
				if err := p.spelunk(path+fmt.Sprintf("[%d]", i), field.Index(i)); err != nil {
					return err
				}
			}
		case reflect.Map:
			for _, key := range field.MapKeys() {
				if err := p.spelunk(path+fmt.Sprintf("[\"%s\"]", key), field.MapIndex(key)); err != nil {
					return err
				}
			}
		case reflect.Interface:
			if field.IsNil() {
				break
			}

			if err := p.spelunk(path, field.Elem()); err != nil {
				return err
			}
		}

		for _, m := range strings.Split(fieldType.Tag.Get(p.Tag), ",") {
			m := strings.TrimSpace(m)
			parts := strings.SplitN(m, ":", 2)

			if p.EveryFieldHandler != nil {
				if err := p.EveryFieldHandler(fieldType.Name, path, m, field); err != nil {
					return err
				}
			}

			if modifier, ok := p.Handlers[parts[0]]; ok {
				if err := modifier(fieldType.Name, path, m, field); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

var _ Handler = Zeroer

// ErrCannotSet is returned by the Zeroer handler if a value cannot be zeroed.
var ErrCannotSet = errors.New("cannot set")

// Zeroer is a simple Handler which sets the value to its zero value.
// If the value cannot be set a ErrCannotSet error is returned.
func Zeroer(_, _, _ string, f reflect.Value) error {
	if f.Kind() == reflect.Ptr && !f.IsNil() {
		f = f.Elem()
	}
	if f.CanSet() == false {
		return ErrCannotSet
	}
	f.Set(reflect.Zero(f.Type()))
	return nil
}
