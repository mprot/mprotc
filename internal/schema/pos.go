package schema

import "fmt"

// Pos describes a position in the schema file.
type Pos struct {
	File   string
	Line   int
	Column int
}

// String returns a string representation of the position.
func (p Pos) String() string {
	var prefix string
	if p.File != "" {
		prefix = p.File + ":"
	}
	return prefix + fmt.Sprintf("%d:%d", p.Line, p.Column)
}
