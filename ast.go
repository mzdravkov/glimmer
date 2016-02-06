package main

import (
	"fmt"
	"github.com/tsuna/gorewrite"
	"go/ast"
	"go/parser"
	"go/token"
	// "reflect"
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

// Visit implements the ast.Visitor interface.
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
	MakeChanGuardExpr, err := parser.ParseExpr("MakeChanGuard")
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
	newArgument, err := parser.ParseExpr(fmt.Sprintf("%s.%s", lenCall.Args[0], "Chan"))
	if err != nil {
		panic("Can't create new argument for a len() call")
	}

	lenCall.Args[0] = newArgument

	return lenCall
}

func RewriteCapCall(capCall *ast.CallExpr) *ast.CallExpr {
	newArgument, err := parser.ParseExpr(fmt.Sprintf("%s.%s", capCall.Args[0], "Chan"))
	if err != nil {
		panic("Can't create new argument for a cap() call")
	}

	capCall.Args[0] = newArgument

	return capCall
}

func RewriteCloseCall(closeCall *ast.CallExpr) *ast.CallExpr {
	newArgument, err := parser.ParseExpr(fmt.Sprintf("%s.%s", closeCall.Args[0], "Chan"))
	if err != nil {
		panic("Can't create new argument for a cap() call")
	}

	closeCall.Args[0] = newArgument

	return closeCall
}

func AddSendStmt(sendStmt *ast.SendStmt) *ast.ExprStmt {
	sendFunc, err := parser.ParseExpr("glimmer.Send")
	if err != nil {
		panic("can't parse glimmer.Send expression")
	}

	callSendExpression := &ast.CallExpr{
		Fun:  sendFunc,
		Args: []ast.Expr{sendStmt.Chan, sendStmt.Value},
	}

	return &ast.ExprStmt{X: callSendExpression}
}

func AddRecvExpr(recvExpr *ast.UnaryExpr) ast.Expr {
	recieveFunc, err := parser.ParseExpr("glimmer.Recieve")
	if err != nil {
		panic("can't parse glimmer.Receive expression")
	}

	callRecieveExpression := &ast.CallExpr{
		Fun:  recieveFunc,
		Args: []ast.Expr{recvExpr.X},
	}

	return callRecieveExpression
}
