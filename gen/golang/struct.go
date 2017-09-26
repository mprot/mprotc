package golang

import (
	"github.com/tsne/mpackc/gen"
	"github.com/tsne/mpackc/schema"
)

type structGenerator struct{}

func (g *structGenerator) Generate(p *gen.Printer, s *schema.Struct) {
	g.printDecl(p, s.Name, s.Fields, s.Doc)
	p.Println()
	g.printEncodeFunc(p, s.Name, s.Fields)
	p.Println()
	g.printDecodeFunc(p, s.Name, s.Fields)
}

func (g *structGenerator) printDecl(p *gen.Printer, name string, fields []schema.Field, doc []string) {
	var maxNameLen, maxTypeLen int
	for _, f := range fields {
		if len(f.Name) > maxNameLen {
			maxNameLen = len(f.Name)
		}
		if typ := typename(f.Type); len(typ) > maxTypeLen {
			maxTypeLen = len(typ)
		}
	}

	printDoc(p, doc, name+" structure.")
	p.Println(`type `, name, ` struct {`)
	for _, f := range fields {
		fname := gen.RPad(f.Name, maxNameLen)
		ftype := gen.RPad(typename(f.Type), maxTypeLen)
		json := gen.SnakeCase(f.Name)
		if f.Tags.OmitEmpty() {
			json += ",omitempty"
		}

		p.Println(`	`, fname, ` `, ftype, " `json:\"", json, "\"`")
	}
	p.Println(`}`)
}

func (g *structGenerator) printEncodeFunc(p *gen.Printer, name string, fields []schema.Field) {
	p.Println(`// EncodeMsgpack implements the Encoder interface for `, name, `.`)
	p.Println(`func (o *`, name, `) EncodeMsgpack(w *msgpack.Writer) (err error) {`)
	p.Println(`	if err = w.WriteMapHeader(`, len(fields), `); err != nil {`)
	p.Println(`		return err`)
	p.Println(`	}`)
	for _, f := range fields {
		notEmpty := notEmptyCheck(f.Type, "o."+f.Name)
		omitEmpty := notEmpty != "" && f.Tags.OmitEmpty()

		p.Println(`	// `, f.Name)
		p.Println(`	if err = w.WriteInt64(`, f.Ordinal, `); err != nil {`)
		p.Println(`		return err`)
		p.Println(`	}`)
		if omitEmpty {
			p.Println(`	if `, notEmpty, ` {`)
			printEncodeCall(p, f.Type, "o."+f.Name, "\t\t")
			p.Println(`	} else {`)
			p.Println(`		if err = w.WriteNil(); err != nil {`)
			p.Println(`			return err`)
			p.Println(`		}`)
			p.Println(`	}`)
		} else {
			printEncodeCall(p, f.Type, "o."+f.Name, "\t")
		}
	}
	p.Println(`	return nil`)
	p.Println(`}`)
}

func (g *structGenerator) printDecodeFunc(p *gen.Printer, name string, fields []schema.Field) {
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
		printDecodeCall(p, f.Type, "o."+f.Name, "\t\t\t")
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

func notEmptyCheck(t schema.Type, specifier string) string {
	switch t.(type) {
	case *schema.Bool:
		return specifier
	case *schema.Int:
		return specifier + ` != 0`
	case *schema.Float:
		return specifier + ` != 0`
	case *schema.String:
		return specifier + ` != ""`
	case *schema.Bytes:
		return `len(` + specifier + `) != 0`
	case *schema.Array:
		return `len(` + specifier + `) != 0`
	case *schema.Map:
		return `len(` + specifier + `) != 0`
	case *schema.Time:
		return `!` + specifier + `.IsZero()`
	case *schema.Pointer:
		return specifier + ` != nil`
	default:
		return ""
	}
}
