package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
// For example goimports? Any other?

func main() {
	createProgram().Run()
}

// info is for storing type information that we get from runnig the type checker
var info types.Info = types.Info{
	Types: make(map[ast.Expr]types.TypeAndValue),
	Defs:  make(map[*ast.Ident]types.Object),
	Uses:  make(map[*ast.Ident]types.Object),
}

// run is the main function for the creation of the modified copy of the user project
func run(path string, flags map[string]string) {
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

	for pkgName, pkg := range packages {
		for fileName, file := range pkg.Files {
			addGlimmerImports(fset, packages)

			funcDeclFinder := &funcDeclFinder{Package: pkgName}
			ast.Walk(funcDeclFinder, file)

			// export the ast to a file in glimmer_tmp directory
			var buf bytes.Buffer
			printer.Fprint(&buf, fset, file)
			newFileName := filepath.Join(glimmerTmpFolderPath, fileName)
			ioutil.WriteFile(newFileName, buf.Bytes(), os.ModePerm)

		}
	}

	createAnotatedFunctionsFile(glimmerTmpFolderPath)
	createConfigFile(glimmerTmpFolderPath, flags)

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

// createAnotatedFunctionsFile creates a file containing a JSON with list
// of all annotated functions and saves it to the directory at the provided path
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

// createConfigFile makes a file with configurations for the runtime.
// It serves to give the passed flags to the runtime
func createConfigFile(dir string, flags map[string]string) {
	format := `
[default]
port=%s
delay=%s
`
	data := fmt.Sprintf(format, flags["port"], flags["delay"])

	ioutil.WriteFile(filepath.Join(dir, "glimmer_config.cfg"), []byte(data), os.ModePerm)
}
