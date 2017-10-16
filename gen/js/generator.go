package js

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tsne/mpackc/gen"
	"github.com/tsne/mpackc/schema"
)

// Options holds all the options for the JavaScript language generator.
type Options struct {
	TypeDecls bool // generate type declarations?
}

// Generator represents a code generator for the JavaScript language.
type Generator struct {
	cnst  constGenerator
	enum  enumGenerator
	strct structGenerator
	union unionGenerator

	typeDecls bool
}

// NewGenerator creates a new JavaScript code generator with the given options.
func NewGenerator(opts Options) *Generator {
	return &Generator{
		enum: enumGenerator{
			typeDecls: opts.TypeDecls,
		},
		typeDecls: opts.TypeDecls,
	}
}

// Generate generates the JavaScript code for the given schema and prints it to p.
func (g *Generator) Generate(w *gen.FileWriter, s *schema.Schema) {
	g.generate(w.Printer(".js"), s)
	if g.typeDecls {
		g.generateTypeDecls(w.Printer(".d.ts"), s)
	}
}

func (g *Generator) generate(p gen.Printer, s *schema.Schema) {
	g.printPreamble(p)
	if len(s.Doc) != 0 {
		printDoc(p, s.Doc, "")
		p.Println()
	}

	g.printImports(p, msgpackImports(s))
	for _, decl := range s.Decls {
		switch decl := decl.(type) {
		case *schema.Const:
			g.cnst.Generate(p, decl)
		case *schema.Enum:
			g.enum.Generate(p, decl)
		case *schema.Struct:
			g.strct.Generate(p, decl)
		case *schema.Union:
			g.union.Generate(p, decl)
		default:
			panic(fmt.Sprintf("unsupported declaration type %T", decl))
		}

		p.Println()
	}

	g.printCollectionTypes(p, s)
}

func (g *Generator) generateTypeDecls(p gen.Printer, s *schema.Schema) {
	g.printPreamble(p)
	g.printImports(p, typescriptImports(s))

	for _, decl := range s.Decls {
		switch decl := decl.(type) {
		case *schema.Const:
			// do nothing
		case *schema.Enum:
			g.enum.GenerateTypeDecls(p, decl)
		case *schema.Struct:
			g.strct.GenerateTypeDecls(p, decl)
		case *schema.Union:
			g.union.GenerateTypeDecls(p, decl)
		default:
			panic(fmt.Sprintf("unsupported declaration type %T", decl))
		}

		p.Println()
	}
}

func (g *Generator) printPreamble(p gen.Printer) {
	p.Println(`// Code generated by mpackc.`)
	p.Println(`// Do not edit.`)
	p.Println()
}

func (g *Generator) printImports(p gen.Printer, imports []string) {
	if imp := strings.Join(imports, ", "); imp != "" {
		p.Println(`import {`, imp, `} from "messagepack";`)
		p.Println()
		p.Println()
	}
}

func (g *Generator) printCollectionTypes(p gen.Printer, s *schema.Schema) {
	types := map[string]string{} // typename => type declaration
	iterTypes(s, func(t schema.Type) {
		switch t := t.(type) {
		case *schema.Array:
			types[msgpackTypename(t)] = fmt.Sprintf("TypedArr(%s)", msgpackTypename(t.Value))
		case *schema.Map:
			types[msgpackTypename(t)] = fmt.Sprintf("TypedMap(%s, %s)", msgpackTypename(t.Key), msgpackTypename(t.Value))
		}
	})

	if len(types) == 0 {
		return
	}

	typedefs := make([]string, 0, len(types))
	for name, decl := range types {
		typedefs = append(typedefs, fmt.Sprintf("const %s = %s;", name, decl))
	}
	sort.Strings(typedefs)

	p.Println(`// Required collection types.`)
	for _, typedef := range typedefs {
		p.Println(typedef)
	}
}
