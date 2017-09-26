package js

import (
	"github.com/tsne/mpackc/gen"
	"github.com/tsne/mpackc/schema"
)

func msgpackType(t schema.Type) string {
	switch t := t.(type) {
	case *schema.Int:
		if t.Unsigned {
			return "Uint"
		}
		return "Int"
	case *schema.Float:
		return "Float"
	case *schema.String:
		return "Str"
	case *schema.Array:
		return "_" + msgpackType(t.Value) + "Arr"
	case *schema.Map:
		return "_" + msgpackType(t.Key) + msgpackType(t.Value) + "Map"
	case *schema.Pointer:
		return msgpackType(t.Value)
	case *schema.DefinedType:
		return t.Name()
	default:
		return gen.TitleFirstWord(t.Name())
	}
}

func msgpackImport(t schema.Type) string {
	switch t.(type) {
	case *schema.DefinedType:
		return ""
	case *schema.Array:
		return "TypedArr"
	case *schema.Map:
		return "TypedMap"
	default:
		return msgpackType(t)
	}
}

func iterTypes(s *schema.Schema, f func(schema.Type)) {
	for _, decl := range s.Decls {
		f(schema.DeclType(decl))

		switch decl := decl.(type) {
		case *schema.Struct:
			for _, field := range decl.Fields {
				f(field.Type)
			}
		case *schema.Union:
			for _, branch := range decl.Branches {
				f(branch.Type)
			}
		}
	}
}
