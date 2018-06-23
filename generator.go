package main

import (
	"fmt"
	"os"

	"github.com/mprot/mprotc/gen"
	"github.com/mprot/mprotc/schema"
)

type generatorOptions struct {
	rootPath   string
	outputPath string
	deprecated bool
	dryRun     bool
}

type generator interface {
	Generate(w *gen.FileWriter, s schema.Schema)
}

type codeGenerator struct {
	gen  generator
	opts generatorOptions
}

func newCodeGenerator(g generator, opts generatorOptions) codeGenerator {
	return codeGenerator{
		gen:  g,
		opts: opts,
	}
}

func (g *codeGenerator) Generate(globPatterns []string) error {
	s, err := schema.Parse(g.opts.rootPath, globPatterns)
	if err != nil {
		return err
	}

	if !g.opts.deprecated {
		s.RemoveDeprecated()
	}

	w, err := gen.NewFileWriter(g.opts.outputPath)
	if err != nil {
		return err
	}

	g.gen.Generate(w, s)
	if g.opts.dryRun {
		w.WalkFiles(func(filename string) {
			fmt.Fprintln(os.Stdout, filename)
		})
		return nil
	}
	return w.Flush()
}
