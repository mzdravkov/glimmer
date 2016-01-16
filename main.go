package main

import (
	"github.com/tucnak/climax"
	"go/ast"
	"go/parser"
	"go/token"
)

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
}

func run(path string) {
	fset := token.NewFileSet()

	packages, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	funcDeclFinder := new(FuncDeclFinder)
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			ast.Walk(funcDeclFinder, file)
		}
	}
}
