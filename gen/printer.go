package gen

import (
	"bytes"
	"fmt"
	"io"
)

// Printer prints lines of code into a memory buffer.
type Printer struct {
	buf bytes.Buffer
}

// Println adds a line of code into the buffer. It always
// adds a newline.
func (p *Printer) Println(args ...interface{}) {
	fmt.Fprint(&p.buf, args...)
	p.buf.WriteByte('\n')
}

// Bytes returns the written lines from the underlying buffer.
func (p *Printer) Bytes() []byte {
	return p.buf.Bytes()
}

// WriteTo writes the contents of the buffer into w and resets
// the buffer.
func (p *Printer) WriteTo(w io.Writer) (int64, error) {
	return p.buf.WriteTo(w)
}
