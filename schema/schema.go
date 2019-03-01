package schema

import (
	"path/filepath"
)

// Schema defines a whole mprot schema, including all files.
type Schema []*File

// Parse parses an mprot schema, which is defined in the files specified
// by the given glob patterns. The given glob patterns are interpreted as
// relative to the given root directory.
func Parse(rootDir string, globPatterns []string) (Schema, error) {
	fset, err := glob(rootDir, globPatterns)
	if err != nil {
		return nil, err
	}

	s := make(Schema, 0, fset.size())

	var (
		p    parser
		errs ErrorList
	)
	for _, filename := range fset.filenames() {
		f, err := p.ParseFile(filename)
		if err != nil {
			if el, ok := err.(ErrorList); ok {
				errs = errs.concat(el)
			} else {
				return nil, err
			}
		}

		if filename, err := filepath.Rel(rootDir, f.Name); err == nil {
			f.Name = filename
		}
		s = append(s, f)
	}
	errs.sort()
	return s, errs.err()
}

// RemoveDeprecated removes all declarations which are marked as deprecated.
func (s Schema) RemoveDeprecated() {
	for _, file := range s {
		file.RemoveDeprecated()
	}
}

// File holds all the data for an mprot schema definition file.
// The order of the constants, enums, structs, and unions depends
// on the parser.
type File struct {
	Name    string
	Doc     []string
	Package *Package
	Imports map[string]*Import // name => import
	Decls   []Decl
}

// RemoveDeprecated removes all declarations which are marked as deprecated.
func (f *File) RemoveDeprecated() {
	usedImports := make(map[string]struct{})
	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *Enum:
			idx := 0
			for i := 0; i < len(decl.Enumerators); i++ {
				if !decl.Enumerators[i].Tags.Deprecated() {
					decl.Enumerators[idx] = decl.Enumerators[i]
					idx++
				}
			}
			decl.Enumerators = decl.Enumerators[:idx]

		case *Struct:
			idx := 0
			for i := 0; i < len(decl.Fields); i++ {
				if field := decl.Fields[i]; !field.Tags.Deprecated() {
					decl.Fields[idx] = field
					idx++
					if pkg := importName(field.Type); pkg != "" {
						usedImports[pkg] = struct{}{}
					}
				}
			}
			decl.Fields = decl.Fields[:idx]

		case *Union:
			idx := 0
			for i := 0; i < len(decl.Branches); i++ {
				if branch := decl.Branches[i]; !branch.Tags.Deprecated() {
					decl.Branches[idx] = branch
					idx++
					if pkg := importName(branch.Type); pkg != "" {
						usedImports[pkg] = struct{}{}
					}
				}
			}
			decl.Branches = decl.Branches[:idx]

		case *Service:
			idx := 0
			for i := 0; i < len(decl.Methods); i++ {
				if method := decl.Methods[i]; !method.Tags.Deprecated() {
					decl.Methods[idx] = method
					idx++
					for _, arg := range method.Args {
						if pkg := importName(arg); pkg != "" {
							usedImports[pkg] = struct{}{}
						}
					}
					if pkg := importName(method.Return); pkg != "" {
						usedImports[pkg] = struct{}{}
					}
				}
			}
			decl.Methods = decl.Methods[:idx]
		}
	}

	for importName := range f.Imports {
		if _, used := usedImports[importName]; !used {
			delete(f.Imports, importName)
		}
	}
}

func (f *File) validate(r errorReporter) {
	for _, decl := range f.Decls {
		decl.validate(r)
	}
}

func importName(typ Type) string {
	if dt, ok := typ.(*DefinedType); ok && dt.Imported() {
		return dt.pkg
	}
	return ""
}
