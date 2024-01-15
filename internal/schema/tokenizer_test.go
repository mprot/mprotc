package schema

import (
	"bytes"
	"strings"
	"testing"
)

func TestTokenizerReset(t *testing.T) {
	var tk tokenizer
	tk.Reset(strings.NewReader("foobar"), "", 1024)

	if tk.rdoff != 1 {
		t.Errorf("unexpected initial read offset: %d", tk.rdoff)
	}
	if tk.choff != 0 {
		t.Errorf("unexpected initial character offset: %d", tk.choff)
	}
	if tk.chpos.Line != 1 || tk.chpos.Column != 1 {
		t.Errorf("unexpected initial character position: %+v", tk.chpos)
	}
	if tk.tokoff != tk.choff {
		t.Errorf("unexpected initial token offset: %d", tk.tokoff)
	}
	if tk.tokpos != tk.chpos {
		t.Errorf("unexpected initial token position: %+v", tk.tokpos)
	}

	tk.Reset(strings.NewReader("\ufefffoobar"), "", 1024)
	if tk.rdoff != 4 {
		t.Errorf("unexpected initial read offset: %d", tk.rdoff)
	}
	if tk.choff != 3 {
		t.Errorf("unexpected initial character offset: %d", tk.choff)
	}
	if tk.chpos.Line != 1 || tk.chpos.Column != 2 {
		t.Errorf("unexpected initial character position: %+v", tk.chpos)
	}
	if tk.tokoff != tk.choff {
		t.Errorf("unexpected initial token offset: %d", tk.tokoff)
	}
	if tk.tokpos != tk.chpos {
		t.Errorf("unexpected initial token position: %+v", tk.tokpos)
	}
}

func TestTokenizerTokens(t *testing.T) {
	type test struct {
		token   token
		literal string
	}

	tests := [...]test{
		{lparen, "("},
		{rparen, ")"},
		{lbrack, "["},
		{rbrack, "]"},
		{lbrace, "{"},
		{rbrace, "}"},
		{asterisk, "*"},
		{assign, "="},
		{period, "."},
		{comma, ","},

		{packg, "package"},
		{imprt, "import"},
		{constant, "const"},
		{enum, "enum"},
		{strct, "struct"},
		{service, "service"},
		{union, "union"},
		{maptype, "map"},

		{ident, "ident"},
		{ident, "Lλ"},
		{ident, "foo1234"},

		{strlit, "`str`"},
		{strlit, "`line 1\r\nline2`"},
		{strlit, "`" + `some text
			             and more text` +
			"`",
		},
		{strlit, `"\a\b\f\n\r\t\v\\\"\'"`},
		{strlit, `"\123 \123456 \x1f \x12345678 \u12ef \u12345678 \U1234cdef \U1234567890"`},

		{intlit, "0"},
		{intlit, "1"},
		{intlit, "12345678901234567890"},
		{intlit, "01234567"},
		{intlit, "0xdeadbeef"},

		{intlit, "-0"},
		{intlit, "-1"},
		{intlit, "-12345678901234567890"},
		{intlit, "-01234567"},
		{intlit, "-0xdeadbeef"},

		{floatlit, "0."},
		{floatlit, ".0"},
		{floatlit, ".1"},
		{floatlit, "2.71828182"},
		{floatlit, "7e0"},
		{floatlit, "7e4"},
		{floatlit, "7.e4"},
		{floatlit, "1e+123"},
		{floatlit, "1e-123"},
		{floatlit, "2.71828e-1000"},

		{floatlit, "-0."},
		{floatlit, "-.0"},
		{floatlit, "-.1"},
		{floatlit, "-2.71828182"},
		{floatlit, "-7e0"},
		{floatlit, "-7e4"},
		{floatlit, "-7.e4"},
		{floatlit, "-1e+123"},
		{floatlit, "-1e-123"},
		{floatlit, "-2.71828e-1000"},

		{comment, "/* block comment */"},
		{comment, "/*\r*/"},
		{comment, "//\r\n"},
		{comment, "// line comment \n"},
	}

	var tk tokenizer

	testLit := func(t *testing.T, test test, scenario string, expectedPos Pos) {
		tok, lit, pos := tk.Next()
		if tok != test.token {
			if tok == invalid {
				t.Errorf("unexpected error for %q (%s): %v", test.literal, scenario, tk.Err())
			} else {
				t.Errorf("unexpected token for %q (%s): %v", test.literal, scenario, tok)
			}
		}

		expectedLit := test.literal
		if tok == comment && strings.HasPrefix(expectedLit, "//") {
			// line comments don't have trailing newlines
			expectedLit = strings.TrimSuffix(expectedLit, "\n")
		}
		if lit != expectedLit {
			t.Errorf("unexpected literal for %q (%s): %q", test.literal, scenario, lit)
		}

		if pos != expectedPos {
			t.Errorf("unexpected position for %q (%s): %s", test.literal, scenario, pos)
		}

		// trailing semicolon
		switch test.token {
		case ident, intlit, floatlit, strlit, rparen, rbrack, rbrace:
			if tok, _, _ = tk.Next(); tok != semicol {
				t.Errorf("unexpected token after %q: %s", test.literal, tok)
			}
		}
	}

	t.Run("single", func(t *testing.T) {
		pos := Pos{Line: 1, Column: 1}
		for _, test := range tests {
			if test.token == comment && strings.HasPrefix(test.literal, "//") {
				// test line comments without trailing newline
				test.literal = strings.TrimSuffix(test.literal, "\n")
			}
			tk.Reset(strings.NewReader(test.literal), "", 1024)
			testLit(t, test, "single", pos)

			// expect eof
			tok, _, _ := tk.Next()
			if tok != eof {
				t.Errorf("unexpected token after %q: %s", test.literal, tok)
			}
			if err := tk.Err(); err != nil {
				t.Errorf("unexpected error after %q: %v", test.literal, err)
			}
		}
	})

	t.Run("all", func(t *testing.T) {
		const newline = "\n \t "

		var buf bytes.Buffer
		buf.WriteString("\ufeff")
		for _, test := range tests {
			buf.WriteString(test.literal)
			buf.WriteString(newline)
		}

		tk.Reset(&buf, "", 1024)
		pos := Pos{Line: 1, Column: 2}
		for _, test := range tests {
			testLit(t, test, "all", pos)
			pos.Line++
			if test.token == strlit || test.token == comment {
				pos.Line += strings.Count(test.literal, "\n")
			}
			pos.Column = len(newline)
		}

		tok, _, _ := tk.Next()
		if tok != eof {
			t.Errorf("unexpected token: %s", tok)
		}
		if err := tk.Err(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestTokenizerWithErrors(t *testing.T) {
	tests := [...]struct {
		literal string
		errmsg  string
	}{
		{"\a", "invalid token U+0007"},
		{`#`, "invalid token '#'"},
		{`…`, "invalid token '…'"},
		{`'`, "invalid token '''"},
		{`/`, "invalid token '/'"},
		{`/-`, "invalid token '/'"},

		{`"abc`, errStringNotTerminated.Error()},
		{"\"abc\n", errStringNotTerminated.Error()},
		{"\"abc\n   ", errStringNotTerminated.Error()},
		{"`", errStringNotTerminated.Error()},

		{`"\c"`, errInvalidEscapeSequence.Error()},
		{`"\12"`, errInvalidEscapeSequence.Error()},
		{`"\x1"`, errInvalidEscapeSequence.Error()},
		{`"\xt"`, errInvalidEscapeSequence.Error()},
		{`"\u123"`, errInvalidEscapeSequence.Error()},
		{`"\ut"`, errInvalidEscapeSequence.Error()},
		{`"\U1234567"`, errInvalidEscapeSequence.Error()},
		{`"\Ut"`, errInvalidEscapeSequence.Error()},

		{"/*", errCommentNotTerminated.Error()},

		{"0E", errInvalidFloatNumber.Error()},
		{"078", errInvalidOctNumber.Error()},
		{"07800000009", errInvalidOctNumber.Error()},
		{"0x", errInvalidHexNumber.Error()},
		{"0X", errInvalidHexNumber.Error()},

		{"\x00", errNullChar.Error()},
		{"\"abc\x00def\"", errNullChar.Error()},
		{"//\x00", errNullChar.Error()},
		{"\"abc\x80def\"", errInvalidEncoding.Error()},
		{"\ufeff\ufeff", errInvalidBom.Error()},
		{"//\ufeff", errInvalidBom.Error()},
		{"/*\ufeff*/", errInvalidBom.Error()},
		{`"` + "abc\ufeffdef" + `"`, errInvalidBom.Error()},
	}

	var tk tokenizer
	for _, test := range tests {
		tk.Reset(strings.NewReader(test.literal), "", 1024)

		tok, _, _ := tk.Next()
		if tok != invalid {
			t.Errorf("unexpected token for %q: %s", test.literal, tok)
		}

		err := tk.Err()
		if err == nil || err.Error() != test.errmsg {
			t.Errorf("unexpected error for %q: %v", test.literal, err)
		}
	}
}

func TestTokenizerNextAfterEOF(t *testing.T) {
	const expectedLit = "ident"
	expectedPos := Pos{Line: 1, Column: 1}

	var tk tokenizer
	tk.Reset(strings.NewReader(expectedLit), "", 1024)

	tok, lit, pos := tk.Next()
	if tok != ident {
		t.Errorf("unexpected token: %s", tok)
	}
	if lit != expectedLit {
		t.Errorf("unexpected literal: %q", lit)
	}
	if pos != expectedPos {
		t.Errorf("unexpected position: %s", pos)
	}

	// consume implicit semicolon
	tok, _, _ = tk.Next()
	if tok != semicol {
		t.Errorf("unexpected token: %s", tok)
	}

	// call next after eof
	for i := 0; i < 3; i++ {
		tok, _, _ = tk.Next()
		if tok != eof {
			t.Errorf("unexpected token after eof: %s", tok)
		}
		if err := tk.Err(); err != nil {
			t.Errorf("unexpected error after eof: %v", err)
		}
	}
}

func TestTokenizerBufferResize(t *testing.T) {
	initialBufSize := 20
	alphabet := "abcdefghijklmnopqrstuvwxyz"
	expectedPos := Pos{Line: 1, Column: 1}

	var tk tokenizer
	tk.Reset(strings.NewReader(alphabet), "", initialBufSize)

	tok, lit, pos := tk.Next()
	if tok != ident {
		t.Errorf("unexpected token: %s", tok)
	}
	if lit != alphabet {
		t.Errorf("unexpected literal: %q", lit)
	}
	if pos != expectedPos {
		t.Errorf("unexpected position: %s", pos)
	}

	if cap(tk.buf) <= initialBufSize {
		t.Errorf("unexpected buffer capacity: %d", cap(tk.buf))
	}
}
