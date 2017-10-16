package golang

import (
	"github.com/tsne/mpackc/gen"
	"github.com/tsne/mpackc/schema"
)

type unionGenerator struct{}

func (g *unionGenerator) Generate(p gen.Printer, u *schema.Union) {
	g.printDecl(p, u.Name, u.Branches, u.Doc)
	p.Println()
	g.printEncodeFunc(p, u.Name, u.Branches)
	p.Println()
	g.printDecodeFunc(p, u.Name, u.Branches)
}

func (g *unionGenerator) printDecl(p gen.Printer, name string, branches []schema.Branch, doc []string) {
	types := make([]schema.Type, 0, len(branches))
	for _, b := range branches {
		types = append(types, b.Type)
	}

	var typenames string
	switch n := len(types); n {
	case 0:
	case 1:
		typenames = typename(types[0])
	case 2:
		typenames = typename(types[0]) + " or " + typename(types[1])
	default:
		for i := 0; i < n-1; i++ {
			typenames += typename(types[i]) + ", "
		}
		typenames += "or " + typename(types[n-1])
	}

	printDoc(p, doc, name+" union.")
	p.Println(`type `, name, ` struct {`)
	p.Println(`	Value interface{} // `, typenames)
	p.Println(`}`)
}

func (g *unionGenerator) printEncodeFunc(p gen.Printer, name string, branches []schema.Branch) {
	p.Println(`// EncodeMsgpack implements the Encoder interface for `, name, `.`)
	p.Println(`func (o `, name, `) EncodeMsgpack(w *msgpack.Writer) (err error) {`)
	p.Println(`	if err = w.WriteArrayHeader(2); err != nil {`)
	p.Println(`		return err`)
	p.Println(`	}`)
	p.Println(`	switch v := o.Value.(type) {`)

	for _, b := range branches {
		p.Println(`	case `, typename(b.Type), `:`)
		p.Println(`		if err = w.WriteInt64(`, b.Ordinal, `); err != nil {`)
		p.Println(`			return err`)
		p.Println(`		}`)
		printEncodeCall(p, b.Type, "v", "\t\t")
	}

	p.Println(`	default:`)
	p.Println(`		return fmt.Errorf("invalid `, name, ` type %T", o.Value)`)
	p.Println(`	}`)
	p.Println(`	return nil`)
	p.Println(`}`)
}

func (g *unionGenerator) printDecodeFunc(p gen.Printer, name string, branches []schema.Branch) {
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
		typ := typename(b.Type)
		p.Println(`	case `, b.Ordinal, `: // `, typ)
		p.Println(`		var v `, typ)
		printDecodeCall(p, b.Type, "v", "\t\t")
		p.Println(`		o.Value = v`)
	}

	p.Println(`	default:`)
	p.Println(`		return fmt.Errorf("invalid ordinal %d for `, name, `", ord)`)
	p.Println(`	}`)
	p.Println(`	return nil`)
	p.Println(`}`)
}
