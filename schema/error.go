package schema

import (
	"fmt"
	"sort"
)

const (
	errInvalidOctNumber      = errorString("invalid octal number")
	errInvalidHexNumber      = errorString("invalid hexadecimal number")
	errInvalidFloatNumber    = errorString("invalid floating-point number")
	errStringNotTerminated   = errorString("string not terminated")
	errCommentNotTerminated  = errorString("comment not terminated")
	errInvalidEscapeSequence = errorString("invalid escape sequence")
	errNullChar              = errorString("invalid character nul")
	errInvalidEncoding       = errorString("invalid encoding")
	errBufferOverflow        = errorString("buffer overflow")
	errInvalidBom            = errorString("invalid byte order mark")
)

type errorReporter interface {
	errorf(format string, args ...interface{})
}

type errorString string

func errorf(format string, args ...interface{}) error {
	return errorString(fmt.Sprintf(format, args...))
}

func (e errorString) Error() string {
	return string(e)
}

type Error struct {
	Pos  Pos
	Text string
}

func (e Error) Error() string {
	return e.Pos.String() + ": " + e.Text
}

type ErrorList []Error

func (e ErrorList) Error() string {
	switch len(e) {
	case 0:
		return "no errors"
	case 1:
		return e[0].Error()
	case 2:
		return fmt.Sprintf("%v (and 1 more error)", e[0])
	default:
		return fmt.Sprintf("%v (and %d more errors)", e[0], len(e)-1)
	}
}

func (e ErrorList) err() error {
	if len(e) == 0 {
		return nil
	}
	return e
}

func (e *ErrorList) add(pos Pos, text string) {
	*e = append(*e, Error{
		Pos:  pos,
		Text: text,
	})
}

func (e ErrorList) concat(el ErrorList) ErrorList {
	return append(e, el...)
}

func (e ErrorList) sort() {
	sort.Slice(e, func(i, j int) bool {
		left, right := e[i].Pos, e[j].Pos
		switch {
		case left.File != right.File:
			return left.File < right.File
		case left.Line != right.Line:
			return left.Line < right.Line
		case left.Column != right.Column:
			return left.Column < right.Column
		default:
			return false
		}
	})
}
