package main

import (
	"fmt"
	"os"

	"github.com/mprot/mprotc/gen/golang"
	"github.com/mprot/mprotc/gen/js"
	"github.com/mprot/mprotc/opts"
)

var cli = commands{
	"go": command{
		options: func(opts *opts.Opts) {
			opts.AddBool("--scoped-enums", false, "Scope the enumerators of the generated enums.")
			opts.AddBool("--unwrap-union", false, "Unwrap union types of the generated struct fields.")
		},

		generator: func(opts *opts.Opts) generator {
			return golang.NewGenerator(golang.Options{
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
