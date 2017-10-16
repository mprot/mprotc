package golang

import (
	"strconv"

	"github.com/tsne/mpackc/gen"
	"github.com/tsne/mpackc/schema"
)

type constGenerator struct{}

func (g *constGenerator) Generate(p gen.Printer, c *schema.Const) {
	val := c.Value
	if _, isStr := c.Type.(*schema.String); isStr {
		val = strconv.Quote(val)
	}

	printDoc(p, c.Doc, "")
	p.Println(`const `, c.Name, ` = `, val)
}
