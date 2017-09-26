package main

import (
	"fmt"
	"os"

	"github.com/tsne/mpackc/gen/golang"
	"github.com/tsne/mpackc/gen/js"
	"github.com/tsne/mpackc/opts"
)

var cli = commands{
	"go": command{
		options: func(opts *opts.Opts) {
			opts.AddBool("--scoped-enums", false, "Scope the enumerators of the generated enums.")
		},

		generator: func(opts *opts.Opts) generator {
			return golang.NewGenerator(golang.Options{
				ScopedEnums: opts.Bool("scoped-enums"),
			})
		},
	},
	"js": command{
		generator: func(opts *opts.Opts) generator {
			return js.NewGenerator()
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
