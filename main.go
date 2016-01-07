package main

import (
	"fmt"
	"github.com/tsuna/gorewrite"
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
		sendOrReceiveRewriter := new(SendOrReceiveRewriter)
		gorewrite.Rewrite(sendOrReceiveRewriter, n.Body)
		return nil // Prune search
	}

	return nil // Prune search
}

type SendOrReceiveRewriter struct{}

// Visit implements the ast.Visitor interface.
func (f *SendOrReceiveRewriter) Rewrite(node ast.Node) (ast.Node, gorewrite.Rewriter) {
	// if node == nil {
	// 	return nil, nil
	// }

	switch n := node.(type) {
	case *ast.SendStmt:
		fmt.Println("send to channl: ", n)
		return AddSendCallExpr(n), nil
	case *ast.UnaryExpr:
		// if we have a reading from channel
		if n.Op == token.ARROW {
			fmt.Println("receive from channel: ", n.OpPos)
		}
	}
	return node, f
}

func AddSendCallExpr(sendStmt *ast.SendStmt) *ast.CallExpr {
	sendFunc, err := parser.ParseExpr("glimmer.Send")
	if err != nil {
		panic("can't parse AddChan expression")
	}

	return &ast.CallExpr{
		Fun:  sendFunc,
		Args: []ast.Expr{sendStmt.Chan, sendStmt.Value},
	}
}

// func AddChanCallExpr(sendStmt *ast.SendStmt) *ast.CallExpr {
// 	addChanFunc, err := parser.ParseExpr("AddChan")
// 	if err != nil {
// 		panic("can't parse AddChan expression")
// 	}

// 	return &ast.CallExpr{
// 		Fun:  addChanFunc,
// 		Args: []Expr{sendStmt.Chan},
// 	}
// }

// func AddSendExprStmt(sendStmt *ast.SendStmt) *ast.ExprStmt {

// }
