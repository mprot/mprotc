package js

import (
	"github.com/mprot/mprotc/gen"
	"github.com/mprot/mprotc/schema"
)

type structGenerator struct{}

func (g *structGenerator) GenerateDecl(p gen.Printer, s *schema.Struct) {
	printDoc(p, s.Doc, s.Name+" structure.")
	p.Println(`export const `, s.Name, ` = Struct({`)

	for _, f := range s.Fields {
		p.Println(`	`, f.Ordinal, `: ["`, fieldName(f), `", `, msgpackTypename(f.Type), `],`)
	}

	p.Println(`});`)
}

func (g *structGenerator) GenerateTypeDecls(p gen.Printer, s *schema.Struct) {
	p.Println(`export declare var `, s.Name, `: Type<`, s.Name, `>;`)
	p.Println(`export interface `, s.Name, ` {`)

	for _, f := range s.Fields {
		p.Println(`	`, fieldName(f), `: `, typescriptTypename(f.Type), `;`)
	}

	p.Println(`}`)
}

func fieldName(f schema.Field) string {
	return gen.LowerFirstWord(f.Name)
}
