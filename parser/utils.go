package parser

import (
	"go/ast"
	"go/types"
	"log"
	"regexp"
	"strconv"
	"strings"
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
		Package:     p.Service.Alias,
		ServiceType: p.Service.Type,
	}, true
}

func (p *Parser) extractArgsList(list *ast.FieldList) (args []Field) {
	if list == nil {
		return args
	}
	params := list.List
	for count, param := range params {
		currentType := types.ExprString(param.Type)
		log.Println("new type", currentType)
		var currentPackage string
		var typePrefix string

		//Detect local type with prefixes
		if ok, modifier := isExportedType(currentType); ok {
			slice := strings.SplitAfter(currentType, modifier)

			typePrefix = strings.Join(slice[0:len(slice)-1], "")
			currentType = slice[len(slice)-1]
			currentPackage = p.Service.Alias
		}

		if len(param.Names) != 0 {
			for _, name := range param.Names {
				args = append(args, Field{
					Name:    name.Name,
					Type:    currentType,
					Package: currentPackage,
					Prefix:  typePrefix,
				})
			}
		} else {
			args = append(args, Field{
				Name:    "arg" + strconv.Itoa(count),
				Type:    currentType,
				Package: currentPackage,
				Prefix:  typePrefix,
			})
		}
	}
	return args
}

func (p *Parser) processType(st *ast.StructType, ts *ast.TypeSpec) (*Type, bool) {
	t := &Type{
		Name:   ts.Name.Name,
		Fields: p.extractArgsList(st.Fields),
	}

	return t, true
}

func isExportedType(t string) (bool, string) {
	prefixes := []string{"[]*", "[]", "*"}
	for _, prefix := range prefixes {
		if ast.IsExported(strings.Trim(t, prefix)) {
			return true, prefix
		}
	}
	return false, ""
}
