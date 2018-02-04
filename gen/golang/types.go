package golang

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"github.com/mprot/mprotc/gen"
	"github.com/mprot/mprotc/schema"
)

func typename(t schema.Type) string {
	switch t.(type) {
	case *schema.Bytes:
		return "[]byte"
	case *schema.Time:
		return "time.Time"
	default:
		return t.Name()
	}
}

func printEncodeCall(p gen.Printer, t schema.Type, specifier string, indent string) {
	switch t := t.(type) {
	case *schema.Pointer:
		p.Println(indent, `if `, specifier, ` == nil {`)
		p.Println(indent, `	if err = w.WriteNil(); err != nil {`)
		p.Println(indent, `		return err`)
		p.Println(indent, `	}`)
		p.Println(indent, `} else {`)
		printEncodeCall(p, t.Value, "*"+specifier, indent+"\t")
		p.Println(indent, `}`)

	case *schema.Array:
		p.Println(indent, `if err = w.WriteArrayHeader(len(`, specifier, `)); err != nil {`)
		p.Println(indent, `	return err`)
		p.Println(indent, `}`)
		p.Println(indent, `for _, e := range `, specifier, ` {`)
		printEncodeCall(p, t.Value, "e", indent+"\t")
		p.Println(indent, `}`)

	case *schema.Map:
		p.Println(indent, `if err = w.WriteMapHeader(len(`, specifier, `)); err != nil {`)
		p.Println(indent, `	return err`)
		p.Println(indent, `}`)
		p.Println(indent, `for k, v := range `, specifier, ` {`)
		printEncodeCall(p, t.Key, "k", indent+"\t")
		printEncodeCall(p, t.Value, "v", indent+"\t")
		p.Println(indent, `}`)

	case *schema.DefinedType:
		p.Println(indent, `if err = `, specifier, `.EncodeMsgpack(w); err != nil {`)
		p.Println(indent, `	return err`)
		p.Println(indent, `}`)

	default:
		typ := gen.TitleFirstWord(t.Name())
		p.Println(indent, `if err = w.Write`, typ, `(`, specifier, `); err != nil {`)
		p.Println(indent, `	return err`)
		p.Println(indent, `}`)
	}
}

func printDecodeCall(p gen.Printer, t schema.Type, specifier string, indent string) {
	switch t := t.(type) {
	case *schema.Pointer:
		p.Println(indent, `if typ, err := r.Peek(); err != nil {`)
		p.Println(indent, `	return err`)
		p.Println(indent, `} else if typ == msgpack.Nil {`)
		p.Println(indent, `	_ = r.ReadNil()`)
		p.Println(indent, `	`, specifier, ` = nil`)
		p.Println(indent, `} else {`)
		p.Println(indent, `	if `, specifier, ` == nil {`)
		p.Println(indent, `		`, specifier, ` = new(`, typename(t.Value), `)`)
		p.Println(indent, `	}`)
		printDecodeCall(p, t.Value, "*"+specifier, indent+"\t")
		p.Println(indent, `}`)

	case *schema.Array:
		length := alnumOnly(specifier) + "Len"
		if t.Size <= 0 {
			p.Println(indent, length, `, err := r.ReadArrayHeader()`)
			p.Println(indent, `if err != nil {`)
			p.Println(indent, `	return err`)
			p.Println(indent, `}`)
		} else {
			p.Println(indent, `const `, length, ` = `, t.Size)
			p.Println(indent, `err := r.ReadArrayHeaderWithSize(`, length, `)`)
			p.Println(indent, `if err != nil {`)
			p.Println(indent, `	return err`)
			p.Println(indent, `}`)
		}
		if t.Size <= 0 {
			p.Println(indent, `if cap(`, specifier, `) < `, length, ` {`)
			p.Println(indent, `	`, specifier, ` = make([]`, typename(t.Value), `, `, length, `)`)
			p.Println(indent, `} else {`)
			p.Println(indent, `	`, specifier, ` = `, specifier, `[:`, length, `]`)
			p.Println(indent, `}`)
		}
		p.Println(indent, `for i := 0; i < `, length, `; i++ {`)
		printDecodeCall(p, t.Value, specifier+"[i]", indent+"\t")
		p.Println(indent, `}`)

	case *schema.Map:
		length := alnumOnly(specifier) + "Len"
		p.Println(indent, length, `, err := r.ReadMapHeader()`)
		p.Println(indent, `if err != nil {`)
		p.Println(indent, `	return err`)
		p.Println(indent, `}`)
		p.Println(indent, `if `, specifier, ` == nil {`)
		p.Println(indent, `	`, specifier, ` = make(map[`, typename(t.Key), `]`, typename(t.Value), `, `, length, `)`)
		p.Println(indent, `}`)
		p.Println(indent, `for i := 0; i < `, length, `; i++ {`)
		p.Println(indent, `	var k `, typename(t.Key))
		printDecodeCall(p, t.Key, "k", indent+"\t")
		p.Println(indent, `	var v `, typename(t.Value))
		printDecodeCall(p, t.Value, "v", indent+"\t")
		p.Println(indent, `	`, specifier, `[k] = v`)
		p.Println(indent, `}`)

	case *schema.DefinedType:
		if specifier[0] == '*' {
			specifier = specifier[1:]
		}
		p.Println(indent, `if err = `, specifier, `.DecodeMsgpack(r); err != nil {`)
		p.Println(indent, `	return err`)
		p.Println(indent, `}`)

	default:
		typ := gen.TitleFirstWord(t.Name())
		p.Println(indent, `if `, specifier, `, err = r.Read`, typ, `(); err != nil {`)
		p.Println(indent, `	return err`)
		p.Println(indent, `}`)
	}
}

func alnumOnly(s string) string {
	var buf bytes.Buffer
	buf.Grow(len(s))

	for s != "" {
		ch, n := utf8.DecodeRuneInString(s)
		if ch == utf8.RuneError {
			buf.WriteByte(s[0])
			n = 1
		} else if unicode.IsLetter(ch) || unicode.IsNumber(ch) {
			buf.WriteRune(ch)
		}
		s = s[n:]
	}
	return buf.String()
}
