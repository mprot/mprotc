package js

import (
	"fmt"
	"strings"

	"github.com/tsne/mpackc/gen"
	"github.com/tsne/mpackc/schema"
)

type unionGenerator struct{}

func (g *unionGenerator) Generate(p gen.Printer, u *schema.Union) {
	branches := collectBranches(u)

	printDoc(p, u.Doc, u.Name+" union.")
	p.Println(`export const `, u.Name, ` = {`)
	g.printEncodeFunc(p, branches)
	p.Println()
	g.printDecodeFunc(p, branches)
	p.Println(`};`)
}

func (g *unionGenerator) GenerateTypeDecls(p gen.Printer, u *schema.Union) {
	types := make([]string, 0, len(u.Branches))
	for _, b := range u.Branches {
		types = append(types, typescriptTypename(b.Type))
	}

	p.Println(`export declare var `, u.Name, `: Type<`, u.Name, `>;`)
	p.Println(`export type `, u.Name, ` = `, strings.Join(types, " | "))
}

func (g *unionGenerator) printEncodeFunc(p gen.Printer, branches *branches) {
	p.Println(`	enc(buf, v) {`)
	p.Println(`		Arr.encHeader(buf, 2);`)
	p.Println(`		switch(typeof v) {`)

	if branches.boolean != nil {
		p.Println(`		case "bool":`)
		p.Println(`			Int.enc(buf, `, branches.boolean.Ordinal, `);`)
		p.Println(`			return `, branches.boolean.msgpackType, `.enc(buf, v);`)
	}
	if branches.integer != nil || branches.float != nil {
		number := branches.float
		if number == nil {
			number = branches.integer
		}

		p.Println(`		case "number":`)
		p.Println(`			Int.enc(buf, `, number.Ordinal, `);`)
		p.Println(`			return `, number.msgpackType, `.enc(buf, v);`)
	}
	if branches.str != nil {
		p.Println(`		case "string":`)
		p.Println(`			Int.enc(buf, `, branches.str.Ordinal, `);`)
		p.Println(`			return `, branches.str.msgpackType, `.enc(buf, v);`)
	}
	if branches.mapping != nil || len(branches.objs) != 0 {
		p.Println(`		case "object":`)
		p.Println(`			v = v || {};`)

		var emptyObj *branch
		for _, obj := range branches.objs {
			if obj.typecheck == "" {
				emptyObj = obj
				continue
			}

			typecheck := fmt.Sprintf(obj.typecheck, "v")
			p.Println(`			if(`, typecheck, `) {`)
			p.Println(`				// `, typescriptTypename(obj.Type))
			p.Println(`				Int.enc(buf, `, obj.Ordinal, `);`)
			p.Println(`				return `, obj.msgpackType, `.enc(buf, v);`)
			p.Println(`			}`)
		}

		if branches.mapping != nil {
			p.Println(`			// `, typescriptTypename(branches.mapping.Type))
			p.Println(`			Int.enc(buf, `, branches.mapping.Ordinal, `);`)
			p.Println(`			return `, branches.mapping.msgpackType, `.enc(buf, v);`)
		} else if emptyObj != nil {
			p.Println(`			// `, typescriptTypename(emptyObj.Type))
			p.Println(`			Int.enc(buf, `, emptyObj.Ordinal, `);`)
			p.Println(`			return `, emptyObj.msgpackType, `.enc(buf, v);`)
		}
	}

	p.Println(`		default:`)
	p.Println(`			throw new TypeError("invalid union type");`)
	p.Println(`		}`)
	p.Println(`	},`)
}

func (g *unionGenerator) printDecodeFunc(p gen.Printer, branches *branches) {
	p.Println(`	dec(buf) {`)
	p.Println(`		Arr.decHeader(buf, 2);`)
	p.Println(`		switch(Int.dec(buf)) {`)

	for _, b := range branches.all {
		p.Println(`		case `, b.Ordinal, `:`)
		p.Println(`			return `, b.msgpackType, `.dec(buf);`)
	}

	p.Println(`		default :`)
	p.Println(`			throw new TypeError("invalid union type");`)
	p.Println(`		}`)
	p.Println(`	},`)
}

type branch struct {
	schema.Branch
	msgpackType string
	typecheck   string // format string with specifier as an argument
}

type branches struct {
	all     []branch
	boolean *branch
	integer *branch
	float   *branch
	str     *branch
	mapping *branch
	objs    []*branch
}

func collectBranches(u *schema.Union) *branches {
	res := &branches{
		all:  make([]branch, 0, len(u.Branches)),
		objs: make([]*branch, 0, len(u.Branches)),
	}

	for i := 0; i < len(u.Branches); i++ {
		res.all = append(res.all, branch{
			Branch:      u.Branches[i],
			msgpackType: msgpackTypename(u.Branches[i].Type),
		})
		b := &res.all[i]

		switch typ := b.Type.(type) {
		case *schema.Bool:
			res.boolean = b

		case *schema.Int:
			res.integer = b

		case *schema.Float:
			res.float = b

		case *schema.String:
			res.str = b

		case *schema.Bytes:
			b.typecheck = `%[1]v instanceof Uint8Array || %[1]v instanceof ArrayBuffer`
			res.objs = append(res.objs, b)

		case *schema.Array:
			b.typecheck = `Array.isArray(%v)`
			res.objs = append(res.objs, b)

		case *schema.Map:
			res.mapping = b

		case *schema.Time:
			b.typecheck = `%v instanceof Date`
			res.objs = append(res.objs, b)

		case *schema.DefinedType:
			switch decl := typ.Decl.(type) {
			case *schema.Enum:
				res.integer = b

			case *schema.Struct:
				fieldchecks := make([]string, 0, len(decl.Fields))
				for _, f := range decl.Fields {
					fieldchecks = append(fieldchecks, `"`+fieldName(f)+`" in %[1]v`)
				}
				b.typecheck = strings.Join(fieldchecks, " && ")
				res.objs = append(res.objs, b)

			default:
				panic(fmt.Sprintf("unsupported declaration type %T", decl))
			}

		default:
			panic(fmt.Sprintf("unsupported type %q", typ.Name()))
		}
	}

	return res
}
