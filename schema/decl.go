package schema

// Decl describes an interface for a declaration.
type Decl interface {
	Pos() Pos
	validate(r errorReporter)
}

// Package holds the data of an mprot package declaration.
type Package struct {
	pos  Pos
	Name string
}

// Pos implements the Decl interface.
func (p *Package) Pos() Pos {
	return p.pos
}

func (p *Package) validate(r errorReporter) {
	// do nothing
}

// Import holds information about an mprot import declaration.
type Import struct {
	pos  Pos
	Path string
	Name string
}

// Pos implements the Decl interface.
func (i *Import) Pos() Pos {
	return i.pos
}

func (i *Import) validate(r errorReporter) {
	// do nothing
}

// Const holds the data of an mprot constant.
type Const struct {
	pos   Pos
	Doc   []string
	Name  string
	Type  Type
	Value string
}

// Pos implements the Decl interface.
func (c *Const) Pos() Pos {
	return c.pos
}

func (c *Const) validate(r errorReporter) {
	// do nothing
}

// Enumerator holds the data of an enumerator value.
type Enumerator struct {
	Name  string
	Value int64
	Tags  Tags
}

// Enum holds the data of an mprot enumeration.
type Enum struct {
	pos         Pos
	Doc         []string
	Name        string
	Enumerators []Enumerator
}

// Pos implements the Decl interface.
func (e *Enum) Pos() Pos {
	return e.pos
}

func (e *Enum) validate(r errorReporter) {
	enumerators := make(map[string]struct{})
	for _, en := range e.Enumerators {
		if _, has := enumerators[en.Name]; has {
			r.errorf("duplicate enumerator %s in enum %s", en.Name, e.Name)
		}
		enumerators[en.Name] = struct{}{}
	}
}

// Field holds the data of a struct field.
type Field struct {
	Name    string
	Type    Type
	Ordinal int64
	Tags    Tags
}

// Struct holds the data of an mprot struct.
type Struct struct {
	pos    Pos
	Doc    []string
	Name   string
	Fields []Field
}

// Pos implements the Decl interface.
func (s *Struct) Pos() Pos {
	return s.pos
}

func (s *Struct) validate(r errorReporter) {
	fields := make(map[string]struct{})
	ordinals := make(map[int64]struct{})
	for _, f := range s.Fields {
		if _, has := fields[f.Name]; has {
			r.errorf("duplicate field %s in struct %s", f.Name, s.Name)
		} else if _, has := ordinals[f.Ordinal]; has && f.Ordinal != 0 {
			r.errorf("duplicate ordinal %d for field %s in struct %s", f.Ordinal, f.Name, s.Name)
		}

		fields[f.Name] = struct{}{}
		ordinals[f.Ordinal] = struct{}{}
	}
}

// Branch holds the data for a union branch.
type Branch struct {
	Type    Type
	Ordinal int64
	Tags    Tags
}

// Union holds the data of an mprot union.
type Union struct {
	pos      Pos
	Doc      []string
	Name     string
	Branches []Branch
}

// Pos implements the Decl interface.
func (u *Union) Pos() Pos {
	return u.pos
}

func (u *Union) validate(r errorReporter) {
	if len(u.Branches) == 0 {
		r.errorf("union %s does not contain a branch", u.Name)
		return
	}

	branches := make(map[string]struct{})
	ordinals := make(map[int64]struct{})
	hasNumericBranch := false
	for _, b := range u.Branches {
		typeid := b.Type.typeid()
		switch typ := b.Type.(type) {
		case *Pointer:
			r.errorf("pointer branch %s in union %s", typ.Name(), u.Name)
		case *Int:
			if hasNumericBranch {
				r.errorf("duplicate numeric branch %s in union %s", typ.Name(), u.Name)
			}
			hasNumericBranch = true
		case *Float:
			if hasNumericBranch {
				r.errorf("duplicate numeric branch %s in union %s", typ.Name(), u.Name)
			}
			hasNumericBranch = true
		case *Raw:
			r.errorf("raw branch in union %s", u.Name)
		case *DefinedType:
			if _, has := branches[typeid]; has {
				r.errorf("duplicate branch %s in union %s", typeid, u.Name)
			} else {
				switch typ.Decl.(type) {
				case *Enum:
					if hasNumericBranch {
						r.errorf("duplicate numeric branch %s in union %s", typ.Name(), u.Name)
					}
					hasNumericBranch = true
				case *Union:
					r.errorf("union branch %s in union %s", typ.Name(), u.Name)
					continue
				}
			}
		default:
			if _, has := branches[typeid]; has {
				r.errorf("duplicate branch %s in union %s (only one %s branch is allowed)", typ.Name(), u.Name, typeid)
			}
		}

		if _, has := ordinals[b.Ordinal]; has && b.Ordinal != 0 {
			r.errorf("duplicate ordinal %d for branch %s in union %s", b.Ordinal, b.Type.Name(), u.Name)
		}

		branches[typeid] = struct{}{}
		ordinals[b.Ordinal] = struct{}{}
	}
}
