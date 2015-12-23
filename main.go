package main

import (
	// "github.com/kisielk/gotool"
	// "golang.org/x/tools/go/loader"
	"fmt"
	"github.com/tucnak/climax"
	"go/ast"
	"go/parser"
	"go/token"
	// "path/filepath"
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

type FuncDeclFinder struct{}

// Visit implements the ast.Visitor interface.
func (f *FuncDeclFinder) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.File:
		return f
	case *ast.FuncDecl: // if it is a function declaration
		// prune search if function has no attached comments
		if n.Doc == nil {
			return nil
		}
		// prune search if function has number of comment lines different than one
		if len(n.Doc.List) != 1 {
			return nil
		}
		// prune search if this one comment line is not the one we need
		if n.Doc.List[0].Text != "// glimmer" {
			return nil
		}
		fmt.Println(n.Name)
		sendOrReceiveFinder := new(SendOrReceiveFinder)
		ast.Walk(sendOrReceiveFinder, n.Body)
		return nil // Prune search
	}

	return nil // Prune search
}

type SendOrReceiveFinder struct{}

// Visit implements the ast.Visitor interface.
func (f *SendOrReceiveFinder) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.SendStmt:
		sendStmtCpy := n
		node = new(ast.BlockStmt)
		append(node)
		return nil
	case *ast.UnaryExpr:
		// if we have a reading from channel
		if n.Op == token.ARROW {
			fmt.Println("receive from channel: ", n.OpPos)
		}

		return nil
	}

	return f
}
