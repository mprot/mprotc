package schema

import "testing"

func TestErrorString(t *testing.T) {
	err := errorString("foobar")
	if msg := err.Error(); msg != "foobar" {
		t.Errorf("unexpected error message: %q", msg)
	}
}

func TestError(t *testing.T) {
	err := &Error{
		Err: errorString("foobar"),
		Pos: Pos{
			File:   "file",
			Line:   2,
			Column: 4,
		},
	}

	if msg := err.Error(); msg != "file:2:4: foobar" {
		t.Errorf("unexpected error message: %q", msg)
	}
}

func TestErrorList(t *testing.T) {
	var errs ErrorList

	if err := errs.err(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if msg := errs.Error(); msg != "no errors" {
		t.Errorf("unexpected error message: %q", msg)
	}

	errs.add(errorString("foo"), Pos{Line: 2, Column: 4})
	if err := errs.err(); err == nil {
		t.Error("expected error, got none")
	}
	if msg := errs.Error(); msg != "2:4: foo" {
		t.Errorf("unexpected error message: %q", msg)
	}

	errs.add(errorString("bar"), Pos{Line: 3, Column: 1})
	if err := errs.err(); err == nil {
		t.Error("expected error, got none")
	}
	if msg := errs.Error(); msg != "2:4: foo (and 1 more errors)" {
		t.Errorf("unexpected error message: %q", msg)
	}

	errs.clear()
	if len(errs) != 0 {
		t.Errorf("unexpected number of errors: %d", len(errs))
	}
}
