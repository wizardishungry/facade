package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"strings"
)

func New(packagePath string) (*Package, error) {
	fset := token.NewFileSet()

	_, err := parser.ParseDir(fset, ".", nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	imp := importer.ForCompiler(fset, "source", nil)
	if imp == nil {
		return nil, fmt.Errorf("importer.ForCompiler() nil")
	}

	pkg, err := imp.Import(packagePath)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(packagePath, "/")

	return &Package{
		pkg:       pkg,
		shortName: parts[len(parts)-1],
	}, nil
}

type Package struct {
	pkg       *types.Package
	shortName string
}

func (p *Package) Write(w io.Writer) error {
	scope := p.pkg.Scope()

	newImports := []*types.Package{
		p.pkg,
	}
	// for _, pkg := range p.pkg.Imports() {
	// 	newImports = append(newImports, pkg)
	// }

	structName := fmt.Sprintf("%sImpl", strings.Title(p.shortName))
	interfaceName := fmt.Sprintf("%sInterface", strings.Title(p.shortName))

	var qf types.Qualifier = func(pkg *types.Package) string {
		return p.shortName // TODO: should be map
	}

	var (
		headerBuf, interfaceBuf, structBuf bytes.Buffer
	)

	// TODO: write doc blocks here
	interfaceBuf.WriteString(fmt.Sprintf("type %s interface {\n", interfaceName))
	structBuf.WriteString(fmt.Sprintf("type %s struct {}\n", structName))
	for _, n := range scope.Names() {
		o := p.pkg.Scope().Lookup(n)
		if !o.Exported() {
			continue
		}
		t := o.Type()
		switch t := t.(type) {
		case *types.Signature:

			for i := 0; i < t.TypeParams().Len(); i++ {
				tp := t.TypeParams().At(i)
				newImports = append(newImports, tp.Obj().Pkg())
				// TODO: check that this works with aliased imports
			}

			var buf bytes.Buffer
			types.WriteSignature(&buf, t, qf)
			fmt.Fprintf(&interfaceBuf, "%s %s\n", o.Name(), buf.String())
			fmt.Fprintf(&structBuf, "func (_ %s) %s %s{\n", structName, o.Name(), buf.String())

			if t.Results() != nil {
				fmt.Fprintf(&structBuf, "return ")
			}
			fmt.Fprintf(&structBuf, "%s.%s(", p.shortName, o.Name())

			for i := 0; i < t.Params().Len(); i++ {
				p := t.Params().At(i)
				fmt.Fprintf(&structBuf, "%s,", p.Name())
			}
			fmt.Fprintf(&structBuf, ")\n}\n")

		}
	}
	interfaceBuf.WriteString("}\n")

	// TODO: write "Do not edit"
	// TODO: support writing build tags fixating to the goos, goarch + version of go that built it
	headerBuf.WriteString("package foo\n")
	headerBuf.WriteString("import(\n")
	for _, imp := range newImports {
		fmt.Fprintf(&headerBuf, "%s \"%s\"\n", imp.Name(), imp.Path())
	}
	headerBuf.WriteString(")\n")

	var buf bytes.Buffer
	buf.Write(headerBuf.Bytes())
	buf.Write(interfaceBuf.Bytes())
	buf.Write(structBuf.Bytes())

	fmted, err := format.Source(buf.Bytes())
	if err != nil {
		os.Stderr.Write(buf.Bytes()) // TODO: remove debugging output
		return err
	}
	_, err = w.Write(fmted)
	return err
}
