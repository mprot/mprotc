package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mprot/mprotc/generator"
	"github.com/mprot/mprotc/internal/cli"
	"github.com/mprot/mprotc/internal/schema"
)

var commands = cli.Commands{
	"go": cli.Command{
		Options: func(opts *cli.Opts) {
			opts.AddString("--import-root", "", "Import root path for all schema imports.")
			opts.AddBool("--scoped-enums", false, "Scope the enumerators of the generated enums.")
			opts.AddBool("--unwrap-unions", false, "Unwrap union types of the generated struct fields.")
			opts.AddBool("--typeid", false, "Generate methods for retrieving a type id.")
		},

		Generator: func(opts *cli.Opts) *generator.Generator {
			return generator.NewGolang(generator.GolangOptions{
				ImportRoot:   opts.String("import-root"),
				ScopedEnums:  opts.Bool("scoped-enums"),
				UnwrapUnions: opts.Bool("unwrap-union"),
				TypeID:       opts.Bool("typeid"),
			})
		},
	},
	"js": cli.Command{
		Options: func(opts *cli.Opts) {
			opts.AddBool("--typedecls", false, "Generate type declarations in a separate .d.ts file.")
		},

		Generator: func(opts *cli.Opts) *generator.Generator {
			return generator.NewJavascript(generator.JavascriptOptions{
				TypeDeclarations: opts.Bool("typedecls"),
			})
		},
	},
}

func main() {
	if len(os.Args) < 2 {
		commands.Exec("help", nil)
		os.Exit(0)
	}

	err := commands.Exec(os.Args[1], os.Args[2:])
	if err != nil {
		printErr(os.Stderr, err)
		os.Exit(1)
	}
}

func printErr(w io.Writer, err error) {
	errs, ok := err.(schema.ErrorList)
	if !ok {
		fmt.Fprintln(w, err)
		return
	}

	const (
		maxFiles       = 5
		maxErrsPerFile = 5
	)

	var (
		filename  string
		fileCount int
		errCount  int
	)
	for _, err := range errs {
		if errCount > maxErrsPerFile {
			continue
		} else if filename != err.Pos.File {
			filename = err.Pos.File
			errCount = 0
			if fileCount++; fileCount > maxFiles {
				break
			}
			fmt.Fprintln(w, "#", filename)
		}
		fmt.Fprintln(w, err.Error())
	}
}
