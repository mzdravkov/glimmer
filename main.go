package main

import (
	"bytes"
	// "fmt"
	"github.com/tucnak/climax"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
)

//TODO: find a good solution to the problem with hijiking only the channels
// used in the annotated channels: Hijiking a channel and not writting to the original one
// will cause all reads from the original channel in unannotated functions to hang forever

func main() {
	program := climax.New("glimmer")

	program.Brief = "Glimmer is a tool that visualises the communication between goroutines"
	program.Version = "0.0.1"

	on := climax.Command{
		Name:  "on",
		Brief: "set the path to the project, onto which you want to run glimmer",
		Usage: `glimmer on /path/to/some/project`,
		Help:  `set the path to the project, onto which you want to run glimmer`,

		Flags: []climax.Flag{},

		Examples: []climax.Example{},

		Handle: func(ctx climax.Context) int {
			run(ctx.Args[0])
			return 0
		},
	}

	program.AddCommand(on)
	program.Run()
}

// glimmer
func testFunc() {
	ch := make(chan int, 2)
	ch <- 1
	<-ch
	println(len(ch))
}

func run(path string) {
	fset := token.NewFileSet()

	packages, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	os.Mkdir("glimmer_tmp", os.ModePerm)

	funcDeclFinder := new(FuncDeclFinder)
	for _, pkg := range packages {
		for filename, file := range pkg.Files {
			ast.Walk(funcDeclFinder, file)

			// export the ast to a file in glimmer_tmp directory
			var buf bytes.Buffer
			printer.Fprint(&buf, fset, file)
			fileName := filepath.Join("glimmer_tmp", filename)
			ioutil.WriteFile(fileName, buf.Bytes(), os.ModePerm)
		}
	}
}
