package js

import (
	"fmt"

	"github.com/mprot/mprotc/gen"
	"github.com/mprot/mprotc/schema"
)

func typescriptImports(f *schema.File) []string {
	hasStruct := false
	hasUnion := false
	iterTypes(f, func(t schema.Type) {
		if defined, ok := t.(*schema.DefinedType); ok {
			switch defined.Decl.(type) {
			case *schema.Struct:
				hasStruct = true
			case *schema.Union:
				hasUnion = true
			}
		}
	})

	if !hasStruct && !hasUnion {
		return nil
	}
	return []string{"Type"}
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
