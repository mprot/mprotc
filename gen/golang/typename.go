package golang

import (
	"strconv"
	"strings"

	"github.com/mprot/mprotc/schema"
)

type typenameFunc func(schema.Type) string

func newTypenameFunc(importNames map[string]string) typenameFunc {
	var typename typenameFunc
	typename = func(t schema.Type) string {
		switch t := t.(type) {
		case *schema.Bytes:
			return "[]byte"

		case *schema.Array:
			size := ""
			if t.Size > 0 {
				size = strconv.FormatInt(int64(t.Size), 10)
			}
			return "[" + size + "]" + typename(t.Value)

		case *schema.Map:
			return "map[" + typename(t.Key) + "]" + typename(t.Value)

		case *schema.Raw:
			return "msgpack.Raw"

		case *schema.Time:
			return "time.Time"

		case *schema.Pointer:
			return "*" + typename(t.Value)

		case *schema.DefinedType:
			if t.Imported() {
				imp := t.Decl.(*schema.Import)
				name := strings.TrimPrefix(t.Name(), imp.Name+".")
				if impName, has := importNames[imp.Name]; has {
					name = impName + "." + name
				}
				return name
			}
		}
		return t.Name()
	}
	return typename
}
