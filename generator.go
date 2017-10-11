package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tsne/mpackc/gen"
	"github.com/tsne/mpackc/schema"
)

type generatorOptions struct {
	language   string
	outputPath string
	deprecated bool
}

type generator interface {
	Generate(p *gen.Printer, s *schema.Schema)
}

type codeGenerator struct {
	generator
	outputPath string
	fileExt    string
	deprecated bool
}

func newCodeGenerator(g generator, opts generatorOptions) codeGenerator {
	fileExt := "." + opts.language
	if fext, ok := g.(interface {
		FileExt() string
	}); ok {
		fileExt = fext.FileExt()
	}

	return codeGenerator{
		generator:  g,
		outputPath: opts.outputPath,
		fileExt:    fileExt,
		deprecated: opts.deprecated,
	}
}

func (g *codeGenerator) Generate(p *gen.Printer, s *schema.Schema, inputFilename string) error {
	if !g.deprecated {
		g.removeDeprecated(s)
	}
	g.generator.Generate(p, s)

	filename := strings.TrimSuffix(filepath.Base(inputFilename), filepath.Ext(inputFilename))
	path := filepath.Join(g.outputPath, filename+g.fileExt)
	return g.write(s, path, p)
}

func (g *codeGenerator) write(s *schema.Schema, filename string, source io.WriterTo) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = source.WriteTo(f)
	return err
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
