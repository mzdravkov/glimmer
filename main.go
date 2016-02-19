package main

import (
	"bytes"
	"encoding/json"
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
)

// TODO: would glimmer work on a clean installation?
// Are there any dependencies that are not packaged with it?
// For example goimports?

func main() {
	createProgram().Run()
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

	createAnotatedFunctionsFile(glimmerTmpFolderPath)

	// run goimports on glimmerTmpFolder to remove the glimmer runtime import from files where it is not used
	if err := exec.Command("goimports", "-w", glimmerTmpFolderPath).Run(); err != nil {
		log.Fatal("Couldn't run goimports on the generated source code", err)
	}

	// compile the new modified version of the client's program
	if os.Chdir(glimmerTmpFolderPath) != nil {
		log.Fatal(err)
	}

	buildCommand := exec.Command("go", "build", "-o", "glimmer_tmp")
	if err := buildCommand.Run(); err != nil {
		log.Fatal("Couldn't run go build on the generated source code: ", err)
	}
}

func createAnotatedFunctionsFile(dir string) {
	s := struct {
		Functions []string
	}{annotatedFunctions}

	data, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(filepath.Join(dir, "glimmer_functions.json"), data, os.ModePerm)
}

func writeConfigFile(dir string) {

}
