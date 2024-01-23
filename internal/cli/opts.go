package cli

import (
	"flag"
	"fmt"
	"strings"
	"unicode"
)

type option struct {
	val   any
	usage string
	help  string
}

// Opts represents a set of command line options.
type Opts struct {
	opts  map[string]option // name => option
	names []string
}

// New creates new options.
func NewOpts() *Opts {
	return &Opts{
		opts: make(map[string]option),
	}
}

// RegisterAt registers all set options at the given flag set.
func (o *Opts) RegisterAt(fset *flag.FlagSet) {
	for _, name := range o.names {
		opt := o.opts[name]
		switch val := opt.val.(type) {
		case *bool:
			fset.BoolVar(val, name, *val, opt.usage)
		case *int:
			fset.IntVar(val, name, *val, opt.usage)
		case *string:
			fset.StringVar(val, name, *val, opt.usage)
		}
	}
}

// ForEach iterates over all set options.
func (o *Opts) ForEach(f func(usage, help string)) {
	for _, name := range o.names {
		opt := o.opts[name]
		f(opt.usage, opt.help)
	}
}

// AddBool adds a boolean option. The option name is determined by the
// usage, which should be something like "--bool-opt".
func (o *Opts) AddBool(usage string, val bool, help string) {
	b := new(bool)
	*b = val
	o.add(b, usage, help)
}

// AddInt adds an integer option. The option name is determined by the
// usage, which should be something like "--int-opt <n>".
func (o *Opts) AddInt(usage string, val int, help string) {
	i := new(int)
	*i = val
	o.add(i, usage, help)
}

// AddString adds an string option. The option name is determined by the
// usage, which should be something like "--string-opt <s>".
func (o *Opts) AddString(usage string, val string, help string) {
	s := new(string)
	*s = val
	o.add(s, usage, help)
}

// Bool returns the value of the boolean option with the given name. If
// this option is not boolean, it will panic.
func (o *Opts) Bool(name string) bool {
	if v := o.get(name); v != nil {
		return *v.(*bool)
	}
	return false
}

// Int returns the value of the integer option with the given name. If
// this option is not an integer, it will panic.
func (o *Opts) Int(name string) int {
	if v := o.get(name); v != nil {
		return *v.(*int)
	}
	return 0
}

// String returns the value of the string option with the given name. If
// this option is not a string, it will panic.
func (o *Opts) String(name string) string {
	if v := o.get(name); v != nil {
		return *v.(*string)
	}
	return ""
}

func (o *Opts) add(val any, usage string, help string) {
	usage = "--" + strings.TrimLeft(usage, "-")
	name := usage[2:]
	if idx := strings.IndexFunc(name, unicode.IsSpace); idx >= 0 {
		name = name[:idx]
	}

	if _, has := o.opts[name]; has {
		panic(fmt.Sprintf("option %q already defined", name))
	}
	o.opts[name] = option{
		val:   val,
		usage: usage,
		help:  help,
	}
	o.names = append(o.names, name)
}

func (o *Opts) get(name string) any {
	if opt, has := o.opts[name]; has {
		return opt.val
	}
	return nil
}
