package main

import (
	"bytes"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tucnak/climax"
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

var info types.Info = types.Info{
	Types: make(map[ast.Expr]types.TypeAndValue),
	Defs:  make(map[*ast.Ident]types.Object),
	Uses:  make(map[*ast.Ident]types.Object),
}

func run(path string) {
	fset := token.NewFileSet()

	packages, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	glimmerTmpFolderPath := filepath.Join(path, "glimmer_tmp")
	os.Mkdir(glimmerTmpFolderPath, os.ModePerm)

	conf := types.Config{
		Importer: importer.Default(),
	}

	// we invoke the type checker for each package and
	// store the data in the types.Info structure
	for _, pkg := range packages {
		files := make([]*ast.File, 0, len(pkg.Files))

		for _, value := range pkg.Files {
			files = append(files, value)
		}
		_, err := conf.Check(path, fset, files, &info)
		if err != nil {
			log.Fatal(err)
		}
	}

	funcDeclFinder := new(FuncDeclFinder)
	for _, pkg := range packages {
		for fileName, file := range pkg.Files {
			AddGlimmerImports(fset, packages)

			ast.Walk(funcDeclFinder, file)

			// export the ast to a file in glimmer_tmp directory
			var buf bytes.Buffer
			printer.Fprint(&buf, fset, file)
			newFileName := filepath.Join(glimmerTmpFolderPath, fileName)
			ioutil.WriteFile(newFileName, buf.Bytes(), os.ModePerm)

		}
	}

	// run goimports on glimmerTmpFolder to remove the glimmer runtime import from files where it is not used
	if exec.Command("goimports", "-w", glimmerTmpFolderPath).Run() != nil {
		panic("Couldn't run goimports on the generated source code")
	}
}
