# mpackc

mpackc is a code generator for a [MessagePack](https://msgpack.org/) compatible codec with a schema definition language. Structs are not serialized to a map with string keys, instead an integer is assigned to each struct field which is used as the key.

## Schema
An mpack schema can be compiled to one or more programming languages. This can be achieved by `mpackc`:
```
mpackc <language> --out=output/path schema1.mpack schema2.mpack
```
For a list of supported languages run `mpackc help` and for language specific help `mpackc help <language>`.

A schema definition looks like this:
```
package foo

const C = 3.141592

enum E {
	This `1`
	That `2`
}

struct S {
	Foo int         `1`
	Bar string      `2 omitempty`  // not written to the stream, if empty
	NotUsed float32 `3 deprecated` // will not be compiled
}

union {
	int `1`
	S   `2`
}
```

The schema supports the following types:
* boolean: `bool`
* signed integer: `int`, `int8`, `int16`, `int32`, `int64`
* unsigned integer: `uint`, `uint8`, `uint16`, `uint32`, `uint64`
* floating-point: `float32`, `float64`
* string: `string`
* binary data: `bytes`
* array: `[]T`
* pointer: `*T`
* map: `map[K]V`
* date/time: `time`
