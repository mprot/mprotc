package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mprot/mprotc/gen/golang"
	"github.com/mprot/mprotc/gen/js"
	"github.com/mprot/mprotc/opts"
	"github.com/mprot/mprotc/schema"
)

var cli = commands{
	"go": command{
		options: func(opts *opts.Opts) {
			opts.AddString("--import-root", "", "Import root path for all schema imports.")
			opts.AddBool("--scoped-enums", false, "Scope the enumerators of the generated enums.")
			opts.AddBool("--unwrap-union", false, "Unwrap union types of the generated struct fields.")
		},

		generator: func(opts *opts.Opts) generator {
			importRoot := opts.String("import-root")
			if importRoot == "" {
				importRoot = opts.String("out")
			}
			return golang.NewGenerator(golang.Options{
				ImportRoot:  importRoot,
				ScopedEnums: opts.Bool("scoped-enums"),
				UnwrapUnion: opts.Bool("unwrap-union"),
			})
		},
	},
	"js": command{
		options: func(opts *opts.Opts) {
			opts.AddBool("--typedecls", false, "Generate type declarations in a separate .d.ts file.")
		},

		generator: func(opts *opts.Opts) generator {
			return js.NewGenerator(js.Options{
				TypeDecls: opts.Bool("typedecls"),
			})
		},
	},
}

func main() {
	if len(os.Args) < 2 {
		cli.Exec("help", nil)
		os.Exit(0)
	}

	err := cli.Exec(os.Args[1], os.Args[2:])
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
