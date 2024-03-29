package golang

import (
	"strings"

	"github.com/mprot/mprotc/internal/gen"
)

func printDoc(p gen.Printer, doc []string, fallback string) {
	lines := doc
	if len(lines) == 0 && fallback != "" {
		lines = strings.Split(fallback, "\n")
	}

	for _, ln := range lines {
		if ln == "" {
			p.Println(`//`)
		} else {
			p.Println(`// `, ln)
		}
	}
}
