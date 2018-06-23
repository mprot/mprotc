package schema

import (
	"fmt"
	"strconv"
	"strings"
)

// Type defines an interface for the various types supported by the
// mprot schema definition.
type Type interface {
	Name() string
	typeid() string
}

// Bool describes the boolean data type.
type Bool struct{}

// Name implements the Type interface.
func (b *Bool) Name() string {
	return "bool"
}

func (b *Bool) typeid() string {
	return "boolean"
}

// Int describes an integral data type with a specific number of bits.
type Int struct {
	Bits     int
	Unsigned bool
}

// Name implements the Type interface.
func (i *Int) Name() string {
	typ := "int"
	if i.Unsigned {
		typ = "uint"
	}
	if i.Bits > 0 {
		typ += strconv.FormatInt(int64(i.Bits), 10)
	}
	return typ
}

func (i *Int) typeid() string {
	return "integer"
}

// Float describes a floating-point data type with a specific number of bits.
type Float struct {
	Bits int
}

// Name implements the Type interface.
func (f *Float) Name() string {
	const typ = "float"
	if f.Bits <= 0 {
		return typ
	}
	return typ + strconv.FormatInt(int64(f.Bits), 10)
}

func (f *Float) typeid() string {
	return "floating-point"
}

// String describes the data type for strings.
type String struct{}

// Name implements the Type interface.
func (s *String) Name() string {
	return "string"
}

func (s *String) typeid() string {
	return "string"
}

// Bytes describes a data type for binary data.
type Bytes struct{}

// Name implements the Type interface.
func (b *Bytes) Name() string {
	return "bytes"
}

func (b *Bytes) typeid() string {
	return "bytes"
}

// Array describes an array of a specific data type.
type Array struct {
	Size  int
	Value Type
}

// Name implements the Type interface.
func (a *Array) Name() string {
	if a.Size > 0 {
		return fmt.Sprintf("[%d]%s", a.Size, a.Value.Name())
	}
	return "[]" + a.Value.Name()
}

func (a *Array) typeid() string {
	return "array"
}

// Map describes a map type with a specific key and value type.
type Map struct {
	Key   Type
	Value Type
}

// Name implements the Type interface.
func (m *Map) Name() string {
	return "map[" + m.Key.Name() + "]" + m.Value.Name()
}

func (m *Map) typeid() string {
	return "map"
}

// Raw describes an blob type containing an encoded value.
type Raw struct{}

// Name implements the Type interface.
func (r *Raw) Name() string {
	return "raw"
}

func (r *Raw) typeid() string {
	return "raw"
}

// Time describes a time data type.
type Time struct{}

// Name implements the Type interface.
func (t *Time) Name() string {
	return "time"
}

func (t *Time) typeid() string {
	return "time"
}

// Pointer describes an pointer to a specific data type.
type Pointer struct {
	Value Type
}

// Name implements the Type interface.
func (p *Pointer) Name() string {
	return "*" + p.Value.Name()
}

func (p *Pointer) typeid() string {
	return "pointer"
}

// DefinedType describes a user defined type.
type DefinedType struct {
	pkg  string // empty for local
	name string
	Decl Decl
}

// DeclType returns the defined type for the given declaration.
// If the declaration type is not supported, nil will be returned.
func DeclType(decl Decl) *DefinedType {
	var name string
	switch decl := decl.(type) {
	case *Enum:
		name = decl.Name
	case *Struct:
		name = decl.Name
	case *Union:
		name = decl.Name
	default:
		return nil
	}

	return &DefinedType{
		name: name,
		Decl: decl,
	}
}

func newDefinedType(name string, decl Decl) *DefinedType {
	pkg := ""
	if idx := strings.IndexByte(name, '.'); idx >= 0 {
		pkg = name[:idx]
		name = name[idx+1:]
	}

	return &DefinedType{
		pkg:  pkg,
		name: name,
		Decl: decl,
	}
}

// Imported reports if the defined type is imported from
// another file.
func (t *DefinedType) Imported() bool {
	return t.pkg != ""
}

// Name implements the Type interface.
func (t *DefinedType) Name() string {
	if t.pkg == "" {
		return t.name
	}
	return t.pkg + "." + t.name
}

func (t *DefinedType) typeid() string {
	return t.name
}
