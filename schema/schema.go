package schema

// Decl describes an interface for a declaration.
type Decl interface {
	Pos() Pos
}

// Schema holds all the data for an mpack schema. The order
// of the constants, enums, structs, and unions depends on
// the parser.
type Schema struct {
	Doc     []string
	Package string
	Decls   []Decl
}

// Const holds the data of an mpack constant.
type Const struct {
	pos   Pos
	Doc   []string
	Name  string
	Type  Type
	Value string
}

// Pos implements the Decl interface.
func (c *Const) Pos() Pos { return c.pos }

// Enumerator holds the data of an enumerator value.
type Enumerator struct {
	Name  string
	Value int64
	Tags  Tags
}

// Enum holds the data of an mpack enumeration.
type Enum struct {
	pos         Pos
	Doc         []string
	Name        string
	Enumerators []Enumerator
}

// Pos implements the Decl interface.
func (e *Enum) Pos() Pos       { return e.pos }
func (e *Enum) typeid() string { return e.Name }

// Field holds the data of a struct field.
type Field struct {
	Name    string
	Type    Type
	Ordinal int64
	Tags    Tags
}

// Struct holds the data of an mpack struct.
type Struct struct {
	pos    Pos
	Doc    []string
	Name   string
	Fields []Field
}

// Pos implements the Decl interface.
func (s *Struct) Pos() Pos       { return s.pos }
func (s *Struct) typeid() string { return s.Name }

// Branch holds the data for a union branch.
type Branch struct {
	Type    Type
	Ordinal int64
	Tags    Tags
}

// Union holds the data of an mpack union.
type Union struct {
	pos      Pos
	Doc      []string
	Name     string
	Branches []Branch
}

// Pos implements the Decl interface.
func (u *Union) Pos() Pos       { return u.pos }
func (u *Union) typeid() string { return u.Name }
