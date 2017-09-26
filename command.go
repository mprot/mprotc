package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tsne/mpackc/gen"
	"github.com/tsne/mpackc/opts"
	"github.com/tsne/mpackc/schema"
)

const binName = "mpackc"

type generator interface {
	Generate(p *gen.Printer, s *schema.Schema)
}

type command struct {
	options   func(opts *opts.Opts)
	generator func(opts *opts.Opts) generator
}

func (c *command) generate(opts *opts.Opts, language string, inputFiles []string) error {
	schemas := make([]*schema.Schema, len(inputFiles))
	for i, inputFile := range inputFiles {
		var err error
		if schemas[i], err = schema.ParseFile(inputFile); err != nil {
			return err
		}
	}

	p := &gen.Printer{}
	gen := c.generator(opts)
	ext := "." + language
	if g, ok := gen.(interface {
		FileExt() string
	}); ok {
		ext = g.FileExt()
	}

	out := opts.String("out")
	for i, schema := range schemas {
		gen.Generate(p, schema)

		filename := strings.TrimSuffix(filepath.Base(inputFiles[i]), filepath.Ext(inputFiles[i]))
		path := filepath.Join(out, filename+ext)
		if err := c.write(schema, path, p); err != nil {
			return err
		}
	}
	return nil
}

func (c *command) write(schema *schema.Schema, filename string, source io.WriterTo) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = source.WriteTo(f)
	return err
}

func (c *command) registerOpts(opts *opts.Opts) {
	if c.options != nil {
		c.options(opts)
	}
}

type commands map[string]command // language => command

func (c commands) Exec(language string, args []string) error {
	opts := opts.New()
	opts.AddString("--out <path>", ".", "Specify the output path for the generated code.")

	cmd, has := c[language]
	if !has {
		if strings.ToLower(strings.TrimLeft(language, "-")) != "help" {
			return fmt.Errorf("unknown language %q", language)
		}
		if len(args) != 0 {
			language = args[0]
			cmd, has = c[language]
		}

		if has {
			cmd.registerOpts(opts)
		} else {
			language = "<language>"
		}
		c.printHelp(language, opts)
		return nil
	}

	cmd.registerOpts(opts)

	fset := flag.NewFlagSet(binName, flag.ContinueOnError)
	fset.Usage = func() { c.printHelp(language, opts) }
	opts.RegisterAt(fset)
	if err := fset.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	return cmd.generate(opts, language, fset.Args())
}

func (c commands) printHelp(language string, opts *opts.Opts) {
	w := os.Stderr

	fmt.Fprintln(w, `Usage:`)
	fmt.Fprintln(w, ` `, binName, language, `[options] [file ...]`)
	fmt.Fprintln(w, ` `, binName, `help`, language)
	fmt.Fprintln(w)
	if opts != nil {
		fmt.Fprintln(w, `Options:`)
		opts.ForEach(func(usage, help string) {
			fmt.Fprintln(w, `  `+usage)
			for _, help := range strings.Split(help, "\n") {
				fmt.Fprintln(w, `     `, help)
			}
		})
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, `Supported Languages:`)
	fmt.Fprintln(w, ` `, strings.Join(c.supportedLangs(), ", "))
}

func (c commands) supportedLangs() []string {
	langs := make([]string, 0, len(c))
	for lang := range c {
		langs = append(langs, lang)
	}
	return langs
}
