package generator

import (
	"github.com/mprot/mprotc/internal/gen/golang"
	"github.com/mprot/mprotc/internal/gen/js"
)

type Options struct {
	RootDirectory    string
	GlobPatterns     []string
	RemoveDeprecated bool
	OutputDirectory  string
}

func (o *Options) sanitize() {
	if o.RootDirectory == "" {
		o.RootDirectory = "."
	}
	if o.OutputDirectory == "" {
		o.OutputDirectory = "."
	}
}

type GolangOptions struct {
	ImportRoot   string
	ScopedEnums  bool
	UnwrapUnions bool
	TypeID       bool
}

func (o *GolangOptions) sanitize(opts *Options) {
	if o.ImportRoot == "" {
		o.ImportRoot = opts.OutputDirectory
	}
}

func (o *GolangOptions) cast() golang.Options {
	return golang.Options{
		ImportRoot:   o.ImportRoot,
		ScopedEnums:  o.ScopedEnums,
		UnwrapUnions: o.UnwrapUnions,
		TypeID:       o.TypeID,
	}
}

type JavascriptOptions struct {
	TypeDeclarations bool
}

func (o *JavascriptOptions) sanitize(_ *Options) {
}

func (o *JavascriptOptions) cast() js.Options {
	return js.Options{
		TypeDecls: o.TypeDeclarations,
	}
}
