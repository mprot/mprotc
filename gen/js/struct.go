package js

import (
	"github.com/mprot/mprotc/gen"
	"github.com/mprot/mprotc/schema"
)

type structGenerator struct{}

func (g *structGenerator) GenerateDecl(p gen.Printer, s *schema.Struct, meta string) {
	meta = meta + "." + s.Name

	printDoc(p, s.Doc, s.Name+" structure.")
	p.Println(`export const `, s.Name, ` = {`)
	g.printEncodeFunc(p, len(s.Fields), meta)
	p.Println()
	g.printDecodeFunc(p, meta)
	p.Println(`};`)
}

func (g *structGenerator) GenerateMetaKey(p gen.Printer, s *schema.Struct) {
	if len(s.Fields) == 0 {
		return
	}

	p.Println(s.Name, `: {`)
	for _, f := range s.Fields {
		p.Println(`	`, f.Ordinal, `: ["`, fieldName(f), `", `, msgpackTypename(f.Type), `],`)
	}
	p.Println(`},`)
}

func (g *structGenerator) GenerateTypeDecls(p gen.Printer, s *schema.Struct) {
	p.Println(`export declare var `, s.Name, `: Type<`, s.Name, `>;`)
	p.Println(`export interface `, s.Name, ` {`)

	for _, f := range s.Fields {
		p.Println(`	`, fieldName(f), `: `, typescriptTypename(f.Type), `;`)
	}

	p.Println(`}`)
}

func (g *structGenerator) printEncodeFunc(p gen.Printer, nFields int, meta string) {
	p.Println(`	enc(buf, v) {`)
	p.Println(`		Map.encHeader(buf, `, nFields, `);`)
	p.Println(`		for(const k in `, meta, `) {`)
	p.Println(`			const f = `, meta, `[k];`)
	p.Println(`			Int.enc(buf, k);`)
	p.Println(`			f[1].enc(buf, v[f[0]]);`)
	p.Println(`		}`)
	p.Println(`	},`)
}

func (g *structGenerator) printDecodeFunc(p gen.Printer, meta string) {
	p.Println(`	dec(buf) {`)
	p.Println(`		const res = {};`)
	p.Println(`		for(let n = Map.decHeader(buf); n > 0; --n) {`)
	p.Println(`			const f = `, meta, `[Int.dec(buf)];`)
	p.Println(`			if(f) {`)
	p.Println(`				res[f[0]] = f[1].dec(buf);`)
	p.Println(`			} else {`)
	p.Println(`				Any.dec(buf);`)
	p.Println(`			}`)
	p.Println(`		}`)
	p.Println(`		return res;`)
	p.Println(`	},`)
}

func fieldName(f schema.Field) string {
	return gen.LowerFirstWord(f.Name)
}
