package js

import (
	"sort"

	"github.com/tsne/mpackc/gen"
	"github.com/tsne/mpackc/schema"
)

func msgpackImports(s *schema.Schema) []string {
	imports := map[string]struct{}{}
	iterTypes(s, func(t schema.Type) {
		if defined, ok := t.(*schema.DefinedType); ok {
			if _, isEnum := defined.Decl.(*schema.Enum); isEnum {
				imports["Int"] = struct{}{}
			}
		} else {
			switch t.(type) {
			case *schema.DefinedType:
				// skip
			case *schema.Array:
				imports["TypedArr"] = struct{}{}
			case *schema.Map:
				imports["TypedMap"] = struct{}{}
			default:
				imports[msgpackTypename(t)] = struct{}{}
			}
		}
	})

	res := make([]string, 0, len(imports))
	for imp := range imports {
		res = append(res, imp)
	}
	sort.Strings(res)
	return res
}

func msgpackTypename(t schema.Type) string {
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
		return "_" + msgpackTypename(t.Value) + "Arr"
	case *schema.Map:
		return "_" + msgpackTypename(t.Key) + msgpackTypename(t.Value) + "Map"
	case *schema.Pointer:
		return msgpackTypename(t.Value)
	case *schema.DefinedType:
		return t.Name()
	default:
		return gen.TitleFirstWord(t.Name())
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
