package js

import (
	"fmt"

	"github.com/mprot/mprotc/gen"
	"github.com/mprot/mprotc/schema"
)

func typescriptImports(s *schema.Schema) []string {
	needTypeImport := false
	iterTypes(s, func(t schema.Type) {
		if defined, ok := t.(*schema.DefinedType); ok {
			switch defined.Decl.(type) {
			case *schema.Struct:
				needTypeImport = true
			case *schema.Union:
				needTypeImport = true
			}
		}
	})

	res := make([]string, 0, 1)
	if needTypeImport {
		res = append(res, "Type")
	}
	return res
}

func typescriptTypename(t schema.Type) string {
	switch t := t.(type) {
	case *schema.Int:
		return "number"
	case *schema.Float:
		return "number"
	case *schema.String:
		return "string"
	case *schema.Array:
		return typescriptTypename(t.Value) + "[]"
	case *schema.Map:
		return fmt.Sprintf("{[key: %s]: %s}", typescriptTypename(t.Key), typescriptTypename(t.Value))
	case *schema.Pointer:
		return msgpackTypename(t.Value)
	case *schema.DefinedType:
		return t.Name()
	default:
		return gen.TitleFirstWord(t.Name())
	}
}
