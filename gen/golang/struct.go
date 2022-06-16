package golang

import (
	"github.com/mprot/mprotc/gen"
	"github.com/mprot/mprotc/schema"
)

type structGenerator struct {
	unwrapUnion bool
	typeid      bool
}

func (g *structGenerator) Generate(p gen.Printer, s *schema.Struct, ti *typeinfo) {
	g.printDecl(p, s.Name, s.Fields, s.Doc, ti)
	p.Println()
	g.printEncodeFunc(p, s.Name, s.Fields, ti)
	p.Println()
	g.printDecodeFunc(p, s.Name, s.Fields, ti)
	if g.typeid {
		p.Println()
		g.printTypeidFunc(p, s.Name, ti.typeid(schema.DeclType(s)))
	}
}

func (g *structGenerator) printDecl(p gen.Printer, name string, fields []schema.Field, doc []string, ti *typeinfo) {
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
		ftype := ti.typename(f.Type)
		if g.unwrapUnion && isUnion(f.Type) {
			ftype = "interface{} // " + ftype
		}

		p.Println(`	`, fname, ` `, ftype)
	}
	p.Println(`}`)
}

func (g *structGenerator) printEncodeFunc(p gen.Printer, name string, fields []schema.Field, ti *typeinfo) {
	p.Println(`// EncodeMsgpack implements the Encoder interface for `, name, `.`)
	p.Println(`func (o `, name, `) EncodeMsgpack(w *msgpack.Writer) (err error) {`)
	p.Println(`	if err = w.WriteMapHeader(`, len(fields), `); err != nil {`)
	p.Println(`		return err`)
	p.Println(`	}`)
	for _, f := range fields {
		p.Println(`	// `, f.Name)
		p.Println(`	if err = w.WriteInt64(`, f.Ordinal, `); err != nil {`)
		p.Println(`		return err`)
		p.Println(`	}`)
		g.printFieldEncode(gen.PrefixedPrinter(p, "\t"), "o", f, ti)
	}
	p.Println(`	return nil`)
	p.Println(`}`)
}

func (g *structGenerator) printDecodeFunc(p gen.Printer, name string, fields []schema.Field, ti *typeinfo) {
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
		g.printFieldDecode(gen.PrefixedPrinter(p, "\t\t\t"), "o", f, ti)
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

func (g *structGenerator) printTypeidFunc(p gen.Printer, name string, typeid string) {
	p.Println(`// TypeID returns the type id for `, name, `.`)
	p.Println(`func (o *`, name, `) TypeID() string {`)
	p.Println(`	return "`, typeid, `"`)
	p.Println(`}`)
}

func (g *structGenerator) printFieldEncode(p gen.Printer, receiver string, field schema.Field, ti *typeinfo) {
	specifier := receiver + "." + field.Name
	if g.unwrapUnion && isUnion(field.Type) {
		specifier = "(" + ti.typename(field.Type) + "{" + specifier + "})"
	}

	cp := newCodecFuncPrinter(specifier, field.Type, "")
	cp.printEncode(p)
}

func (g *structGenerator) printFieldDecode(p gen.Printer, receiver string, field schema.Field, ti *typeinfo) {
	fieldSpecifier := receiver + "." + field.Name
	if g.unwrapUnion && isUnion(field.Type) {
		p.Println(`var u `, ti.typename(field.Type))
		cp := newCodecFuncPrinter("u", field.Type, "")
		cp.printDecode(p, ti, false)
		p.Println(fieldSpecifier, ` = u.Value`)
	} else {
		cp := newCodecFuncPrinter(fieldSpecifier, field.Type, "")
		cp.printDecode(p, ti, false)
	}
}

func isUnion(t schema.Type) bool {
	dt, ok := t.(*schema.DefinedType)
	if ok {
		_, ok = dt.Decl.(*schema.Union)
	}
	return ok
}
