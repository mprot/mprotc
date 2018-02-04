package golang

import (
	"math"

	"github.com/mprot/mprotc/gen"
	"github.com/mprot/mprotc/schema"
)

type enumGenerator struct {
	scoped bool
}

func (g *enumGenerator) Generate(p gen.Printer, e *schema.Enum) {
	intType := "int"
	for _, e := range e.Enumerators {
		if e.Value < math.MinInt32 || e.Value > math.MaxInt32 {
			intType = "int64"
			break
		}
	}

	g.printDecl(p, e.Name, e.Enumerators, intType, e.Doc)
	p.Println()
	g.printEncodeFunc(p, e.Name, intType)
	p.Println()
	g.printDecodeFunc(p, e.Name, intType)
}

func (g *enumGenerator) printDecl(p gen.Printer, name string, enumerators []schema.Enumerator, intType string, doc []string) {
	maxNameLen := 0
	for _, e := range enumerators {
		if len(e.Name) > maxNameLen {
			maxNameLen = len(e.Name)
		}
	}

	printDoc(p, doc, name+" enumeration.")
	p.Println(`type `, name, ` `, intType)
	if len(enumerators) != 0 {
		p.Println()
		p.Println(`// Enumerators for `, name, `.`)
		p.Println(`const (`)
		for _, e := range enumerators {
			enumerator := gen.RPad(e.Name, maxNameLen)
			if g.scoped {
				enumerator = name + enumerator
			}

			p.Println(`	`, enumerator, ` `, name, ` = `, e.Value)
		}
		p.Println(`)`)
	}
}

func (g *enumGenerator) printEncodeFunc(p gen.Printer, name string, intType string) {
	writeInt := "Write" + gen.TitleFirstWord(intType)

	p.Println(`// EncodeMsgpack implements the Encoder interface for `, name, `.`)
	p.Println(`func (o `, name, `) EncodeMsgpack(w *msgpack.Writer) error {`)
	p.Println(`	return w.`, writeInt, `(`, intType, `(o))`)
	p.Println(`}`)
}

func (g *enumGenerator) printDecodeFunc(p gen.Printer, name string, intType string) {
	readInt := "Read" + gen.TitleFirstWord(intType)

	p.Println(`// DecodeMsgpack implements the Decoder interface for `, name, `.`)
	p.Println(`func (o *`, name, `) DecodeMsgpack(r *msgpack.Reader) error {`)
	p.Println(`	val, err := r.`, readInt, `()`)
	p.Println(`	if err != nil {`)
	p.Println(`		return err`)
	p.Println(`	}`)
	p.Println(`	*o = `, name, `(val)`)
	p.Println(`	return nil`)
	p.Println(`}`)
}
