package golang

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mprot/mprotc/internal/gen"
	"github.com/mprot/mprotc/internal/schema"
)

type codecFuncPrinter struct {
	varname    string      // name of the encodable/decodable variable
	vartype    schema.Type // the type which shall be encoded/decoded
	returnStmt string      // return statement of the function in case of error
}

func newCodecFuncPrinter(varname string, vartype schema.Type, returnStmt string) codecFuncPrinter {
	if returnStmt == "" {
		returnStmt = "err"
	} else {
		returnStmt += ", err"
	}

	return codecFuncPrinter{
		varname:    varname,
		vartype:    vartype,
		returnStmt: returnStmt,
	}
}

func (cp *codecFuncPrinter) printEncode(p gen.Printer) {
	switch t := cp.vartype.(type) {
	case *schema.Pointer:
		p.Println(`if `, cp.varname, ` == nil {`)
		p.Println(`	if err = w.WriteNil(); err != nil {`)
		p.Println(`		return `, cp.returnStmt)
		p.Println(`	}`)
		p.Println(`} else {`)
		deref := codecFuncPrinter{
			varname:    "*" + cp.varname,
			vartype:    t.Value,
			returnStmt: cp.returnStmt,
		}
		deref.printEncode(gen.PrefixedPrinter(p, "\t"))
		p.Println(`}`)

	case *schema.Array:
		p.Println(`if err = w.WriteArrayHeader(len(`, cp.varname, `)); err != nil {`)
		p.Println(`	return `, cp.returnStmt)
		p.Println(`}`)
		p.Println(`for _, e := range `, cp.varname, ` {`)
		elem := codecFuncPrinter{
			varname:    "e",
			vartype:    t.Value,
			returnStmt: cp.returnStmt,
		}
		elem.printEncode(gen.PrefixedPrinter(p, "\t"))
		p.Println(`}`)

	case *schema.Map:
		p.Println(`if err = w.WriteMapHeader(len(`, cp.varname, `)); err != nil {`)
		p.Println(`	return `, cp.returnStmt)
		p.Println(`}`)
		p.Println(`for k, v := range `, cp.varname, ` {`)
		key := codecFuncPrinter{
			varname:    "k",
			vartype:    t.Key,
			returnStmt: cp.returnStmt,
		}
		key.printEncode(gen.PrefixedPrinter(p, "\t"))
		val := codecFuncPrinter{
			varname:    "v",
			vartype:    t.Value,
			returnStmt: cp.returnStmt,
		}
		val.printEncode(gen.PrefixedPrinter(p, "\t"))
		p.Println(`}`)

	case *schema.DefinedType:
		encodevar := strings.TrimPrefix(cp.varname, "*")
		p.Println(`if err = `, encodevar, `.EncodeMsgpack(w); err != nil {`)
		p.Println(`	return `, cp.returnStmt)
		p.Println(`}`)

	default:
		typ := gen.TitleFirstWord(t.Name())
		p.Println(`if err = w.Write`, typ, `(`, cp.varname, `); err != nil {`)
		p.Println(`	return `, cp.returnStmt)
		p.Println(`}`)
	}
}

func (cp *codecFuncPrinter) printDecode(p gen.Printer, ti *typeinfo, noCopy bool) {
	switch t := cp.vartype.(type) {
	case *schema.Pointer:
		p.Println(`if typ, err := r.Peek(); err != nil {`)
		p.Println(`	return `, cp.returnStmt)
		p.Println(`} else if typ == msgpack.Nil {`)
		p.Println(`	_ = r.ReadNil()`)
		p.Println(`	`, cp.varname, ` = nil`)
		p.Println(`} else {`)
		p.Println(`	if `, cp.varname, ` == nil {`)
		p.Println(`		`, cp.varname, ` = new(`, ti.typename(t.Value), `)`)
		p.Println(`	}`)
		deref := codecFuncPrinter{
			varname:    "*" + cp.varname,
			vartype:    t.Value,
			returnStmt: cp.returnStmt,
		}
		deref.printDecode(gen.PrefixedPrinter(p, "\t"), ti, noCopy)
		p.Println(`}`)

	case *schema.Array:
		length := alnumOnly(cp.varname) + "Len"
		if t.Size <= 0 {
			p.Println(length, `, err := r.ReadArrayHeader()`)
			p.Println(`if err != nil {`)
			p.Println(`	return `, cp.returnStmt)
			p.Println(`}`)
		} else {
			p.Println(`const `, length, ` = `, t.Size)
			p.Println(`err := r.ReadArrayHeaderWithSize(`, length, `)`)
			p.Println(`if err != nil {`)
			p.Println(`	return `, cp.returnStmt)
			p.Println(`}`)
		}
		if t.Size <= 0 {
			p.Println(`if cap(`, cp.varname, `) < `, length, ` {`)
			p.Println(`	`, cp.varname, ` = make([]`, ti.typename(t.Value), `, `, length, `)`)
			p.Println(`} else {`)
			p.Println(`	`, cp.varname, ` = `, cp.varname, `[:`, length, `]`)
			p.Println(`}`)
		}
		p.Println(`for i := 0; i < `, length, `; i++ {`)
		elem := codecFuncPrinter{
			varname:    cp.varname + "[i]",
			vartype:    t.Value,
			returnStmt: cp.returnStmt,
		}
		elem.printDecode(gen.PrefixedPrinter(p, "\t"), ti, noCopy)
		p.Println(`}`)

	case *schema.Map:
		length := alnumOnly(cp.varname) + "Len"
		p.Println(length, `, err := r.ReadMapHeader()`)
		p.Println(`if err != nil {`)
		p.Println(`	return `, cp.returnStmt)
		p.Println(`}`)
		p.Println(`if `, cp.varname, ` == nil {`)
		p.Println(`	`, cp.varname, ` = make(map[`, ti.typename(t.Key), `]`, ti.typename(t.Value), `, `, length, `)`)
		p.Println(`}`)
		p.Println(`for i := 0; i < `, length, `; i++ {`)
		p.Println(`	var k `, ti.typename(t.Key))
		key := codecFuncPrinter{
			varname:    "k",
			vartype:    t.Key,
			returnStmt: cp.returnStmt,
		}
		key.printDecode(gen.PrefixedPrinter(p, "\t"), ti, noCopy)
		p.Println(`	var v `, ti.typename(t.Value))
		val := codecFuncPrinter{
			varname:    "v",
			vartype:    t.Value,
			returnStmt: cp.returnStmt,
		}
		val.printDecode(gen.PrefixedPrinter(p, "\t"), ti, noCopy)
		p.Println(`	`, cp.varname, `[k] = v`)
		p.Println(`}`)

	case *schema.Raw:
		p.Println(`if `, cp.varname, `, err = r.ReadRaw(`, cp.varname, `); err != nil {`)
		p.Println(`	return `, cp.returnStmt)
		p.Println(`}`)

	case *schema.DefinedType:
		varname := strings.TrimPrefix(cp.varname, "*")
		p.Println(`if err = `, varname, `.DecodeMsgpack(r); err != nil {`)
		p.Println(`	return `, cp.returnStmt)
		p.Println(`}`)

	case *schema.Bytes:
		if noCopy {
			p.Println(`if `, cp.varname, `, err = r.ReadBytesNoCopy(); err != nil {`)
		} else {
			p.Println(`if `, cp.varname, `, err = r.ReadBytes(nil); err != nil {`)
		}
		p.Println(`	return `, cp.returnStmt)
		p.Println(`}`)

	default:
		typ := gen.TitleFirstWord(t.Name())
		p.Println(`if `, cp.varname, `, err = r.Read`, typ, `(); err != nil {`)
		p.Println(`	return `, cp.returnStmt)
		p.Println(`}`)
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
