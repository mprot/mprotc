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

	import "external1.mprot"
	import ext "external2.mprot"

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
		Raw  raw                ` + "`" + `16` + "`" + `
		AI   []int              ` + "`" + `17` + "`" + `
		AF   []float32          ` + "`" + `18` + "`" + `
		AS   [2]string          ` + "`" + `19` + "`" + `
		MSS  map[string]string  ` + "`" + `20` + "`" + `
		MFI  map[float64]int    ` + "`" + `21` + "`" + `
		T    time               ` + "`" + `22` + "`" + `
		PS   *string            ` + "`" + `23` + "`" + `
		PE  *E                  ` + "`" + `24` + "`" + `
		E    E                  ` + "`" + `25` + "`" + `
		X1   external1.X        ` + "`" + `26` + "`" + `
		X2   ext.X              ` + "`" + `27` + "`" + `
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

	// my service doc comment
	service Svc {
		// F1 doc
		F1()                        "1"
		F2() int                    "2"
		F3(bool)                    "3"
		F4(bytes) S                 "4"
		F5(ext.X, string, float32)  "5"
		F6(float64, raw, bytes) E   "6"
	}
	`

	imports := [...]*Import{
		{
			pos:  Pos{Line: 5, Column: 2},
			Path: "external1.mprot",
			Name: "external1",
		},
		{
			pos:  Pos{Line: 6, Column: 2},
			Path: "external2.mprot",
			Name: "ext",
		},
	}

	consts := [...]*Const{
		{
			pos:   Pos{Line: 10, Column: 2},
			Name:  "CS",
			Type:  &String{},
			Value: "foo",
		},
		{
			pos:   Pos{Line: 11, Column: 2},
			Name:  "CI",
			Type:  &Int{Bits: 64},
			Value: "7",
		},
		{
			pos:   Pos{Line: 13, Column: 2},
			Doc:   []string{"constant doc comment"},
			Name:  "CF",
			Type:  &Float{Bits: 64},
			Value: "3.1415",
		},
	}

	enums := [...]*Enum{
		{
			pos:  Pos{Line: 54, Column: 2},
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
			pos:  Pos{Line: 20, Column: 2},
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
				{Name: "Raw", Type: &Raw{}, Ordinal: 16, Tags: Tags{}},
				{Name: "AI", Type: &Array{Value: &Int{}}, Ordinal: 17, Tags: Tags{}},
				{Name: "AF", Type: &Array{Value: &Float{Bits: 32}}, Ordinal: 18, Tags: Tags{}},
				{Name: "AS", Type: &Array{Size: 2, Value: &String{}}, Ordinal: 19, Tags: Tags{}},
				{Name: "MSS", Type: &Map{Key: &String{}, Value: &String{}}, Ordinal: 20, Tags: Tags{}},
				{Name: "MFI", Type: &Map{Key: &Float{Bits: 64}, Value: &Int{}}, Ordinal: 21, Tags: Tags{}},
				{Name: "T", Type: &Time{}, Ordinal: 22, Tags: Tags{}},
				{Name: "PS", Type: &Pointer{Value: &String{}}, Ordinal: 23, Tags: Tags{}},
				{Name: "PE", Type: &Pointer{Value: &DefinedType{name: "E", Decl: enums[0]}}, Ordinal: 24, Tags: Tags{}},
				{Name: "E", Type: &DefinedType{name: "E", Decl: enums[0]}, Ordinal: 25, Tags: Tags{}},
				{Name: "X1", Type: &DefinedType{pkg: "external1", name: "X", Decl: imports[0]}, Ordinal: 26, Tags: Tags{}},
				{Name: "X2", Type: &DefinedType{pkg: "ext", name: "X", Decl: imports[1]}, Ordinal: 27, Tags: Tags{}},
			},
		},
	}

	unions := [...]*Union{
		{
			pos:  Pos{Line: 61, Column: 2},
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

	services := [...]*Service{
		{
			pos:  Pos{Line: 69, Column: 2},
			Doc:  []string{"my service doc comment"},
			Name: "Svc",
			Methods: []Method{
				{Doc: []string{"F1 doc"}, Name: "F1", Args: nil, Return: nil, Ordinal: 1, Tags: Tags{}},
				{Name: "F2", Args: nil, Return: &Int{}, Ordinal: 2, Tags: Tags{}},
				{Name: "F3", Args: []Type{&Bool{}}, Return: nil, Ordinal: 3, Tags: Tags{}},
				{Name: "F4", Args: []Type{&Bytes{}}, Return: &DefinedType{name: "S", Decl: structs[0]}, Ordinal: 4, Tags: Tags{}},
				{Name: "F5", Args: []Type{&DefinedType{pkg: "ext", name: "X", Decl: imports[1]}, &String{}, &Float{Bits: 32}}, Return: nil, Ordinal: 5, Tags: Tags{}},
				{Name: "F6", Args: []Type{&Float{Bits: 64}, &Raw{}, &Bytes{}}, Return: &DefinedType{name: "E", Decl: enums[0]}, Ordinal: 6, Tags: Tags{}},
			},
		},
	}

	expectedPackage := &Package{
		pos:  Pos{Line: 3, Column: 2},
		Name: "foo",
	}

	expectedImports := map[string]*Import{
		imports[0].Name: imports[0],
		imports[1].Name: imports[1],
	}

	expectedDecls := [...]Decl{
		consts[0],
		consts[1],
		consts[2],
		structs[0],
		enums[0],
		unions[0],
		services[0],
	}

	var p parser
	file, err := p.Parse(strings.NewReader(input), "")
	if err != nil {
		t.Fatalf("unexpected parsing error: %v", err)
	}

	if !reflect.DeepEqual(file.Doc, []string{"package comment"}) {
		t.Errorf("unexpected package doc: %+v", file.Doc)
	}

	if !reflect.DeepEqual(file.Package, expectedPackage) {
		t.Errorf("unexpected package name: %+v", file.Package)
	}

	if !reflect.DeepEqual(file.Imports, expectedImports) {
		t.Errorf("unexpected imports: %#v", file.Imports)
	}

	if len(file.Decls) != len(expectedDecls) {
		t.Errorf("unexpected number of declarations: %d", len(file.Decls))
	} else {
		for i, decl := range file.Decls {
			if !reflect.DeepEqual(decl, expectedDecls[i]) {
				t.Errorf("unexpected declaration: %#v", decl)
			}
		}
	}
}

func TestParserErrors(t *testing.T) {
	const input = `
	package foo
	import ""                     // invalid import path ""
	import ".mprot"               // invalid import path ".mprot"
	import ext "external1.mprot"
	import ext "external2.mprot"  // import "ext" already defined
	*                             // unexpected token "*"
	const Const = abc             // unexpected "abc"
	strut S {}                    // unexpected token "strut"

	struct T {
		A int 123                              // unexpected token "123"
		B int "a"                              // invalid ordinal
		C int "1 foo:"                         // invalid tag format
		D int ` + "`" + `2 foo:"bar` + "`" + ` // value string not closed
		E **int "3"                            // pointer pointer
		F *[]int "4"                           // pointer to array
		G *map[string]int "5"                  // pointer to map
		H *raw "6"                             // pointer to raw
		I [][]int "7"                          // multidimensional array
		J [0]string "8"                        // invalid array size
		K [-2]string "9"                       // invalid array size
		L map[{]string "10"                    // unexpected token '{'
		L int "11"                             // duplicate struct field
		M string "11"                          // duplicate ordinal
		N X "12"                               // undefined type
		O int                                  // missing tag string
		P Svc "13"                             // service field
	}

	enum T {}                                  // type redeclared

	enum E {
		E1 "1"
		E1 "2"                                 // duplicate enumerator
		E2                                     // missing tag string
	}

	enum F {}

	union U {
		*E           "1"                       // pointer type
		E            "2"
		E            "3"                       // duplicate branch
		T            "3"                       // duplicate ordinal
		Y            "4"                       // undefined type
		[]E          "5"
		[]T          "6"                       // duplicate branch (array)
		map[string]E "7"
		map[int]T    "8"                       // duplicate branch (map)
		int64        "9"                       // duplicate numeric branch
		float32      "10"                      // duplicate numeric branch
		F            "11"                      // duplicate numeric branch
		raw          "12"                      // raw branch
		Svc          "13"                      // service branch
		ext.X                                  // missing tag string
	}

	union V {}                                 // no branches

	union W {
		U "1"                                  // union branch
	}

	service Svc {
		F()     "1"
		F()     "2"                            // duplicate method
		G(Svc)  "3"                            // service argument
		H() Svc "4"                            // service return type
		I()                                    // missing tag string
	}
	`

	expectedErrors := [...]string{
		// parse errors
		`invalid import path ""`,
		`invalid import path ".mprot"`,
		`import "ext" already defined`,
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
		`pointer type *raw not supported`,
		`array type [][]int not supported`,
		`invalid array size 0`,
		`invalid array size -2`,
		`unexpected token "{"`,
		`missing tag string`,
		`type T redeclared (see position 11:2)`,
		`missing tag string`, // enum
		`missing tag string`, // union
		`missing tag string`, // service
		// resolve errors
		`undefined type X`,
		`undefined type Y`,
		// type validation errors
		`duplicate field L in struct T`,
		`duplicate ordinal 11 for field M in struct T`,
		`service field P in struct T`,
		`duplicate enumerator E1 in enum E`,
		`pointer branch *E in union U`,
		`duplicate branch E in union U`,
		`duplicate ordinal 3 for branch T in union U`,
		`duplicate branch []T in union U (only one array branch is allowed)`,
		`duplicate branch map[int]T in union U (only one map branch is allowed)`,
		`duplicate numeric branch int64 in union U`,
		`duplicate numeric branch float32 in union U`,
		`duplicate numeric branch F in union U`,
		`raw branch in union U`,
		`service branch Svc in union U`,
		`union V does not contain a branch`,
		`union branch U in union W`,
		`duplicate method F in service Svc`,
		`service argument in method G of service Svc`,
		`method H of service Svc returns service type`,
	}

	var p parser
	_, err := p.Parse(strings.NewReader(input), "")
	if err == nil {
		t.Fatal("unexpected parse error, got none")
	}
	errs, ok := err.(ErrorList)
	if !ok {
		t.Fatalf("unexpected error type: %T", err)
	}
	if len(errs) < len(expectedErrors) {
		t.Errorf("too few parse errors errors: %d/%d", len(errs), len(expectedErrors))
		t.FailNow()
	} else {
		for i, err := range errs[:len(expectedErrors)] {
			if err.Text != expectedErrors[i] {
				t.Errorf("unexpected error message: %s", err.Text)
			}
		}
	}
}
