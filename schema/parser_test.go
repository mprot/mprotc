package schema

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	const input = `
	// package comment
	package foo

	// line comment

	const CS = "foo"
	const CI = 7
	// constant doc comment
	const CF = 3.1415

	/*
		my struct
		doc comment
	*/
	// another doc line
	struct S {
		B    bool               ` + "`" + ` 1 tagkey:"tagval" ` + "`" + `
		I    int                ` + "`" + ` 2 tagkey:"f\too"` + "`" + `
		I8   int8               ` + "`" + ` 3 tagkey` + "`" + `
		I16  int16              ` + "`" + ` 4` + "`" + `
		I32  int32              ` + "`" + ` 5` + "`" + `
		I64  int64              ` + "`" + ` 6` + "`" + `
		UI   uint               ` + "`" + ` 7` + "`" + `
		UI8  uint8              ` + "`" + ` 8` + "`" + `
		UI16 uint16             ` + "`" + ` 9` + "`" + `
		UI32 uint32             ` + "`" + `10` + "`" + `
		UI64 uint64             ` + "`" + `11` + "`" + `
		F32  float32            ` + "`" + `12` + "`" + `
		F64  float64            ` + "`" + `13` + "`" + `
		S    string             ` + "`" + `14` + "`" + `
		Bin  bytes              ` + "`" + `15` + "`" + `
		AI   []int              ` + "`" + `16` + "`" + `
		AF   []float32          ` + "`" + `17` + "`" + `
		AS   [2]string          ` + "`" + `18` + "`" + `
		MSS  map[string]string  ` + "`" + `19` + "`" + `
		MFI  map[float64]int    ` + "`" + `20` + "`" + `
		T    time               ` + "`" + `21` + "`" + `
		PS   *string            ` + "`" + `22` + "`" + `
		PE  *E                  ` + "`" + `23` + "`" + `
		E    E                  ` + "`" + `24` + "`" + `
	}
	
	; // empty statement

	// my enum
	// doc comment
	enum E {
		Val1 "1"
		Val2 "2"
		Val3 "3"
	}

	// my union doc comment
	union U {
		S            "1"
		E            "2"
		[]S          "3"
		map[string]S "4"
	}
	`

	consts := [...]*Const{
		{
			pos:   Pos{Line: 7, Column: 2},
			Name:  "CS",
			Type:  &String{},
			Value: "foo",
		},
		{
			pos:   Pos{Line: 8, Column: 2},
			Name:  "CI",
			Type:  &Int{Bits: 64},
			Value: "7",
		},
		{
			pos:   Pos{Line: 10, Column: 2},
			Doc:   []string{"constant doc comment"},
			Name:  "CF",
			Type:  &Float{Bits: 64},
			Value: "3.1415",
		},
	}

	enums := [...]*Enum{
		{
			pos:  Pos{Line: 48, Column: 2},
			Doc:  []string{"my enum", "doc comment"},
			Name: "E",
			Enumerators: []Enumerator{
				{Name: "Val1", Value: 1, Tags: Tags{}},
				{Name: "Val2", Value: 2, Tags: Tags{}},
				{Name: "Val3", Value: 3, Tags: Tags{}},
			},
		},
	}

	structs := [...]*Struct{
		{
			pos:  Pos{Line: 17, Column: 2},
			Doc:  []string{"\t\tmy struct", "\t\tdoc comment", "", "another doc line"},
			Name: "S",
			Fields: []Field{
				{Name: "B", Type: &Bool{}, Ordinal: 1, Tags: Tags{"tagkey": "tagval"}},
				{Name: "I", Type: &Int{}, Ordinal: 2, Tags: Tags{"tagkey": "f\\too"}},
				{Name: "I8", Type: &Int{Bits: 8}, Ordinal: 3, Tags: Tags{"tagkey": ""}},
				{Name: "I16", Type: &Int{Bits: 16}, Ordinal: 4, Tags: Tags{}},
				{Name: "I32", Type: &Int{Bits: 32}, Ordinal: 5, Tags: Tags{}},
				{Name: "I64", Type: &Int{Bits: 64}, Ordinal: 6, Tags: Tags{}},
				{Name: "UI", Type: &Int{Unsigned: true}, Ordinal: 7, Tags: Tags{}},
				{Name: "UI8", Type: &Int{Bits: 8, Unsigned: true}, Ordinal: 8, Tags: Tags{}},
				{Name: "UI16", Type: &Int{Bits: 16, Unsigned: true}, Ordinal: 9, Tags: Tags{}},
				{Name: "UI32", Type: &Int{Bits: 32, Unsigned: true}, Ordinal: 10, Tags: Tags{}},
				{Name: "UI64", Type: &Int{Bits: 64, Unsigned: true}, Ordinal: 11, Tags: Tags{}},
				{Name: "F32", Type: &Float{Bits: 32}, Ordinal: 12, Tags: Tags{}},
				{Name: "F64", Type: &Float{Bits: 64}, Ordinal: 13, Tags: Tags{}},
				{Name: "S", Type: &String{}, Ordinal: 14, Tags: Tags{}},
				{Name: "Bin", Type: &Bytes{}, Ordinal: 15, Tags: Tags{}},
				{Name: "AI", Type: &Array{Value: &Int{}}, Ordinal: 16, Tags: Tags{}},
				{Name: "AF", Type: &Array{Value: &Float{Bits: 32}}, Ordinal: 17, Tags: Tags{}},
				{Name: "AS", Type: &Array{Size: 2, Value: &String{}}, Ordinal: 18, Tags: Tags{}},
				{Name: "MSS", Type: &Map{Key: &String{}, Value: &String{}}, Ordinal: 19, Tags: Tags{}},
				{Name: "MFI", Type: &Map{Key: &Float{Bits: 64}, Value: &Int{}}, Ordinal: 20, Tags: Tags{}},
				{Name: "T", Type: &Time{}, Ordinal: 21, Tags: Tags{}},
				{Name: "PS", Type: &Pointer{Value: &String{}}, Ordinal: 22, Tags: Tags{}},
				{Name: "PE", Type: &Pointer{Value: &DefinedType{name: "E", Decl: enums[0]}}, Ordinal: 23, Tags: Tags{}},
				{Name: "E", Type: &DefinedType{name: "E", Decl: enums[0]}, Ordinal: 24, Tags: Tags{}},
			},
		},
	}

	unions := [...]*Union{
		{
			pos:  Pos{Line: 55, Column: 2},
			Doc:  []string{"my union doc comment"},
			Name: "U",
			Branches: []Branch{
				{Type: &DefinedType{name: "S", Decl: structs[0]}, Ordinal: 1, Tags: Tags{}},
				{Type: &DefinedType{name: "E", Decl: enums[0]}, Ordinal: 2, Tags: Tags{}},
				{Type: &Array{Value: &DefinedType{name: "S", Decl: structs[0]}}, Ordinal: 3, Tags: Tags{}},
				{Type: &Map{Key: &String{}, Value: &DefinedType{name: "S", Decl: structs[0]}}, Ordinal: 4, Tags: Tags{}},
			},
		},
	}

	expectedDecls := [...]Decl{
		consts[0],
		consts[1],
		consts[2],
		structs[0],
		enums[0],
		unions[0],
	}

	schema, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected parsing error: %v", err)
	}

	if !reflect.DeepEqual(schema.Doc, []string{"package comment"}) {
		t.Errorf("unexpected package doc: %+v", schema.Doc)
	}

	if schema.Package != "foo" {
		t.Errorf("unexpected package name: %q", schema.Package)
	}

	if len(schema.Decls) != len(expectedDecls) {
		t.Errorf("unexpected number of declarations: %d", len(schema.Decls))
	} else {
		for i, decl := range schema.Decls {
			if !reflect.DeepEqual(decl, expectedDecls[i]) {
				t.Errorf("unexpected declaration: %#v", decl)
			}
		}
	}
}

func TestParseWithError(t *testing.T) {
	const input = `
	package foo
	*                 // unexpected token "*"
	const Const = abc // unexpected "abc"
	strut S {}        // unexpected token "strut"

	struct T {
		A int 123                              // unexpected token "123"
		B int "a"                              // invalid ordinal
		C int "1 foo:"                         // invalid tag format
		D int ` + "`" + `2 foo:"bar` + "`" + ` // value string not closed
		E **int "3"                            // pointer pointer
		F *[]int "4"                           // pointer to array
		G *map[string]int "5"                  // pointer to map
		H [][]int "6"                          // multidimensional array
		I [0]string "7"                        // invalid array size
		J [-2]string "8"                       // invalid array size
		K map[{]string "9"                     // unexpected token '{'
		K int "10"                             // duplicate struct field
		L string "10"                          // duplicate ordinal
		M X "11"                               // undefined type
	}

	enum T {}                                  // type redeclared

	enum E {
		E1 "1"
		E1 "2"                                 // duplicate enumerator
	}

	union U {
		*E           "1"                       // pointer type
		E            "2"
		E            "3"                       // duplicate branch
		T            "3"                       // duplicate ordinal
		Y            "4"
		[]E          "5"
		[]T          "6"                       // duplicate branch (array)
		map[string]E "7"
		map[int]T    "8"                       // duplicate branch (map)
	}

	union V {}                                 // no branches
	`

	expectedErrors := [...]string{
		`unexpected token "*"`,
		`unexpected token "abc" in constant declaration`,
		`unexpected identifier "strut"`,
		`unexpected token "123" (string expected)`,
		`invalid ordinal "a"`,
		`invalid tag format "1 foo:"`,
		`tag value string not closed for "foo"`,
		`pointer type **int not supported`,
		`pointer type *[]int not supported`,
		`pointer type *map[string]int not supported`,
		`array type [][]int not supported`,
		`invalid array size 0`,
		`invalid array size -2`,
		`unexpected token "{"`,
		`duplicate field K in struct T`,
		`duplicate ordinal 10 for field L in struct T`,
		`type T redeclared (see position 7:2)`,
		`duplicate enumerator E1 in enum E`,
		`pointer types are not supported as a union branch`,
		`duplicate branch E in union U`,
		`duplicate ordinal 3 for branch T in union U`,
		`duplicate branch []T in union U (only one array type is allowed)`,
		`duplicate branch map[int]T in union U (only one map type is allowed)`,
		`union V has to contain at least one branch`,
		// resolving is the last step
		`undefined type X`,
		`undefined type Y`,
	}

	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("unexpected parse error, got none")
	}
	errs, ok := err.(ErrorList)
	if !ok {
		t.Fatalf("unexpected error type: %T", err)
	}
	if len(errs) < len(expectedErrors) {
		t.Errorf("unexpected number of errors: %d", len(errs))
		t.FailNow()
	} else {
		for i, e := range errs[:len(expectedErrors)] {
			if err, ok := e.(Error); !ok {
				t.Errorf("unexpected error type: %T", e)
			} else if msg := err.Err.Error(); msg != expectedErrors[i] {
				t.Errorf("unexpected error message: %s", msg)
			}
		}
	}
}
