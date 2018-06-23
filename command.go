package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mprot/mprotc/opts"
)

const binName = "mprotc"

type command struct {
	options   func(opts *opts.Opts)
	generator func(opts *opts.Opts) generator
}

func (c *command) exec(opts *opts.Opts, globPatterns []string) error {
	gen := newCodeGenerator(c.generator(opts), generatorOptions{
		rootPath:   opts.String("root"),
		outputPath: opts.String("out"),
		deprecated: opts.Bool("deprecated"),
		dryRun:     opts.Bool("dryrun"),
	})
	return gen.Generate(globPatterns)
}

func (c *command) registerOpts(opts *opts.Opts) {
	if c.options != nil {
		c.options(opts)
	}
}

type commands map[string]command // language => command

func (c commands) Exec(language string, args []string) error {
	opts := opts.New()
	opts.AddString("--root <path>", ".", "Specify the root path of the mprot schema files.")
	opts.AddString("--out <path>", ".", "Specify the output path for the generated code.")
	opts.AddBool("--deprecated", false, "Include the deprecated fields in the generated code.")
	opts.AddBool("--dryrun", false, "Print the names of the generated files only.")

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

	return cmd.exec(opts, fset.Args())
}

func (c commands) printHelp(language string, opts *opts.Opts) {
	w := os.Stderr

	fmt.Fprintln(w, `Usage:`)
	fmt.Fprintln(w, ` `, binName, language, `[options] [schema-file ...]`)
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
