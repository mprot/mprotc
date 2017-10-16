package js

import (
	"github.com/tsne/mpackc/gen"
	"github.com/tsne/mpackc/schema"
)

type enumGenerator struct {
	typeDecls bool
}

func (g *enumGenerator) Generate(p gen.Printer, e *schema.Enum) {
	printDoc(p, e.Doc, e.Name+" enumeration.")
	if g.typeDecls {
		// When we are generating a separate enum type declaration,
		// we do not need the enumerators in the MsgPack type.
		p.Println(`export const `, e.Name, ` = Int`)
	} else {
		p.Println(`export const `, e.Name, ` = {`)

		maxNameLen := g.maxNameLen(e.Enumerators)
		for _, e := range e.Enumerators {
			spaces := gen.RPad("", maxNameLen-len(e.Name))
			p.Println(`	`, e.Name, `: `, spaces, e.Value, `,`)
		}
		p.Println()

		p.Println(`	enc: Int.enc,`)
		p.Println(`	dec: Int.dec,`)
		p.Println(`};`)
	}
}

func (g *enumGenerator) GenerateTypeDecls(p gen.Printer, e *schema.Enum) {
	p.Println(`export declare const enum `, e.Name, ` {`)

	maxNameLen := g.maxNameLen(e.Enumerators)
	for _, e := range e.Enumerators {
		spaces := gen.RPad("", maxNameLen-len(e.Name))
		p.Println(`	`, e.Name, spaces, ` = `, e.Value, `,`)
	}

	p.Println(`}`)
}

func (g *enumGenerator) maxNameLen(enumerators []schema.Enumerator) int {
	n := 0
	for _, e := range enumerators {
		if len(e.Name) > n {
			n = len(e.Name)
		}
	}
	return n
}
