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
		Name:        name,
		Arguments:   args,
		Results:     results,
		Package:     p.Package.Alias,
		ServiceType: p.ServiceType,
	}, true
}

func (p *Parser) extractArgsList(list *ast.FieldList) (args []Field) {
	if list == nil {
		return args
	}
	params := list.List
	for count, param := range params {
		currentType := types.ExprString(param.Type)
		var currentPackage string
		if ast.IsExported(currentType) {
			currentType = currentType
			currentPackage = p.Package.Alias
		}
		if len(param.Names) != 0 {
			for _, name := range param.Names {
				args = append(args, Field{
					Name:    name.Name,
					Type:    currentType,
					Package: currentPackage,
				})
			}
		} else {
			args = append(args, Field{
				Name:    "arg" + strconv.Itoa(count),
				Type:    currentType,
				Package: currentPackage,
			})
		}
	}
	log.Println(args)
	return args
}

//TODO STUB
//name works now fields
func (p *Parser) processType(st *ast.StructType, ts *ast.TypeSpec) (*Type, bool) {
	t := &Type{
		Name:   ts.Name.Name,
		Fields: p.extractArgsList(st.Fields),
	}
	return t, true
}
