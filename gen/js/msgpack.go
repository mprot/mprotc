package js

import (
	"sort"

	"github.com/tsne/mpackc/gen"
	"github.com/tsne/mpackc/schema"
)

func msgpackImports(s *schema.Schema) []string {
	imports := map[string]struct{}{}
	iterTypes(s, func(t schema.Type) {
		addMsgpackImports(imports, t)
	})

	res := make([]string, 0, len(imports))
	for imp := range imports {
		res = append(res, imp)
	}
	sort.Strings(res)
	return res
}

func addMsgpackImports(imports map[string]struct{}, t schema.Type) {
	switch t := t.(type) {
	case *schema.Array:
		imports["TypedArr"] = struct{}{}

	case *schema.Map:
		imports["TypedMap"] = struct{}{}

	case *schema.DefinedType:
		switch decl := t.Decl.(type) {
		case *schema.Enum:
			imports["Int"] = struct{}{}
		case *schema.Struct:
			imports["Int"] = struct{}{} // field ordinal
			imports["Map"] = struct{}{} // map header
			imports["Any"] = struct{}{} // skip value in decoding
			for _, f := range decl.Fields {
				addMsgpackImports(imports, f.Type)
			}
		case *schema.Union:
			imports["Int"] = struct{}{} // branch ordinal
			imports["Arr"] = struct{}{} // array header
			for _, b := range decl.Branches {
				addMsgpackImports(imports, b.Type)
			}
		}

	default:
		imports[msgpackTypename(t)] = struct{}{}
	}
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
