package main

import (
	"fmt"
	"github.com/tsuna/gorewrite"
	"go/ast"
	"go/parser"
	"go/token"
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
	switch n := node.(type) {
	case *ast.SendStmt:
		fmt.Println("send to channl: ", n)
		return AddSendExprStmt(n), nil
	case *ast.UnaryExpr:
		// if we have a reading from channel
		if n.Op == token.ARROW {
			fmt.Println("receive from channel: ", n)
			return AddRecvExprStmt(n), nil
		}
		//TODO: make a special case for result, ok := <-ch
	}
	return node, f
}

func AddSendExprStmt(sendStmt *ast.SendStmt) *ast.ExprStmt {
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

func AddRecvExprStmt(recvExpr *ast.UnaryExpr) ast.Expr {
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
