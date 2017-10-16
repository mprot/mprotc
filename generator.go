package main

import (
	"github.com/tsne/mpackc/gen"
	"github.com/tsne/mpackc/schema"
)

type generatorOptions struct {
	outputPath  string
	deprecated  bool
	packageName string
}

type generator interface {
	Generate(w *gen.FileWriter, s *schema.Schema)
}

type codeGenerator struct {
	generator
	outputPath  string
	deprecated  bool
	packageName string
}

func newCodeGenerator(g generator, opts generatorOptions) codeGenerator {
	return codeGenerator{
		generator:   g,
		outputPath:  opts.outputPath,
		deprecated:  opts.deprecated,
		packageName: opts.packageName,
	}
}

func (g *codeGenerator) Generate(s *schema.Schema, inputFilename string) error {
	if !g.deprecated {
		g.removeDeprecated(s)
	}
	if g.packageName != "" {
		s.Package = g.packageName
	}

	w := gen.NewFileWriter(g.outputPath, inputFilename)
	g.generator.Generate(w, s)
	return w.Flush()
}

func (g *codeGenerator) removeDeprecated(s *schema.Schema) {
	for _, decl := range s.Decls {
		switch decl := decl.(type) {
		case *schema.Enum:
			idx := 0
			for i := 0; i < len(decl.Enumerators); i++ {
				if !decl.Enumerators[i].Tags.Deprecated() {
					decl.Enumerators[idx] = decl.Enumerators[i]
					idx++
				}
			}
			decl.Enumerators = decl.Enumerators[:idx]

		case *schema.Struct:
			idx := 0
			for i := 0; i < len(decl.Fields); i++ {
				if !decl.Fields[i].Tags.Deprecated() {
					decl.Fields[idx] = decl.Fields[i]
					idx++
				}
			}
			decl.Fields = decl.Fields[:idx]

		case *schema.Union:
			idx := 0
			for i := 0; i < len(decl.Branches); i++ {
				if !decl.Branches[i].Tags.Deprecated() {
					decl.Branches[idx] = decl.Branches[i]
					idx++
				}
			}
			decl.Branches = decl.Branches[:idx]
		}
	}
}
