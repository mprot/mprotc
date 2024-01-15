package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mprot/mprotc/generator"
)

const binName = "mprotc"

type Command struct {
	Options   func(opts *Opts)
	Generator func(opts *Opts) *generator.Generator
}

func (c *Command) exec(opts *Opts, globPatterns []string) error {
	dryRun := opts.Bool("dryrun")

	gen := c.Generator(opts)
	err := gen.Generate(generator.Options{
		RootDirectory:    opts.String("root"),
		GlobPatterns:     globPatterns,
		RemoveDeprecated: !opts.Bool("deprecated"),
		OutputDirectory:  opts.String("out"),
	})
	if err != nil {
		return err
	}

	if dryRun {
		gen.IterateFiles(func(filename string) {
			fmt.Fprintln(os.Stdout, filename)
		})
		return nil
	}

	return gen.Dump()
}

func (c *Command) registerOpts(opts *Opts) {
	if c.Options != nil {
		c.Options(opts)
	}
}

type Commands map[string]Command // language => command

func (c Commands) Exec(language string, args []string) error {
	opts := NewOpts()
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

func (c Commands) printHelp(language string, opts *Opts) {
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

func (c Commands) supportedLangs() []string {
	langs := make([]string, 0, len(c))
	for lang := range c {
		langs = append(langs, lang)
	}
	return langs
}
