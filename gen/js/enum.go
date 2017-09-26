package js

import (
	"github.com/tsne/mpackc/gen"
	"github.com/tsne/mpackc/schema"
)

type enumGenerator struct{}

func (g *enumGenerator) Generate(p *gen.Printer, e *schema.Enum) {
	printDoc(p, e.Doc, e.Name+" enumeration.")
	p.Println(`export const `, e.Name, ` = {`)
	g.printEnumerators(p, e.Enumerators)
	p.Println()
	p.Println(`	enc: Int.enc,`)
	p.Println(`	dec: Int.dec,`)
	p.Println(`};`)
}

func (g *enumGenerator) printEnumerators(p *gen.Printer, enumerators []schema.Enumerator) {
	maxNameLen := 0
	for _, e := range enumerators {
		if len(e.Name) > maxNameLen {
			maxNameLen = len(e.Name)
		}
	}

	for _, e := range enumerators {
		spaces := gen.RPad("", maxNameLen-len(e.Name))

		p.Println(`	`, e.Name, `: `, spaces, e.Value, `,`)
	}
}
