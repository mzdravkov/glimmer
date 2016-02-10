package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/tsuna/gorewrite"
)

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

		// prune search if there is no line in the comment group that is a glimmer annotation
		isAnnotated := false
		for _, v := range n.Doc.List {
			if v.Text == "// glimmer" || v.Text == "//glimmer" {
				isAnnotated = true
				break
			}
		}
		if !isAnnotated {
			return nil
		}

		fmt.Println(n.Name)

		chanOperationsRewriter := new(ChanOperationsRewriter)
		gorewrite.Rewrite(chanOperationsRewriter, n.Body)

		return nil // Prune search
	}

	return nil // Prune search
}

type ChanOperationsRewriter struct{}

func (f *ChanOperationsRewriter) Rewrite(node ast.Node) (ast.Node, gorewrite.Rewriter) {
	switch n := node.(type) {
	case *ast.SendStmt:
		fmt.Println("send to channel: ", n)
		return AddSendStmt(n), nil
	case *ast.UnaryExpr:
		// if we have a reading from channel
		if n.Op == token.ARROW {
			fmt.Println("receive from channel: ", n)
			return AddRecvExpr(n), nil
		}
		//TODO: make a special case for result, ok := <-ch
	case *ast.CallExpr:
		// TODO: investigate whether there isn't a better approach to do this than sprintf
		switch fmt.Sprintf("%s", n.Fun) {
		case "make":
			return RewriteMakeCall(n), nil
		case "len":
			return RewriteLenCall(n), nil
		case "cap":
			return RewriteCapCall(n), nil
		case "close":
			return RewriteCloseCall(n), nil
		}
	}
	return node, f
}

func RewriteMakeCall(makeCall *ast.CallExpr) *ast.CallExpr {
	MakeChanGuardExpr, err := parser.ParseExpr("glimmer.MakeChanGuard")
	if err != nil {
		panic("Can't create expression for calling MakeChanGuard")
	}

	makeChanGuardCall := &ast.CallExpr{
		Fun:  MakeChanGuardExpr,
		Args: []ast.Expr{makeCall},
	}

	return makeChanGuardCall
}

func RewriteLenCall(lenCall *ast.CallExpr) *ast.CallExpr {
	expr := fmt.Sprintf("%s.Len", lenCall.Args[0])
	newLenCall, err := parser.ParseExpr(expr)
	if err != nil {
		fmt.Println("Can't create a new len() call:")
		panic(err)
	}

	return &ast.CallExpr{
		Fun: newLenCall,
	}
}

func RewriteCapCall(capCall *ast.CallExpr) *ast.CallExpr {
	expr := fmt.Sprintf("%s.Cap", capCall.Args[0])
	newCapCall, err := parser.ParseExpr(expr)
	if err != nil {
		fmt.Println("Can't create a new cap() call:")
		panic(err)
	}

	return &ast.CallExpr{
		Fun: newCapCall,
	}
}

func RewriteCloseCall(closeCall *ast.CallExpr) *ast.CallExpr {
	expr := fmt.Sprintf("%s.Close", closeCall.Args[0])
	newCloseCall, err := parser.ParseExpr(expr)
	if err != nil {
		fmt.Println("Can't create a new close() call:")
		panic(err)
	}

	return &ast.CallExpr{
		Fun: newCloseCall,
	}
}

func AddSendStmt(sendStmt *ast.SendStmt) *ast.ExprStmt {
	expr := fmt.Sprintf("%s.Send", sendStmt.Chan)
	sendFunc, err := parser.ParseExpr(expr)
	if err != nil {
		panic("can't parse glimmer.Send expression")
	}

	callSendExpression := &ast.CallExpr{
		Fun:  sendFunc,
		Args: []ast.Expr{sendStmt.Value},
	}

	return &ast.ExprStmt{X: callSendExpression}
}

func AddRecvExpr(recvExpr *ast.UnaryExpr) ast.Expr {
	expr := fmt.Sprintf("%s.Recieve", recvExpr.X)
	recieveFunc, err := parser.ParseExpr(expr)
	if err != nil {
		panic("can't parse glimmer.Receive expression")
	}

	return &ast.CallExpr{
		Fun: recieveFunc,
	}
}

func AddGlimmerImports(fset *token.FileSet, packages map[string]*ast.Package) {
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			astutil.AddNamedImport(fset, file, "glimmer", "github.com/mzdravkov/glimmer/inject")
		}
	}
}
