package parser

import (
	"go/ast"
	"go/types"
	"regexp"
	"strings"
)

func processFunction(n *ast.FuncDecl) (*Function, bool) {
	name := n.Name.Name
	if ok, err := regexp.MatchString("^[A-Z]", name); !ok || err != nil {
		return nil, false
	}
	args := extractArgsList(n.Type.Params)
	results := extractArgsList(n.Type.Results)
	imports := extractImports(n)

	return &Function{
		Name:      name,
		Arguments: args,
		Results:   results,
		Imports:   imports,
	}, true
}

func extractArgsList(list *ast.FieldList) (args []FunctionArgument) {
	if list == nil {
		return args
	}
	params := list.List
	for _, param := range params {
		var paramName string
		if len(param.Names) != 0 {
			var names []string
			for _, name := range param.Names {
				names = append(names, name.Name)
			}
			paramName = strings.Join(names, ", ")
		}
		args = append(args, FunctionArgument{
			Name: paramName,
			Type: types.ExprString(param.Type),
		})
	}
	return args
}

func extractImports(decl *ast.FuncDecl) (imports []string) {
	return imports
}
