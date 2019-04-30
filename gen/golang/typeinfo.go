package golang

import (
	"strconv"
	"strings"

	"github.com/mprot/mprotc/gen"
	"github.com/mprot/mprotc/schema"
)

type typeinfo struct {
	importNames map[string]string
}

func newTypeinfo(importNames map[string]string) *typeinfo {
	return &typeinfo{importNames}
}

func (ti *typeinfo) typename(t schema.Type) string {
	switch t := t.(type) {
	case *schema.Bytes:
		return "[]byte"

	case *schema.Array:
		size := ""
		if t.Size > 0 {
			size = strconv.FormatInt(int64(t.Size), 10)
		}
		return "[" + size + "]" + ti.typename(t.Value)

	case *schema.Map:
		return "map[" + ti.typename(t.Key) + "]" + ti.typename(t.Value)

	case *schema.Raw:
		return "msgpack.Raw"

	case *schema.Time:
		return "time.Time"

	case *schema.Pointer:
		return "*" + ti.typename(t.Value)

	case *schema.DefinedType:
		if !t.Imported() {
			return t.Name()
		}

		name := t.Name()
		if imp := t.Decl.(*schema.Import); imp != nil {
			name = strings.TrimPrefix(name, imp.Name+".")
			if impName, has := ti.importNames[imp.Name]; has {
				name = impName + "." + name
			}
		}
		return name

	default:
		return t.Name()
	}
}

func (ti *typeinfo) typeid(t schema.Type) string {
	switch t := t.(type) {
	case *schema.Int:
		return "int"

	case *schema.Float:
		return "float"

	case *schema.Array:
		return "array"

	case *schema.Map:
		return "map"

	case *schema.Pointer:
		return ti.typeid(t.Value)

	case *schema.DefinedType:
		name := t.Name()
		if t.Imported() {
			imp := t.Decl.(*schema.Import)
			name = strings.TrimPrefix(name, imp.Name+".")
		}
		return gen.SnakeCase(name)

	default:
		return gen.SnakeCase(t.Name())
	}
}
