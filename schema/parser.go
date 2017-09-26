package schema

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Parse parses a schema which is read from the given reader.
func Parse(r io.Reader) (*Schema, error) {
	var p parser
	return p.Parse(r, "")
}

// ParseFile parses a schema from a file.
func ParseFile(filename string) (*Schema, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var p parser
	return p.Parse(f, filename)
}

type parser struct {
	t          tokenizer
	tok        token
	lit        string
	doc        []string
	pos        Pos
	errs       ErrorList
	idents     map[string]*DefinedType // type name => type
	unresolved []*DefinedType          // unsresolved type names
}

func (p *parser) Parse(r io.Reader, filename string) (*Schema, error) {
	p.t.Init(r, filename, 4096)
	p.errs.clear()
	p.idents = make(map[string]*DefinedType)
	p.next() // scan initial token

	s := p.parseSchema()

	// resolve yet unresolved identifiers
	for _, typ := range p.unresolved {
		if typ.Decl == nil {
			p.errorf("undefined type %s", typ.name)
		}
	}

	return s, p.errs.err()
}

func (p *parser) parseSchema() *Schema {
	s := &Schema{Doc: p.docComments()}

	p.expect(pkg)
	s.Package = p.parseIdent()
	p.expect(semicol)

	for {
		switch p.tok {
		case eof:
			return s
		case constant:
			s.Decls = append(s.Decls, p.parseConst())
		case enum:
			s.Decls = append(s.Decls, p.parseEnum())
		case strct:
			s.Decls = append(s.Decls, p.parseStruct())
		case union:
			s.Decls = append(s.Decls, p.parseUnion())
		case semicol:
			p.next()
		case invalid:
			p.scanError()
			p.next()
		case ident:
			p.errorf("unexpected identifier %q", p.lit)
			p.skipStatement()
		default:
			p.errorf("unexpected token %q", p.lit)
			p.next()
		}
	}
}

func (p *parser) parseConst() *Const {
	c := &Const{pos: p.pos, Doc: p.docComments()}

	p.expect(constant)
	c.Name = p.parseIdent()
	p.expect(assign)
	switch p.tok {
	case intlit:
		c.Type = p.resolve("int64")
		c.Value = p.lit
	case floatlit:
		c.Type = p.resolve("float64")
		c.Value = p.lit
	case strlit:
		c.Type = p.resolve("string")
		c.Value = p.lit[1 : len(p.lit)-1] // trim delimiters
	default:
		p.errorf("unexpected token %q in constant declaration", p.lit)
	}
	p.next()
	p.expect(semicol)
	return c
}

func (p *parser) parseEnum() *Enum {
	e := &Enum{pos: p.pos, Doc: p.docComments()}

	p.expect(enum)
	e.Name = p.parseIdent()
	p.expect(lbrace)

	names := make(map[string]struct{})
	for p.tok == ident {
		name := p.parseIdent()
		value, tags := p.parseTagString(true)
		p.expect(semicol)

		if name == "" {
			continue
		} else if _, has := names[name]; has {
			p.errorf("duplicate enumerator %s in enum %s", name, e.Name)
		}

		e.Enumerators = append(e.Enumerators, Enumerator{
			Name:  name,
			Value: value,
			Tags:  tags,
		})
		names[name] = struct{}{}
	}

	p.expect(rbrace)
	p.expect(semicol)

	p.register(e.Name, e)
	return e
}

func (p *parser) parseStruct() *Struct {
	s := &Struct{pos: p.pos, Doc: p.docComments()}

	p.expect(strct)
	s.Name = p.parseIdent()
	p.expect(lbrace)

	names := make(map[string]struct{})
	ordinals := make(map[int64]struct{})
	for p.tok != rbrace {
		name := p.parseIdent()
		typ := p.parseType()
		ordinal, tags := p.parseTagString(false)
		p.expect(semicol)

		if name == "" {
			continue
		} else if _, has := names[name]; has {
			p.errorf("duplicate field %s in struct %s", name, s.Name)
		} else if _, has := ordinals[ordinal]; has && ordinal != 0 {
			p.errorf("duplicate ordinal %d for field %s in struct %s", ordinal, name, s.Name)
		}

		s.Fields = append(s.Fields, Field{
			Name:    name,
			Type:    typ,
			Ordinal: ordinal,
			Tags:    tags,
		})
		names[name] = struct{}{}
		ordinals[ordinal] = struct{}{}
	}

	p.expect(rbrace)
	p.expect(semicol)

	p.register(s.Name, s)
	return s
}

func (p *parser) parseUnion() *Union {
	u := &Union{pos: p.pos, Doc: p.docComments()}

	p.expect(union)
	u.Name = p.parseIdent()
	p.expect(lbrace)

	names := make(map[string]struct{})
	ordinals := make(map[int64]struct{})
	for p.tok != rbrace {
		typ := p.parseType()
		ordinal, tags := p.parseTagString(false)
		p.expect(semicol)

		switch typ.(type) {
		case nil:
			continue
		case *Pointer:
			p.errorf("pointer types are not supported as a union branch")
		case *DefinedType:
			typeid := typ.typeid()
			if _, has := names[typeid]; has {
				p.errorf("duplicate branch %s in union %s", typeid, u.Name)
			}
		default:
			typeid := typ.typeid()
			if _, has := names[typeid]; has {
				p.errorf("duplicate branch %s in union %s (only one %s type is allowed)", typ.Name(), u.Name, typeid)
			}
		}

		if _, has := ordinals[ordinal]; has && ordinal != 0 {
			p.errorf("duplicate ordinal %d for branch %s in union %s", ordinal, typ.Name(), u.Name)
		}

		u.Branches = append(u.Branches, Branch{
			Type:    typ,
			Ordinal: ordinal,
			Tags:    tags,
		})
		names[typ.typeid()] = struct{}{}
		ordinals[ordinal] = struct{}{}
	}

	p.expect(rbrace)
	p.expect(semicol)

	if len(u.Branches) == 0 {
		p.errorf("union %s has to contain at least one branch", u.Name)
	}

	p.register(u.Name, u)
	return u
}

func (p *parser) parseTagString(negativeOrdinals bool) (int64, Tags) {
	if p.tok != strlit {
		p.expect(strlit)
		return 0, nil
	}
	lit := p.lit[1 : len(p.lit)-1] // trim string delimiters

	// skip whitespaces
	for len(lit) != 0 && lit[0] == ' ' {
		lit = lit[1:]
	}

	// ordinal
	i := 0
	for i < len(lit) && lit[i] != ' ' {
		i++
	}
	ordinal, err := strconv.ParseInt(lit[:i], 10, 64)
	if err != nil || (!negativeOrdinals && ordinal <= 0) {
		p.errorf("invalid ordinal %q", lit[:i])
	}
	lit = lit[i:]
	i = 0

	// tags
	tags := make(map[string]string)
	for {
		// skip whitespaces
		for len(lit) != 0 && lit[0] == ' ' {
			lit = lit[1:]
		}
		if len(lit) == 0 {
			break
		}

		// key
		i = 0
		for i < len(lit) && lit[i] != ' ' && lit[i] != ':' && lit[i] != '"' {
			i++
		}
		if i == len(lit) || lit[i] == ' ' {
			// key without value
			tags[lit[:i]] = ""
			lit = lit[i:]
			continue
		}
		if i == 0 || i+1 == len(lit) || lit[i] != ':' || lit[i+1] != '"' {
			p.errorf("invalid tag format %s", p.lit)
			break
		}
		key := lit[:i]
		lit = lit[i+1:]

		// value
		i = 2 // skip ':' and '"'
		for i < len(lit) && lit[i] != '"' {
			if lit[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(lit) {
			p.errorf("tag value string not closed for %q", key)
			break
		}
		tags[key] = lit[1:i]
		lit = lit[i+1:]
	}
	p.next()
	return ordinal, tags
}

func (p *parser) parseType() Type {
	switch p.tok {
	case lbrack: // []type or [n]type
		p.expect(lbrack)
		var size uint64
		if p.tok == intlit {
			if sz, err := strconv.ParseUint(p.lit, 10, 32); err != nil || sz <= 0 {
				p.errorf("invalid array size %s", p.lit)
			} else {
				size = sz
			}
			p.next()
		}
		p.expect(rbrack)
		val := p.parseType()
		if _, ok := val.(*Array); ok {
			p.errorf("array type []%s not supported", val.Name())
		}
		return &Array{Size: int(size), Value: val}

	case asterisk: // *type
		p.expect(asterisk)
		val := p.parseType()
		_, isPtr := val.(*Pointer)
		_, isArr := val.(*Array)
		_, isMap := val.(*Map)
		if isPtr || isArr || isMap {
			p.errorf("pointer type *%s not supported", val.Name())
		}
		return &Pointer{Value: val}

	case maptype: // map[type]type
		p.expect(maptype)
		p.expect(lbrack)
		key := p.parseType()
		p.expect(rbrack)
		val := p.parseType()
		return &Map{Key: key, Value: val}

	case ident:
		name := p.lit
		p.next()
		return p.resolve(name)

	default:
		if p.tok != invalid { // invalid already reported an error
			p.errorf("unexpected token %q", p.lit)
		}
		p.next()
		return nil
	}
}

func (p *parser) parseIdent() string {
	if p.tok != ident {
		p.expect(ident)
		return ""
	}

	ident := p.lit
	p.next()
	return ident
}

func (p *parser) resolve(ident string) Type {
	switch ident {
	case "bool":
		return &Bool{}
	case "int":
		return &Int{}
	case "int8":
		return &Int{Bits: 8}
	case "int16":
		return &Int{Bits: 16}
	case "int32":
		return &Int{Bits: 32}
	case "int64":
		return &Int{Bits: 64}
	case "uint":
		return &Int{Unsigned: true}
	case "uint8":
		return &Int{Bits: 8, Unsigned: true}
	case "uint16":
		return &Int{Bits: 16, Unsigned: true}
	case "uint32":
		return &Int{Bits: 32, Unsigned: true}
	case "uint64":
		return &Int{Bits: 64, Unsigned: true}
	case "float32":
		return &Float{Bits: 32}
	case "float64":
		return &Float{Bits: 64}
	case "string":
		return &String{}
	case "bytes":
		return &Bytes{}
	case "time":
		return &Time{}
	default:
		if typ, has := p.idents[ident]; has {
			return typ
		}
		typ := p.register(ident, nil)
		p.unresolved = append(p.unresolved, typ)
		return typ
	}
}

func (p *parser) register(name string, decl Decl) *DefinedType {
	typ := p.idents[name]
	switch {
	case typ == nil:
		typ = &DefinedType{name: name, Decl: decl}
		p.idents[name] = typ
	case typ.Decl == nil:
		typ.Decl = decl
	default:
		p.errorf("type %s redeclared (see position %s)", name, typ.Decl.Pos())
	}
	return typ
}

func (p *parser) expect(tok token) {
	if p.tok != tok {
		switch {
		case p.tok == invalid:
			p.scanError()
		case p.lit == "\n":
			p.errorf("unexpected newline (%s expected)", tok)
		default:
			p.errorf("unexpected token %q (%s expected)", p.lit, tok)
		}
	}
	p.next()
}

func (p *parser) scanError() {
	p.errorf(p.t.Err().Error())
}

func (p *parser) errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.errs.add(errorString(msg), p.pos)
}

func (p *parser) next() {
	p.tok, p.lit, p.pos = p.t.Next()
	p.scanDocComment()
}

func (p *parser) scanDocComment() {
	p.doc = p.doc[:0]

	for p.tok == comment {
		lineCount := len(p.doc)
		p.doc = appendCommentLines(p.doc, p.lit)
		lineCount = len(p.doc) - lineCount

		line := p.pos.Line
		p.tok, p.lit, p.pos = p.t.Next()
		if line+lineCount < p.pos.Line {
			p.doc = p.doc[:0]
		}
	}

	// remove leading and trailing blank lines
	lead := 0
	for lead < len(p.doc) && p.doc[lead] == "" {
		lead++
	}
	if lead != 0 {
		n := copy(p.doc, p.doc[lead:])
		p.doc = p.doc[:n]
	}

	trail := len(p.doc)
	for trail > 0 && p.doc[trail-1] == "" {
		trail--
	}
	p.doc = p.doc[:trail]
}

func (p *parser) docComments() []string {
	var s []string
	if n := len(p.doc); n != 0 {
		s = make([]string, len(p.doc))
		copy(s, p.doc)
	}
	return s
}

func (p *parser) skipStatement() {
	// scan until the next semicolon appears which is not in a nested scope
	level := 0
	for {
		p.next()
		switch p.tok {
		case eof, invalid:
			return
		case semicol:
			if level == 0 {
				return
			}
		case lbrace:
			level++
		case rbrace:
			level--
		}
	}
}

func appendCommentLines(lines []string, c string) []string {
	if c[1] == '/' {
		// line comment
		c = c[2:]
		if len(c) != 0 && c[0] == ' ' {
			c = c[1:]
		}
		lines = append(lines, trimTrailingSpaces(c))
	} else {
		// block comment
		c = c[2 : len(c)-2]
		for len(c) != 0 {
			ln := c
			idx := strings.IndexByte(c, '\n')
			if idx < 0 {
				c = c[:0]
			} else {
				ln = c[:idx]
				c = c[idx+1:]
			}
			lines = append(lines, trimTrailingSpaces(ln))
		}
	}
	return lines
}

func trimTrailingSpaces(s string) string {
	n := len(s)
	for n != 0 {
		switch s[n-1] {
		case ' ', '\t', '\r', '\n':
			n--
		default:
			return s[:n]
		}
	}
	return ""
}