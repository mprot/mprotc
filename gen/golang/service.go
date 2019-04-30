package golang

import (
	"strconv"
	"strings"

	"github.com/mprot/mprotc/gen"
	"github.com/mprot/mprotc/schema"
)

type serviceGenerator struct{}

func (g *serviceGenerator) Generate(p gen.Printer, s *schema.Service, ti *typeinfo) {
	g.printDecl(p, s.Name, s.Methods, s.Doc, ti)
	p.Println()
	g.printRegisterFunc(p, s.Name, s.Methods, ti)
	p.Println()
	g.printClient(p, s.Name, s.Methods, ti)
}

func (g *serviceGenerator) printDecl(p gen.Printer, name string, methods []schema.Method, doc []string, ti *typeinfo) {
	printDoc(p, doc, name+" service.")
	p.Println(`type `, name, ` interface {`)
	for _, m := range methods {
		callTypes := make([]string, 0, 1+len(m.Args))
		callTypes = append(callTypes, "context.Context")
		for _, arg := range m.Args {
			callTypes = append(callTypes, ti.typename(arg))
		}

		returnType := "error"
		if m.Return != nil {
			returnType = "(" + ti.typename(m.Return) + ", error)"
		}

		printDoc(gen.PrefixedPrinter(p, "\t"), m.Doc, "")
		p.Println(`	`, m.Name, `(`, strings.Join(callTypes, ", "), `) `, returnType)
	}
	p.Println(`}`)
}

func (g *serviceGenerator) printRegisterFunc(p gen.Printer, name string, methods []schema.Method, ti *typeinfo) {
	funcName := "Register" + gen.TitleFirstWord(name)

	p.Println(`// `, funcName, ` register`)
	p.Println(`func `, funcName, `(r mrpc.Registry, svc `, name, `) {`)
	p.Println(`	r.Register(mrpc.ServiceSpec{`)
	p.Println(`		Name:    "`, name, `",`)
	p.Println(`		Service: svc,`)
	if len(methods) != 0 {
		p.Println(`		Methods: []mrpc.MethodSpec{`)
		for _, m := range methods {
			p.Println(`			{`)
			p.Println(`				ID: `, m.Ordinal, `,`)
			p.Println(`				Handler: func(ctx context.Context, svc interface{}, body []byte) (p []byte, err error) {`)

			// decode arguments
			p.Println(`					r := msgpack.NewReaderBytes(body)`)
			argNames := make([]string, 0, 1+len(m.Args))
			argNames = append(argNames, "ctx")
			for i, argType := range m.Args {
				arg := "arg" + strconv.FormatInt(int64(i), 10)
				p.Println(`					var `, arg, ` `, ti.typename(argType))

				argvar := newCodecFuncPrinter(arg, argType, "nil")
				argvar.printDecode(gen.PrefixedPrinter(p, "\t\t\t\t\t"), ti, true)

				argNames = append(argNames, arg)
			}

			// call service method and encode result
			callArgs := strings.Join(argNames, ", ")
			if m.Return == nil {
				p.Println(`					return nil, svc.(`, name, `).`, m.Name, `(`, callArgs, `)`)
			} else {
				p.Println(`					resp, err := svc.(`, name, `).`, m.Name, `(`, callArgs, `)`)
				p.Println(`					if err != nil {`)
				p.Println(`						return nil, err`)
				p.Println(`					}`)
				p.Println(`					buf := bytes.NewBuffer(body)`)
				p.Println(`					w := msgpack.NewWriter(buf)`)
				res := newCodecFuncPrinter("resp", m.Return, "nil")
				res.printEncode(gen.PrefixedPrinter(p, "\t\t\t\t\t"))
				p.Println(`					return buf.Bytes(), nil`)
			}

			p.Println(`				},`)
			p.Println(`			},`)
		}
		p.Println(`		},`)
	}
	p.Println(`	})`)
	p.Println(`}`)
}

func (g *serviceGenerator) printClient(p gen.Printer, name string, methods []schema.Method, ti *typeinfo) {
	clientName := gen.TitleFirstWord(name) + "Client"

	p.Println(`// `, clientName, ` defines the client API for `, name, `.`)
	p.Println(`type `, clientName, ` struct {`)
	p.Println(`	c mrpc.Caller`)
	p.Println(`}`)
	p.Println()
	p.Println(`// New`, clientName, ` creates a new client for `, name, `.`)
	p.Println(`func New`, clientName, `(c mrpc.Caller) `, clientName, ` {`)
	p.Println(`	return `, clientName, `{c: c}`)
	p.Println(`}`)
	if len(methods) != 0 {
		for _, m := range methods {
			params := make([]string, 0, 1+len(m.Args))
			params = append(params, "ctx context.Context")
			args := make([]string, 0, len(m.Args))
			for i, argType := range m.Args {
				arg := "arg" + strconv.FormatInt(int64(i), 10)
				params = append(params, arg+" "+ti.typename(argType))
				args = append(args, arg)
			}

			returnType := "(err error)"
			returnStmt := ""
			errorReturnStmt := "err"
			if m.Return != nil {
				returnType = "(res " + ti.typename(m.Return) + ", err error)"
				returnStmt = "res"
				errorReturnStmt = "res, err"
			}

			p.Println()
			p.Println(`// `, m.Name, ` calls the `, name, `.`, m.Name, ` function on the server side.`)
			p.Println(`func (c `, clientName, `) `, m.Name, `(`, strings.Join(params, ", "), `) `, returnType, ` {`)
			p.Println(`	var buf bytes.Buffer`)
			p.Println(`	w := msgpack.NewWriter(&buf)`)

			for i, argType := range m.Args {
				argvar := newCodecFuncPrinter(args[i], argType, returnStmt)
				argvar.printEncode(gen.PrefixedPrinter(p, "\t"))
			}

			p.Println(`	resp, err := c.c.Call(ctx, mrpc.Request{`)
			p.Println(`		Service: "`, name, `",`)
			p.Println(`		Method:  `, m.Ordinal, `,`)
			p.Println(`		Body:    buf.Bytes(),`)
			p.Println(`	})`)
			p.Println(`	if err != nil {`)
			p.Println(`		return `, errorReturnStmt)
			p.Println(`	} else if err = mrpc.ResponseError(resp); err != nil {`)
			p.Println(`		return `, errorReturnStmt)
			p.Println(`	}`)
			if m.Return == nil {
				p.Println(`	return nil`)
			} else {
				p.Println(`	r := msgpack.NewReaderBytes(resp.Body)`)
				resp := newCodecFuncPrinter("res", m.Return, "res")
				resp.printDecode(gen.PrefixedPrinter(p, "\t"), ti, true)
				p.Println(`	return res, nil`)
			}
			p.Println(`}`)
		}
	}
}
