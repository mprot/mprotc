package generator

import (
	"github.com/mprot/mprotc/internal/gen"
	"github.com/mprot/mprotc/internal/gen/golang"
	"github.com/mprot/mprotc/internal/gen/js"
	"github.com/mprot/mprotc/internal/schema"
)

type internalGenerator interface {
	Generate(w *gen.FileWriter, s schema.Schema)
}

type Generator struct {
	newGen     func(opts *Options) internalGenerator
	fileWriter *gen.FileWriter
}

func NewGolang(o GolangOptions) *Generator {
	return &Generator{
		newGen: func(opts *Options) internalGenerator {
			return golang.NewGenerator(o.cast())
		},
	}
}

func NewJavascript(o JavascriptOptions) *Generator {
	return &Generator{
		newGen: func(opts *Options) internalGenerator {
			return js.NewGenerator(o.cast())
		},
	}
}

func (g *Generator) Generate(opts Options) error {
	opts.sanitize()

	s, err := schema.Parse(opts.RootDirectory, opts.GlobPatterns)
	if err != nil {
		return err
	}

	if opts.RemoveDeprecated {
		s.RemoveDeprecated()
	}

	g.fileWriter, err = gen.NewFileWriter(opts.OutputDirectory)
	if err != nil {
		return err
	}

	generator := g.newGen(&opts)
	generator.Generate(g.fileWriter, s)
	return nil
}

func (g *Generator) IterateFiles(iter func(filename string)) {
	if g.fileWriter != nil {
		g.fileWriter.WalkFiles(iter)
	}
}

func (g *Generator) Dump() error {
	if g.fileWriter == nil {
		return nil
	}
	return g.fileWriter.Flush()
}
