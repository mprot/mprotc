package schema

import (
	"io"
	"unicode"
	"unicode/utf8"
)

type token string

const (
	eof      token = "eof"
	invalid  token = "invalid"
	semicol  token = ";"
	lbrack   token = "["
	rbrack   token = "]"
	lbrace   token = "{"
	rbrace   token = "}"
	asterisk token = "*"
	assign   token = "="
	period   token = "."
	ident    token = "ident"
	strlit   token = "string"
	intlit   token = "int"
	floatlit token = "float"
	comment  token = "comment"
	// keywords
	packg    token = "package"
	imprt    token = "import"
	constant token = "const"
	enum     token = "enum"
	strct    token = "struct"
	union    token = "union"
	maptype  token = "map"

	bom     = 0xfeff
	runeEOF = -1
)

type tokenizer struct {
	r            io.Reader
	buf          []byte // current read buffer
	rdoff        int    // current read offset in buf
	implicitSemi bool   // insert implicit semicolon?

	ch    rune  // current character
	choff int   // current character offset
	chpos Pos   // current position in r
	err   error // current error

	tokoff int // current token offset in buf (token's start position)
	tokpos Pos // current token position in r
}

func (t *tokenizer) Reset(r io.Reader, filename string, bufsize int) {
	t.r = r
	t.rdoff = 0
	t.implicitSemi = false

	t.ch = '\n' // initialize pos correctly
	t.choff = 0
	t.chpos = Pos{File: filename}
	t.err = nil
	if cap(t.buf) < bufsize {
		t.buf = make([]byte, 0, bufsize)
	} else {
		t.buf = t.buf[:0]
	}

	t.beginToken()

	// read first character
	t.nextChar()
	if t.ch == bom {
		t.nextChar()
	}

	t.beginToken()
}

func (t *tokenizer) Err() error {
	if t.err == io.EOF {
		return nil
	}
	return t.err
}

func (t *tokenizer) Next() (token, string, Pos) {
	t.skipWhitespaces()
	implicitSemi := false
	defer func() { t.implicitSemi = implicitSemi }()

	if isAlpha(t.ch) {
		tok, lit, pos := t.scanIdentifier()
		if tok == ident {
			implicitSemi = true
		}
		return tok, lit, pos
	}
	if t.ch == '-' || isDigit(t.ch) {
		implicitSemi = true
		return t.scanNumber(false)
	}

	switch t.ch {
	case utf8.RuneError:
		implicitSemi = t.implicitSemi // preserve flag on error
		return invalid, "", t.chpos
	case runeEOF:
		if t.implicitSemi {
			return semicol, "\n", t.chpos
		}
		return eof, "", t.chpos
	case bom:
		return t.invalidToken(errInvalidBom, true)
	case '`':
		implicitSemi = true
		return t.scanString(true)
	case '"':
		implicitSemi = true
		return t.scanString(false)
	case ';', '\n':
		// for newline t.implicitSemi was true here
		t.nextChar()
		return t.token(semicol)
	case '[':
		t.nextChar()
		return t.token(lbrack)
	case ']':
		implicitSemi = true
		t.nextChar()
		return t.token(rbrack)
	case '{':
		t.nextChar()
		return t.token(lbrace)
	case '}':
		implicitSemi = true
		t.nextChar()
		return t.token(rbrace)
	case '*':
		t.nextChar()
		return t.token(asterisk)
	case '=':
		t.nextChar()
		return t.token(assign)
	case '.':
		t.nextChar()
		if isDigit(t.ch) {
			implicitSemi = true
			return t.scanNumber(true)
		}
		return t.token(period)
	case '/':
		if t.implicitSemi {
			return semicol, "\n", t.tokpos
		}
		t.nextChar()
		switch t.ch {
		case '/':
			return t.scanLineComment()
		case '*':
			return t.scanBlockComment()
		default:
			return t.invalidToken(errorString("invalid token '/'"), false)
		}
	default:
		implicitSemi = t.implicitSemi // preserve flag on invalid tokens
		var err error
		if unicode.IsGraphic(t.ch) {
			err = errorf("invalid token '%c'", t.ch)
		} else {
			err = errorf("invalid token %#U", t.ch)
		}
		return t.invalidToken(err, true)
	}
}

func (t *tokenizer) scanIdentifier() (token, string, Pos) {
	for isAlpha(t.ch) || isDigit(t.ch) {
		t.nextChar()
	}
	return t.token(ident)
}

func (t *tokenizer) scanNumber(sawPeriod bool) (token, string, Pos) {
	if !sawPeriod {
		if t.ch == '-' {
			t.nextChar()
		}

		if t.ch == '0' {
			t.nextChar()
			if t.ch == 'x' || t.ch == 'X' {
				// hex
				t.nextChar()
				if t.skipDigits(16) == 0 {
					return t.invalidToken(errInvalidHexNumber, false)
				}
				return t.token(intlit)
			}

			t.skipDigits(8)
			octal := t.skipDigits(10) == 0
			if t.ch != '.' && t.ch != 'e' && t.ch != 'E' {
				// oct
				if !octal {
					return t.invalidToken(errInvalidOctNumber, false)
				}
				return t.token(intlit)
			}
		} else {
			t.skipDigits(10)
		}
	}

	// dec or float
	tok := intlit
	if sawPeriod || t.ch == '.' {
		tok = floatlit
		t.nextChar()
		t.skipDigits(10)
	}
	if t.ch == 'e' || t.ch == 'E' {
		tok = floatlit
		t.nextChar()
		if t.ch == '+' || t.ch == '-' {
			t.nextChar()
		}
		if t.skipDigits(10) == 0 {
			return t.invalidToken(errInvalidFloatNumber, false)
		}
	}
	return t.token(tok)
}

func (t *tokenizer) scanString(raw bool) (token, string, Pos) {
	delim := t.ch
	t.nextChar()
	for {
		switch t.ch {
		case utf8.RuneError, runeEOF:
			return t.invalidToken(errStringNotTerminated, true)
		case bom:
			return t.invalidToken(errInvalidBom, true)
		case delim:
			t.nextChar() // skip delim
			return t.token(strlit)
		case '\n':
			if !raw {
				return t.invalidToken(errStringNotTerminated, false)
			}
		case '\\':
			if !raw {
				if err := t.skipEscape(); err != nil {
					return t.invalidToken(err, false)
				}
				continue // next character already read
			}
		}
		t.nextChar()
	}
}

func (t *tokenizer) scanLineComment() (token, string, Pos) {
	for {
		t.nextChar()
		switch t.ch {
		case '\n':
			t.nextChar() // skip newline
			tok, lit, pos := t.token(comment)
			lit = lit[:len(lit)-1] // strip trailing newline
			return tok, lit, pos
		case runeEOF:
			return t.token(comment)
		case bom:
			return t.invalidToken(errInvalidBom, true)
		case utf8.RuneError:
			t.nextChar()
			return invalid, "", t.tokpos
		}
	}
}

func (t *tokenizer) scanBlockComment() (token, string, Pos) {
	for {
		t.nextChar()
		if t.ch == '*' {
			t.nextChar()
			if t.ch == '/' {
				t.nextChar()
				return t.token(comment)
			}
		}
		switch t.ch {
		case utf8.RuneError, runeEOF:
			return t.invalidToken(errCommentNotTerminated, true)
		case bom:
			return t.invalidToken(errInvalidBom, true)
		}
	}
}

func (t *tokenizer) skipDigits(base int) (n int) {
	for {
		switch {
		case '0' <= t.ch && t.ch <= '9':
			if int(t.ch-'0') >= base {
				return
			}
		case ('a' <= t.ch && t.ch <= 'f') || ('A' <= t.ch && t.ch <= 'F'):
			if base != 16 {
				return
			}
		default:
			return
		}
		t.nextChar()
		n++
	}
}

func (t *tokenizer) skipEscape() error {
	t.nextChar() // skip '\\'
	switch t.ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', '"', '\'':
		t.nextChar()
	case '0', '1', '2', '3', '4', '5', '6', '7': // `\` oct oct oct
		if t.skipDigits(8) < 3 {
			return errInvalidEscapeSequence
		}
	case 'x': // `\` "x" hex hex
		t.nextChar()
		if t.skipDigits(16) < 2 {
			return errInvalidEscapeSequence
		}
	case 'u': // `\` "u" hex hex hex hex
		t.nextChar()
		if t.skipDigits(16) < 4 {
			return errInvalidEscapeSequence
		}
	case 'U': // `\` "U" hex hex hex hex hex hex hex hex
		t.nextChar()
		if t.skipDigits(16) < 8 {
			return errInvalidEscapeSequence
		}
	default:
		return errInvalidEscapeSequence
	}
	return nil
}

func (t *tokenizer) skipWhitespaces() {
	for t.ch == ' ' || t.ch == '\t' || t.ch == '\r' || (!t.implicitSemi && t.ch == '\n') {
		t.nextChar()
	}
	t.beginToken()
}

func (t *tokenizer) invalidToken(err error, advance bool) (token, string, Pos) {
	t.setErr(err)
	lit := string(t.ch)
	pos := t.tokpos
	if advance {
		t.nextChar()
		t.beginToken()
	}
	return invalid, lit, pos
}

func (t *tokenizer) token(tok token) (token, string, Pos) {
	lit := string(t.buf[t.tokoff:t.choff])
	pos := t.tokpos
	t.beginToken()

	if tok == ident {
		tok = lookupKeyword(lit)
	}
	return tok, lit, pos
}

func (t *tokenizer) beginToken() {
	t.tokoff = t.choff
	t.tokpos = t.chpos
}

func (t *tokenizer) nextChar() {
	if t.rdoff == len(t.buf) {
		// fill buffer
		if t.tokoff != 0 {
			copy(t.buf, t.buf[t.tokoff:])
			t.buf = t.buf[:len(t.buf)-t.tokoff]
			t.rdoff -= t.tokoff
			t.tokoff = 0
		}
		if len(t.buf) == cap(t.buf) {
			// grow the buffer
			const maxInt = int(^uint(0) >> 1)
			if cap(t.buf) > maxInt/2 {
				t.setErr(errBufferOverflow)
				return
			}
			buf := make([]byte, len(t.buf), 2*cap(t.buf))
			copy(buf, t.buf)
			t.buf = buf
		}

		for {
			n, err := t.r.Read(t.buf[len(t.buf):cap(t.buf)])
			t.buf = t.buf[:len(t.buf)+n]
			if err != nil {
				t.setErr(err)
				break
			}
			if n > 0 {
				break
			}
		}
	}

	t.choff = t.rdoff
	if t.ch == '\n' {
		t.chpos.Line++
		t.chpos.Column = 1
	} else {
		t.chpos.Column++
	}

	ch, n := utf8.DecodeRune(t.buf[t.rdoff:])
	switch ch {
	case 0:
		t.ch = utf8.RuneError
		t.setErr(errNullChar)
	case utf8.RuneError:
		if n == 0 {
			t.ch = runeEOF
		} else {
			t.ch = utf8.RuneError
			t.setErr(errInvalidEncoding)
		}
	default:
		t.ch = ch
		t.rdoff += n
	}
}

func (t *tokenizer) setErr(err error) {
	if t.err == nil || t.err == io.EOF {
		t.err = err
	}
}

func isAlpha(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

func lookupKeyword(lit string) token {
	switch lit {
	case "package":
		return packg
	case "import":
		return imprt
	case "const":
		return constant
	case "enum":
		return enum
	case "struct":
		return strct
	case "union":
		return union
	case "map":
		return maptype
	default:
		return ident
	}
}
