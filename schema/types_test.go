package schema

import "testing"

func TestBool(t *testing.T) {
	var b Bool
	switch {
	case b.Name() != "bool":
		t.Errorf("unexpected name for bool: %s", b.Name())
	case b.typeid() != "boolean":
		t.Errorf("unexpected type id for bool: %s", b.typeid())
	}
}

func TestInt(t *testing.T) {
	infos := []struct {
		name string
		i    Int
	}{
		{"int", Int{}},
		{"int8", Int{Bits: 8}},
		{"int16", Int{Bits: 16}},
		{"int32", Int{Bits: 32}},
		{"int64", Int{Bits: 64}},
		{"uint", Int{Unsigned: true}},
		{"uint8", Int{Bits: 8, Unsigned: true}},
		{"uint16", Int{Bits: 16, Unsigned: true}},
		{"uint32", Int{Bits: 32, Unsigned: true}},
		{"uint64", Int{Bits: 64, Unsigned: true}},
	}

	for _, info := range infos {
		switch {
		case info.i.Name() != info.name:
			t.Errorf("unexpected name for %s: %s", info.name, info.i.Name())
		case info.i.typeid() != "integer":
			t.Errorf("unexpected type id for %s: %s", info.name, info.i.typeid())
		}
	}
}

func TestFloat(t *testing.T) {
	infos := []struct {
		name string
		f    Float
	}{
		{"float", Float{}},
		{"float32", Float{Bits: 32}},
		{"float64", Float{Bits: 64}},
	}

	for _, info := range infos {
		switch {
		case info.f.Name() != info.name:
			t.Errorf("unexpected name for %s: %s", info.name, info.f.Name())
		case info.f.typeid() != "floating-point":
			t.Errorf("unexpected type id for %s: %s", info.name, info.f.typeid())
		}
	}
}

func TestString(t *testing.T) {
	var s String
	switch {
	case s.Name() != "string":
		t.Errorf("unexpected name for string: %s", s.Name())
	case s.typeid() != "string":
		t.Errorf("unexpected type id for string: %s", s.typeid())
	}
}

func TestBytes(t *testing.T) {
	var b Bytes
	switch {
	case b.Name() != "bytes":
		t.Errorf("unexpected name for bytes: %s", b.Name())
	case b.typeid() != "bytes":
		t.Errorf("unexpected type id for bytes: %s", b.typeid())
	}
}

func TestArray(t *testing.T) {
	infos := []struct {
		name string
		a    Array
	}{
		{"[]bool", Array{Value: &Bool{}}},
		{"[]int", Array{Value: &Int{}}},
		{"[]float32", Array{Value: &Float{32}}},
		{"[]string", Array{Value: &String{}}},
		{"[][]string", Array{Value: &Array{Value: &String{}}}},
		{"[]map[int]string", Array{Value: &Map{&Int{}, &String{}}}},

		{"[2]bool", Array{Size: 2, Value: &Bool{}}},
		{"[3]int", Array{Size: 3, Value: &Int{}}},
		{"[4]float32", Array{Size: 4, Value: &Float{32}}},
		{"[5]string", Array{Size: 5, Value: &String{}}},
		{"[6][]string", Array{Size: 6, Value: &Array{Value: &String{}}}},
		{"[7]map[int]string", Array{Size: 7, Value: &Map{&Int{}, &String{}}}},
	}

	for _, info := range infos {
		switch {
		case info.a.Name() != info.name:
			t.Errorf("unexpected name for %s: %s", info.name, info.a.Name())
		case info.a.typeid() != "array":
			t.Errorf("unexpected type id for %s: %s", info.name, info.a.typeid())
		}
	}
}

func TestMap(t *testing.T) {
	infos := []struct {
		name string
		m    Map
	}{
		{"map[int]bool", Map{&Int{}, &Bool{}}},
		{"map[string]int", Map{&String{}, &Int{}}},
		{"map[uint]float32", Map{&Int{Unsigned: true}, &Float{32}}},
		{"map[bool]string", Map{&Bool{}, &String{}}},
		{"map[string][]string", Map{&String{}, &Array{Value: &String{}}}},
		{"map[string]map[int]string", Map{&String{}, &Map{&Int{}, &String{}}}},
	}

	for _, info := range infos {
		switch {
		case info.m.Name() != info.name:
			t.Errorf("unexpected name for %s: %s", info.name, info.m.Name())
		case info.m.typeid() != "map":
			t.Errorf("unexpected type id for %s: %s", info.name, info.m.typeid())
		}
	}
}

func TestPointer(t *testing.T) {
	infos := []struct {
		name string
		p    Pointer
	}{
		{"*bool", Pointer{&Bool{}}},
		{"*int", Pointer{&Int{}}},
		{"*float32", Pointer{&Float{32}}},
		{"*string", Pointer{&String{}}},
		{"**string", Pointer{&Pointer{&String{}}}},
	}

	for _, info := range infos {
		switch {
		case info.p.Name() != info.name:
			t.Errorf("unexpected name for %s: %s", info.name, info.p.Name())
		case info.p.typeid() != "pointer":
			t.Errorf("unexpected type id for %s: %s", info.name, info.p.typeid())
		}
	}
}

func TestDefinedType(t *testing.T) {
	infos := []struct {
		name string
		typ  DefinedType
	}{
		{"foo", DefinedType{name: "foo"}},
		{"bar", DefinedType{name: "bar"}},
		{"foobar", DefinedType{name: "foobar"}},
	}

	for _, info := range infos {
		switch {
		case info.typ.Name() != info.name:
			t.Errorf("unexpected name for %s: %s", info.name, info.typ.Name())
		case info.typ.typeid() != info.name:
			t.Errorf("unexpected type id for %s: %s", info.name, info.typ.typeid())
		}
	}
}
