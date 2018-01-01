package parser

import (
	"go/ast"
	"go/types"
	"log"
	"regexp"
	"strconv"
)

func (p *Parser) processFunction(n *ast.FuncDecl) (*Function, bool) {
	name := n.Name.Name
	if ok, err := regexp.MatchString("^[A-Z]", name); !ok || err != nil {
		return nil, false
	}
	args := p.extractArgsList(n.Type.Params)
	results := p.extractArgsList(n.Type.Results)

	return &Function{
		Name:      name,
		Arguments: args,
		Results:   results,
		Package:   p.Package.Alias,
	}, true
}

func (p *Parser) extractArgsList(list *ast.FieldList) (args []FunctionArgument) {
	if list == nil {
		return args
	}
	params := list.List
	for count, param := range params {
		currentType := types.ExprString(param.Type)
		if ast.IsExported(currentType) {
			currentType = p.Package.Alias + "." + currentType
		}
		log.Println(currentType)
		if len(param.Names) != 0 {
			for _, name := range param.Names {
				args = append(args, FunctionArgument{
					Name: name.Name,
					Type: currentType,
				})
			}
		} else {
			args = append(args, FunctionArgument{
				Name: "arg" + strconv.Itoa(count),
				Type: currentType,
			})
		}
	}
	return args
}

//TODO STUB
func (p *Parser) processType(n *ast.StructType) (*Type, bool) {
	return &Type{
		Name: "NewType",
		Field: Field{
			Name: "Ok",
			Type: "bool",
		},
	}, true
}
