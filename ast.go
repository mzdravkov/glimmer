package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"log"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/tsuna/gorewrite"
)

var annotatedFunctions []string = make([]string, 0)

type FuncDeclFinder struct{}

// Visit implements the ast.Visitor interface.
// it searches for function declarations with a glimmer annotation
func (f *FuncDeclFinder) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.File:
		return f
	case *ast.FuncDecl: // if it is a function declaration
		if _, ok := flags["examine-all"]; !ok {
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
		}

		annotatedFunctions = append(annotatedFunctions, n.Name.Name)

		chanOperationsRewriter := new(ChanOperationsRewriter)
		gorewrite.Rewrite(chanOperationsRewriter, n.Body)

		return nil // Prune search
	}

	return nil // Prune search
}

type ChanOperationsRewriter struct{}

// Rewrite implements the gorewrite.Rewriter interface.
// it should be called for a function body node and it
// searches for send and recieve statement and rewrites them
// to log the event of communication
func (r *ChanOperationsRewriter) Rewrite(node ast.Node) (ast.Node, gorewrite.Rewriter) {
	switch n := node.(type) {
	case *ast.SendStmt:
		fmt.Println("send to channel: ", n)
		fmt.Println(info.TypeOf(n.Chan))
		return AddSendStmt(n), nil
	case *ast.UnaryExpr:
		// if we have a reading from channel
		if n.Op == token.ARROW {
			fmt.Println("receive from channel: ", n)
			fmt.Println(info.TypeOf(n.X))
			return AddRecvExpr(n), nil
		}
	case *ast.AssignStmt: // case for value, ok := <-ch
		if len(n.Lhs) != 2 {
			return node, nil
		}

		if len(n.Rhs) != 1 {
			return node, nil
		}

		switch n.Rhs[0].(type) {
		case *ast.UnaryExpr:
			return AddRecvWithBoolAssignStmt(n), nil
		default:
			return node, nil
		}
	}

	return node, r
}

// This returns the glimmer substitute of a send statement
func AddSendStmt(sendStmt *ast.SendStmt) *ast.ExprStmt {
	callSendExpression := &ast.CallExpr{
		Fun:  createSendFunc(&sendStmt.Chan, &sendStmt.Value),
		Args: []ast.Expr{sendStmt.Chan, sendStmt.Value},
	}

	return &ast.ExprStmt{
		X: callSendExpression,
	}
}

// This returns the glimmer substitute of a recieve expression
func AddRecvExpr(recvExpr *ast.UnaryExpr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  createRecieveFunc(&recvExpr.X),
		Args: []ast.Expr{recvExpr.X},
	}
}

// This returns the glimmer substitute of a recieve with bool assignment statement
func AddRecvWithBoolAssignStmt(recvAssignStmt *ast.AssignStmt) *ast.AssignStmt {
	unaryChanExpr := recvAssignStmt.Rhs[0].(*ast.UnaryExpr)
	return &ast.AssignStmt{
		Lhs: recvAssignStmt.Lhs,
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun:  createRecieveWithBoolFunc(&unaryChanExpr.X),
				Args: []ast.Expr{unaryChanExpr.X},
			},
		},
	}
}

// Adds the glimmer runtime import to each file in the provided packages
func AddGlimmerImports(fset *token.FileSet, packages map[string]*ast.Package) {
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			astutil.AddNamedImport(fset, file, "glimmer", "github.com/mzdravkov/glimmer/inject")
		}
	}
}

// Creates a function that serves as a substitute for a recieve expression
func createRecieveFunc(ch *ast.Expr) *ast.FuncLit {
	chType := info.TypeOf(*ch)
	if chType == nil {
		log.Fatal("Can't get the type of a channel in a recieve expression")
	}

	funcType := createRecieveFuncType(chType, false)

	assignStmtRhs, err := parser.ParseExpr("<-ch")
	if err != nil {
		panic("Can't parse rhs expression for an assignment stmt inside recieve function")
	}

	processRecieveFunc, err := parser.ParseExpr("glimmer.ProcessRecieve")
	if err != nil {
		panic("Can't parse callProcessRecieveFunc")
	}

	chExpr, assignStmtLhs, sleepFunc, reflectValueOf, locksExpr := getCommonExpressions()

	body := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ExprStmt{ // glimmer.Locks[ch].Recieve.Lock()
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{ // glimmer.Locks[ch].Recieve.Lock
						X: &ast.SelectorExpr{ // glimmer.Locks[ch].Recieve
							X:   locksExpr,
							Sel: &ast.Ident{Name: "Recieve"},
						},
						Sel: &ast.Ident{Name: "Lock"},
					},
				},
			},
			&ast.ExprStmt{ // glimmer.Sleep()
				X: &ast.CallExpr{
					Fun: sleepFunc,
				},
			},
			&ast.AssignStmt{ // value := <-ch
				Lhs: []ast.Expr{assignStmtLhs},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{assignStmtRhs},
			},
			&ast.ExprStmt{ // glimmer.ProcessRecieve(reflect.ValueOf(ch).Pointer(), value)
				X: &ast.CallExpr{
					Fun: processRecieveFunc,
					Args: []ast.Expr{
						&ast.CallExpr{ // reflect.ValueOf(ch).Pointer()
							Fun: &ast.SelectorExpr{ // reflect.ValueOf(ch).Pointer
								X: &ast.CallExpr{ // reflect.ValueOf(ch)
									Fun:  reflectValueOf,
									Args: []ast.Expr{chExpr},
								},
								Sel: &ast.Ident{Name: "Pointer"},
							},
						},
						assignStmtLhs,
					},
				},
			},
			&ast.ReturnStmt{ //return value
				Results: []ast.Expr{assignStmtLhs},
			},
		},
	}

	return &ast.FuncLit{
		Type: funcType,
		Body: body,
	}
}

// Creates a function that serves as a substitute for a recieve with bool expression
func createRecieveWithBoolFunc(ch *ast.Expr) *ast.FuncLit {
	chType := info.TypeOf(*ch)
	if chType == nil {
		log.Fatal("Can't get the type of a channel in a recieve expression")
	}

	funcType := createRecieveFuncType(chType, true)

	assignStmtLhsOk, err := parser.ParseExpr("ok")
	if err != nil {
		panic("Can't parse lhs 'ok' expression for an assignment stmt inside recieve function")
	}

	assignStmtRhs, err := parser.ParseExpr("<-ch")
	if err != nil {
		panic("Can't parse rhs expression for an assignment stmt inside recieve function")
	}

	processRecieveFunc, err := parser.ParseExpr("glimmer.ProcessRecieve")
	if err != nil {
		panic("Can't parse ProcessRecieveFunc")
	}

	chExpr, assignStmtLhsValue, sleepFunc, reflectValueOf, locksExpr := getCommonExpressions()

	body := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ExprStmt{ // glimmer.Locks[ch].Recieve.Lock()
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{ // glimmer.Locks[ch].Recieve.Lock
						X: &ast.SelectorExpr{ // glimmer.Locks[ch].Recieve
							X:   locksExpr,
							Sel: &ast.Ident{Name: "Recieve"},
						},
						Sel: &ast.Ident{Name: "Lock"},
					},
				},
			},
			&ast.ExprStmt{ // glimmer.Sleep()
				X: &ast.CallExpr{
					Fun: sleepFunc,
				},
			},
			&ast.AssignStmt{ // value, ok := <-ch
				Lhs: []ast.Expr{assignStmtLhsValue, assignStmtLhsOk},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{assignStmtRhs},
			},
			&ast.ExprStmt{ // glimmer.ProcessRecieve(reflect.ValueOf(ch).Pointer(), value)
				X: &ast.CallExpr{
					Fun: processRecieveFunc,
					Args: []ast.Expr{
						&ast.CallExpr{ // reflect.ValueOf(ch).Pointer()
							Fun: &ast.SelectorExpr{ // reflect.ValueOf(ch).Pointer
								X: &ast.CallExpr{ // reflect.ValueOf(ch)
									Fun:  reflectValueOf,
									Args: []ast.Expr{chExpr},
								},
								Sel: &ast.Ident{Name: "Pointer"},
							},
						},
						assignStmtLhsValue,
					},
				},
			},
			&ast.ReturnStmt{ // return value, ok
				Results: []ast.Expr{assignStmtLhsValue, assignStmtLhsOk},
			},
		},
	}

	return &ast.FuncLit{
		Type: funcType,
		Body: body,
	}
}

// Creates an ast.FuncType with one argument, which is a channel type and
// the return results are the element type of the channel and (if withBool is true)
// a boolean value
func createRecieveFuncType(chType types.Type, withBool bool) *ast.FuncType {
	paramType, err := parser.ParseExpr(fmt.Sprintf("%s", chType))
	if err != nil {
		panic("Can't parse channel type for the parameter of a recieve function")
	}
	params := &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Names: []*ast.Ident{&ast.Ident{Name: "ch"}},
				Type:  paramType,
			},
		},
	}

	// TODO: when I was using reflect.Type instead of types.Type I could use
	// chType.Elem() to get the type of the values for the channel.
	// types.Type doesn't have such method and for now I can't find a better way
	// to do this than to get the string representation and to remove the first
	// "chan " from the string. It's highly recommended to find a good
	// and not so hacky way to do this.
	resultType, err := parser.ParseExpr(fmt.Sprintf("%s", chType.String()[5:]))
	if err != nil {
		panic("Can't parse channel element type for the return type of a recieve function")
	}
	results := &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Names: []*ast.Ident{&ast.Ident{}},
				Type:  resultType,
			},
		},
	}

	if withBool {
		boolType, err := parser.ParseExpr("bool")
		if err != nil {
			panic("Can't parse bool type")
		}
		okResult := &ast.Field{
			Names: []*ast.Ident{&ast.Ident{}},
			Type:  boolType,
		}
		results.List = append(results.List, okResult)
	}

	return &ast.FuncType{
		Params:  params,
		Results: results,
	}
}

// Creates a function that serves as a substitute for a send statement
func createSendFunc(ch, value *ast.Expr) *ast.FuncLit {
	chType := info.TypeOf(*ch)
	if chType == nil {
		log.Fatal("Can't get the type of a channel in a send statement")
	}

	valueType := info.TypeOf(*value)
	if valueType == nil {
		log.Fatal("Can't get the type of a value in a send statement")
	}

	chanParamType, err := parser.ParseExpr(fmt.Sprintf("%s", chType))
	if err != nil {
		panic("Can't parse channel type for the parameter of a send function")
	}

	valueParamType, err := parser.ParseExpr(fmt.Sprintf("%s", valueType))
	if err != nil {
		panic("Can't parse value type for the paramater of a send function")
	}

	params := &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Names: []*ast.Ident{&ast.Ident{Name: "ch"}},
				Type:  chanParamType,
			},
			&ast.Field{
				Names: []*ast.Ident{&ast.Ident{Name: "value"}},
				Type:  valueParamType,
			},
		},
	}

	results := &ast.FieldList{
		List: []*ast.Field{},
	}

	funcType := &ast.FuncType{
		Params:  params,
		Results: results,
	}

	processSendFunc, err := parser.ParseExpr("glimmer.ProcessSend")
	if err != nil {
		panic("Can't parse glimmer.ProcessSend reference")
	}

	chExpr, valueExpr, sleepFunc, reflectValueOf, locksExpr := getCommonExpressions()

	body := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ExprStmt{ // glimmer.Locks[ch].Send.Lock()
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{ // glimmer.Locks[ch].Send.Lock
						X: &ast.SelectorExpr{ // glimmer.Locks[ch].Send
							X:   locksExpr,
							Sel: &ast.Ident{Name: "Send"},
						},
						Sel: &ast.Ident{Name: "Lock"},
					},
				},
			},
			&ast.ExprStmt{ // glimmer.Sleep()
				X: &ast.CallExpr{
					Fun: sleepFunc,
				},
			},
			&ast.SendStmt{ // ch <- value
				Chan:  chExpr,
				Value: valueExpr,
			},
			&ast.ExprStmt{ // glimmer.ProcessRecieve(reflect.ValueOf(ch).Pointer(), value)
				X: &ast.CallExpr{
					Fun: processSendFunc,
					Args: []ast.Expr{
						&ast.CallExpr{ // reflect.ValueOf(ch).Pointer()
							Fun: &ast.SelectorExpr{ // reflect.ValueOf(ch).Pointer
								X: &ast.CallExpr{ // reflect.ValueOf(ch)
									Fun:  reflectValueOf,
									Args: []ast.Expr{chExpr},
								},
								Sel: &ast.Ident{Name: "Pointer"},
							},
						},
						valueExpr,
					},
				},
			},
		},
	}

	return &ast.FuncLit{
		Type: funcType,
		Body: body,
	}
}

func getCommonExpressions() (ast.Expr, ast.Expr, ast.Expr, ast.Expr, ast.Expr) {
	chExpr, err := parser.ParseExpr("ch")
	if err != nil {
		panic("Can't parse chan expression")
	}

	valueExpr, err := parser.ParseExpr("value")
	if err != nil {
		panic("Can't parse value expression")
	}

	sleepFunc, err := parser.ParseExpr("glimmer.Sleep")
	if err != nil {
		panic("Can't parse glimmer.Sleep expression")
	}

	reflectValueOf, err := parser.ParseExpr("reflect.ValueOf")
	if err != nil {
		panic("Can't parse reflect.ValueOf expression")
	}

	locksExpr, err := parser.ParseExpr("glimmer.Locks(reflect.ValueOf(ch).Pointer())")
	if err != nil {
		panic("Can't parse glimmer.Locks expression")
	}

	return chExpr, valueExpr, sleepFunc, reflectValueOf, locksExpr
}
