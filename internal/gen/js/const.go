package js

import (
	"strconv"

	"github.com/mprot/mprotc/internal/gen"
	"github.com/mprot/mprotc/internal/schema"
)

type constGenerator struct{}

func (g *constGenerator) GenerateDecl(p gen.Printer, c *schema.Const) {
	val := c.Value
	if _, isStr := c.Type.(*schema.String); isStr {
		val = strconv.Quote(val)
	}

	printDoc(p, c.Doc, "")
	p.Println(`export const `, c.Name, ` = `, val, `;`)
}
