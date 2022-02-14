package main

import (
	"fmt"
	"go/importer"
	"go/parser"
	"go/token"
	"runtime/debug"

	"jonwillia.ms/oselect"
)

func main() {
	const path = "jonwillia.ms/oselect"
	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("not ok")
	}
	for _, dep := range info.Deps {
		fmt.Println(dep.Path)
	}
	_ = oselect.Param[string]{}
	// const path = "net/http"
	fset := token.NewFileSet()
	_, err := parser.ParseDir(fset, ".", nil, parser.AllErrors|parser.Trace)
	if err != nil {
		panic(err)
	}
	i := importer.ForCompiler(fset, "source", nil)
	_, err = i.Import(path)
	if err != nil {
		panic(err)
	}
}
