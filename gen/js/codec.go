package js

import (
	"sort"
	"strconv"

	"github.com/mprot/mprotc/schema"
)

type codecContext struct {
	decl      schema.Decl
	codecName string
	typeID    int
}

func (c *codecContext) Decl() schema.Decl {
	return c.decl
}

func (c *codecContext) Key() string {
	return strconv.FormatInt(int64(c.typeID), 10)
}

func (c *codecContext) EncodeFunc() string {
	return c.codecName + `.enc`
}

func (c *codecContext) DecodeFunc() string {
	return c.codecName + `.dec`
}

type codec struct {
	name string
	ids  map[schema.Decl]int // schema declaration => type id
}

func newCodec(name string) *codec {
	return &codec{
		name: name,
		ids:  make(map[schema.Decl]int),
	}
}

func (c *codec) Name() string {
	return c.name
}

func (c *codec) Context(decl schema.Decl) codecContext {
	id, has := c.ids[decl]
	if !has {
		id = len(c.ids)
		c.ids[decl] = id
	}

	return codecContext{
		decl:      decl,
		codecName: c.name,
		typeID:    id,
	}
}

func (c *codec) Size() int {
	return len(c.ids)
}

func (c *codec) Contexts() []codecContext {
	ctx := make([]codecContext, 0, len(c.ids))
	for decl, id := range c.ids {
		ctx = append(ctx, codecContext{
			decl:      decl,
			codecName: c.name,
			typeID:    id,
		})
	}

	sort.Slice(ctx, func(i, j int) bool { return ctx[i].typeID < ctx[j].typeID })
	return ctx
}
