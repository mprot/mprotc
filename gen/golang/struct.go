package golang

import (
	"github.com/mprot/mprotc/gen"
	"github.com/mprot/mprotc/schema"
)

type structGenerator struct {
	unwrapUnion bool
}

func (g *structGenerator) Generate(p gen.Printer, s *schema.Struct, types *typegen) {
	g.printDecl(p, s.Name, s.Fields, s.Doc, types)
	p.Println()
	g.printEncodeFunc(p, s.Name, s.Fields, types)
	p.Println()
	g.printDecodeFunc(p, s.Name, s.Fields, types)
}

func (g *structGenerator) printDecl(p gen.Printer, name string, fields []schema.Field, doc []string, types *typegen) {
	var maxNameLen int
	for _, f := range fields {
		if len(f.Name) > maxNameLen {
			maxNameLen = len(f.Name)
		}
	}

	printDoc(p, doc, name+" structure.")
	p.Println(`type `, name, ` struct {`)
	for _, f := range fields {
		fname := gen.RPad(f.Name, maxNameLen)
		ftype := types.typename(f.Type)
		if g.unwrapUnion && isUnion(f.Type) {
			ftype = "interface{} // " + ftype
		}

		p.Println(`	`, fname, ` `, ftype)
	}
	p.Println(`}`)
}

func (g *structGenerator) printEncodeFunc(p gen.Printer, name string, fields []schema.Field, types *typegen) {
	p.Println(`// EncodeMsgpack implements the Encoder interface for `, name, `.`)
	p.Println(`func (o *`, name, `) EncodeMsgpack(w *msgpack.Writer) (err error) {`)
	p.Println(`	if err = w.WriteMapHeader(`, len(fields), `); err != nil {`)
	p.Println(`		return err`)
	p.Println(`	}`)
	for _, f := range fields {
		p.Println(`	// `, f.Name)
		p.Println(`	if err = w.WriteInt64(`, f.Ordinal, `); err != nil {`)
		p.Println(`		return err`)
		p.Println(`	}`)
		g.printFieldEncode(p, "o", f, types, "\t")
	}
	p.Println(`	return nil`)
	p.Println(`}`)
}

func (g *structGenerator) printDecodeFunc(p gen.Printer, name string, fields []schema.Field, types *typegen) {
	p.Println(`// DecodeMsgpack implements the Decoder interface for `, name, `.`)
	p.Println(`func (o *`, name, `) DecodeMsgpack(r *msgpack.Reader) error {`)
	p.Println(`	n, err := r.ReadMapHeader()`)
	p.Println(`	if err != nil {`)
	p.Println(`		return err`)
	p.Println(`	}`)
	p.Println(`	for i := 0; i < n; i++ {`)
	p.Println(`		ord, err := r.ReadInt64()`)
	p.Println(`		if err != nil {`)
	p.Println(`			return err`)
	p.Println(`		}`)
	p.Println(`		switch ord {`)
	for _, f := range fields {
		p.Println(`		case `, f.Ordinal, `: // `, f.Name)
		g.printFieldDecode(p, "o", f, types, "\t\t\t")
	}
	p.Println(`		default:`)
	p.Println(`			if err := r.Skip(); err != nil {`)
	p.Println(`				return err`)
	p.Println(`			}`)
	p.Println(`		}`)
	p.Println(`	}`)
	p.Println(`	return nil`)
	p.Println(`}`)
}

func (g *structGenerator) printFieldEncode(p gen.Printer, receiver string, field schema.Field, types *typegen, indent string) {
	specifier := receiver + "." + field.Name
	if g.unwrapUnion && isUnion(field.Type) {
		specifier = "(" + types.typename(field.Type) + "{" + specifier + "})"
	}
	types.printEncodeCall(p, field.Type, specifier, indent)
}

func (g *structGenerator) printFieldDecode(p gen.Printer, receiver string, field schema.Field, types *typegen, indent string) {
	fieldSpecifier := receiver + "." + field.Name
	if g.unwrapUnion && isUnion(field.Type) {
		p.Println(indent, `var u `, types.typename(field.Type))
		types.printDecodeCall(p, field.Type, "u", indent)
		p.Println(indent, fieldSpecifier, ` = u.Value`)
	} else {
		types.printDecodeCall(p, field.Type, fieldSpecifier, indent)
	}
}
