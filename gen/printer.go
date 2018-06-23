package gen

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Printer defines an interface for printing lines of code into a memory buffer.
type Printer interface {
	Println(args ...interface{}) // always appends newline
}

// PrefixedPrinter returns a printer which prefixes each printed line with the
// given prefix.
func PrefixedPrinter(p Printer, prefix string) Printer {
	return &prefixedPrinter{
		p:      p,
		prefix: prefix,
	}
}

type printer struct {
	bytes.Buffer
}

func (p *printer) Println(args ...interface{}) {
	fmt.Fprint(p, args...)
	p.WriteByte('\n')
}

func (p *printer) writeToFile(filename string) error {
	if dir := filepath.Dir(filename); dir != "." && dir != string(filepath.Separator) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := f.Write(p.Bytes())
	switch {
	case err != nil:
		return err
	case n != p.Len():
		return io.ErrShortWrite
	default:
		return nil
	}
}

type prefixedPrinter struct {
	p      Printer
	prefix string
}

func (p *prefixedPrinter) Println(args ...interface{}) {
	a := make([]interface{}, 0, 1+len(args))
	a = append(a, p.prefix)
	a = append(a, args...)
	p.p.Println(a...)
}
