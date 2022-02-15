package main

import (
	"errors"
	"flag"
	"fmt"
	"go/build"
	"io"
	"io/fs"
	"os"
	"runtime"
	"sync"

	"jonwillia.ms/facade/pkg/generator"
)

func MultiWriter(writers ...io.Writer) io.Writer {
	return io.MultiWriter(writers...)
}

func main() {

	// var os_if os_i.Interface
	// var os_if2 interface{ Exit(code int) }
	// os_if = &os_i.Impl{}
	// os_if2 = os_if
	// os_if2.Exit(1)

	outPath := flag.String("out", "", "File or directory destination for generated code (optional)")
	outPkg := flag.String("outpkg", "main", "Package name for generated code")
	makeWorld := flag.Bool("world", false, "Build the entire go std library. (Ignores outpkg)")

	flag.Parse()

	var (
		out   io.Writer = os.Stdout
		isDir bool
		mode  fs.FileMode
	)
	if *outPath != "" {
		fi, err := os.Stat(*outPath)
		if err != nil {
			panic(err)
		}
		isDir = fi.IsDir()
		mode = fi.Mode()
		if !isDir {
			if len(flag.Args()) != 1 {
				panic("exactly one package allowed when writing to a file")
			}
		}
	}

	args := flag.Args()
	if *makeWorld {
		if !isDir {
			panic("must make world to a directory")
		}
		args = generator.AllPackages
	}

	var wg sync.WaitGroup
	wg.Add(len(args))

	sem := make(chan struct{}, runtime.GOMAXPROCS(0))

	for _, arg := range args {
		go func(arg string) {
			sem <- struct{}{}
			defer func() { <-sem }()
			defer wg.Done()
			fmt.Fprintf(os.Stderr, "Building %s\n", arg)

			outPkg := *outPkg

			if *makeWorld {
				outPkg = generator.ShortName(arg)
			}

			gen, err := generator.New(arg, outPkg, *makeWorld)
			if err != nil {
				if !errors.Is(err, &build.NoGoError{}) {
					fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", arg, err)
					return
				}
				panic(err)
			}

			if *outPath != "" {
				outPath := *outPath // shadow

				if isDir {
					fileName := gen.ShortName()
					if *makeWorld {
						outPath += "/" + arg
						err := os.MkdirAll(outPath, mode)
						if err != nil {
							panic(err)
						}
					}
					outPath += "/" + fileName + ".go"
				}

				f, err := os.Create(outPath)
				if err != nil {
					panic(err)
				}
				defer f.Close()
				out = f
			}
			err = gen.Write(out)
			if err != nil {
				panic(err)
			}
		}(arg)
	}

	wg.Wait()
}
