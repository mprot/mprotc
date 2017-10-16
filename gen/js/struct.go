package js

import (
	"github.com/tsne/mpackc/gen"
	"github.com/tsne/mpackc/schema"
)

type structGenerator struct{}

func (g *structGenerator) Generate(p gen.Printer, s *schema.Struct) {
	printDoc(p, s.Doc, s.Name+" structure.")
	p.Println(`export const `, s.Name, ` = {`)
	g.printEncodeFunc(p, s.Fields)
	p.Println()
	g.printDecodeFunc(p, s.Fields)
	p.Println(`};`)
}

func (g *structGenerator) GenerateTypeDecls(p gen.Printer, s *schema.Struct) {
	p.Println(`export declare var `, s.Name, `: Type<`, s.Name, `>;`)
	p.Println(`export interface `, s.Name, ` {`)

	for _, f := range s.Fields {
		p.Println(`	`, fieldName(f), `: `, typescriptTypename(f.Type), `;`)
	}

	p.Println(`}`)
}

func (g *structGenerator) printEncodeFunc(p gen.Printer, fields []schema.Field) {
	p.Println(`	enc(buf, v) {`)
	p.Println(`		Map.encHeader(buf, `, len(fields), `);`)

	for _, f := range fields {
		p.Println(`		Int.enc(buf, `, f.Ordinal, `);`)
		p.Println(`		`, msgpackTypename(f.Type), `.enc(buf, v.`, fieldName(f), `);`)
	}

	p.Println(`	},`)
}

func (g *structGenerator) printDecodeFunc(p gen.Printer, fields []schema.Field) {
	p.Println(`	dec(buf) {`)
	p.Println(`		const res = {};`)
	p.Println(`		let n = Map.decHeader(buf);`)
	p.Println(`		while(n-- > 0) {`)
	p.Println(`			switch(Int.dec(buf)) {`)

	for _, f := range fields {
		fname := fieldName(f)
		p.Println(`			case `, f.Ordinal, `: // `, fname)
		p.Println(`				res.`, fname, ` = `, msgpackTypename(f.Type), `.dec(buf); break;`)
	}

	p.Println(`			default:`)
	p.Println(`				Any.dec(buf)`)
	p.Println(`			}`)
	p.Println(`		}`)
	p.Println(`		return res;`)
	p.Println(`	},`)
}

func fieldName(f schema.Field) string {
	return gen.LowerFirstWord(f.Name)
}
