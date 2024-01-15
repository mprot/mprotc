# Go source code translations
The Go source code uses [msgpack-go](https://github.com/mprot/msgpack-go) for the MessagePack encoding.

## Constant
```golang
const Pi = 3.141592
```

## Enumeration
```golang
type E int // or int64 if the ordinal exceeds 32 bit

const (
    This E = 1 // 'EThis' for scoped enums
    That E = 2 // 'EThat' for scoped enums
)

func (e E) EncodeMsgpack(w *msgpack.Writer) error  { ... }
func (e *E) DecodeMsgpack(r *msgpack.Reader) error { ... }
```

## Struct
```golang
type S struct {
    Foo int
    Bar float32
}

func (s *S) EncodeMsgpack(w *msgpack.Writer) error { ... }
func (s *S) DecodeMsgpack(r *msgpack.Reader) error { ... }
```

## Union
```golang
type U struct {
    Value interface{} // has to be type asserted
}

func (u U) EncodeMsgpack(w *msgpack.Writer) error  { ... }
func (u *U) DecodeMsgpack(r *msgpack.Reader) error { ... }
```
