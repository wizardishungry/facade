package main

import (
	"bytes"
	"fmt"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"testing"

	"jonwillia.ms/oselect"
	_ "jonwillia.ms/oselect"
)

func Test(t *testing.T) {
	// const path = "jonwillia.ms/oselect" // does not work tests!
	_ = oselect.Param[string]{}
	const path = "net/http"
	i := importer.Default()
	fset := token.NewFileSet()
	_, err := parser.ParseDir(fset, ".", nil, parser.AllErrors)
	if err != nil {
		panic(err)
	}
	i = importer.ForCompiler(fset, "source", nil)
	if i == nil {
		panic("nil importer")
	}
	pkg, err := i.Import(path)
	if err != nil {
		panic(err)
	}
	scope := pkg.Scope()
	newPkg := types.NewPackage("foo", "foo")
	newPkg.SetImports(
		[]*types.Package{
			pkg,
		},
	)

	var qf types.Qualifier = types.RelativeTo(newPkg)

	qf = func(p *types.Package) string {
		return "http" // TODO should be map
	}

	var (
		headerBuf, interfaceBuf, structBuf bytes.Buffer
	)

	headerBuf.WriteString("package foo\n")

	// todo doc block here
	interfaceBuf.WriteString("type someInterface interface {\n")
	structBuf.WriteString("type someStruct struct {}\n")
	for _, n := range scope.Names() {
		o := pkg.Scope().Lookup(n)
		if !o.Exported() {
			continue
		}
		t := o.Type()
		switch t := t.(type) {
		case *types.Signature:

			var buf bytes.Buffer
			types.WriteSignature(&buf, t, qf)
			fmt.Fprintf(&interfaceBuf, "%s %s\n", o.Name(), buf.String())
			fmt.Fprintf(&structBuf, "func (_ %s) %s %s{\n", "someStruct", o.Name(), buf.String())

			if t.Results() != nil {
				fmt.Fprintf(&structBuf, "return ")
			}
			fmt.Fprintf(&structBuf, "%s.%s(", "http", o.Name())

			for i := 0; i < t.Params().Len(); i++ {
				p := t.Params().At(i)
				fmt.Fprintf(&structBuf, "%s,", p.Name())
			}
			fmt.Fprintf(&structBuf, ")\n}\n")

		}
	}
	interfaceBuf.WriteString("}\n")

	var buf bytes.Buffer
	buf.Write(headerBuf.Bytes())
	buf.Write(interfaceBuf.Bytes())
	buf.Write(structBuf.Bytes())

	fmted, err := format.Source(buf.Bytes())
	if err != nil {
		os.Stdout.Write(buf.Bytes())
		panic(err)
	}
	os.Stdout.Write(fmted)
	return

}
