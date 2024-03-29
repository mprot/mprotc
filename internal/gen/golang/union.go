package golang

import (
	"github.com/mprot/mprotc/internal/gen"
	"github.com/mprot/mprotc/internal/schema"
)

type unionGenerator struct {
	typeid bool
}

func (g *unionGenerator) Generate(p gen.Printer, u *schema.Union, ti *typeinfo) {
	g.printDecl(p, u.Name, u.Branches, u.Doc, ti)
	p.Println()
	g.printEncodeFunc(p, u.Name, u.Branches, ti)
	p.Println()
	g.printDecodeFunc(p, u.Name, u.Branches, ti)
	if g.typeid {
		p.Println()
		g.printTypeidFunc(p, u.Name, u.Branches, ti)
	}
}

func (g *unionGenerator) printDecl(p gen.Printer, name string, branches []schema.Branch, doc []string, ti *typeinfo) {
	branchTypes := make([]schema.Type, 0, len(branches))
	for _, b := range branches {
		branchTypes = append(branchTypes, b.Type)
	}

	var typenames string
	switch n := len(branchTypes); n {
	case 0:
	case 1:
		typenames = ti.typename(branchTypes[0])
	case 2:
		typenames = ti.typename(branchTypes[0]) + " or " + ti.typename(branchTypes[1])
	default:
		for i := 0; i < n-1; i++ {
			typenames += ti.typename(branchTypes[i]) + ", "
		}
		typenames += "or " + ti.typename(branchTypes[n-1])
	}

	printDoc(p, doc, name+" union.")
	p.Println(`type `, name, ` struct {`)
	p.Println(`	Value interface{} // `, typenames)
	p.Println(`}`)
}

func (g *unionGenerator) printEncodeFunc(p gen.Printer, name string, branches []schema.Branch, ti *typeinfo) {
	p.Println(`// EncodeMsgpack implements the Encoder interface for `, name, `.`)
	p.Println(`func (o `, name, `) EncodeMsgpack(w *msgpack.Writer) (err error) {`)
	p.Println(`	if err = w.WriteArrayHeader(2); err != nil {`)
	p.Println(`		return err`)
	p.Println(`	}`)
	p.Println(`	switch v := o.Value.(type) {`)

	for _, b := range branches {
		p.Println(`	case `, ti.typename(b.Type), `:`)
		p.Println(`		if err = w.WriteInt64(`, b.Ordinal, `); err != nil {`)
		p.Println(`			return err`)
		p.Println(`		}`)
		cp := newCodecFuncPrinter("v", b.Type, "")
		cp.printEncode(gen.PrefixedPrinter(p, "\t\t"))
	}

	p.Println(`	default:`)
	p.Println(`		return fmt.Errorf("invalid `, name, ` type %T", o.Value)`)
	p.Println(`	}`)
	p.Println(`	return nil`)
	p.Println(`}`)
}

func (g *unionGenerator) printDecodeFunc(p gen.Printer, name string, branches []schema.Branch, ti *typeinfo) {
	p.Println(`// DecodeMsgpack implements the Decoder interface for `, name, `.`)
	p.Println(`func (o *`, name, `) DecodeMsgpack(r *msgpack.Reader) error {`)
	p.Println(`	if err := r.ReadArrayHeaderWithSize(2); err != nil {`)
	p.Println(`		return err`)
	p.Println(`	}`)
	p.Println(`	ord, err := r.ReadInt64()`)
	p.Println(`	if err != nil {`)
	p.Println(`		return err`)
	p.Println(`	}`)
	p.Println(`	switch ord {`)

	for _, b := range branches {
		typ := ti.typename(b.Type)
		p.Println(`	case `, b.Ordinal, `: // `, typ)
		p.Println(`		var v `, typ)
		cp := newCodecFuncPrinter("v", b.Type, "")
		cp.printDecode(gen.PrefixedPrinter(p, "\t\t"), ti, false)
		p.Println(`		o.Value = v`)
	}

	p.Println(`	default:`)
	p.Println(`		return fmt.Errorf("invalid ordinal %d for `, name, `", ord)`)
	p.Println(`	}`)
	p.Println(`	return nil`)
	p.Println(`}`)
}

func (g *unionGenerator) printTypeidFunc(p gen.Printer, name string, branches []schema.Branch, ti *typeinfo) {
	p.Println(`// TypeID returns the type id for the underlying value of `, name, `.`)
	p.Println(`func (o *`, name, `) TypeID() string {`)
	p.Println(`	switch o.Value.(type) {`)

	for _, b := range branches {
		typ := ti.typename(b.Type)
		id := ti.typeid(b.Type)
		p.Println(`	case `, typ, `:`)
		p.Println(`		return "`, id, `"`)
	}

	p.Println(`	default:`)
	p.Println(`		return ""`)
	p.Println(`	}`)
	p.Println(`}`)
}
