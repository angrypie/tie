package parser

import "go/ast"

type FunctionArgument struct {
	Name string
	Type ast.Expr
}

type Function struct {
	Name      string
	Arguments []FunctionArgument
}
