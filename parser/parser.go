package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type Parser struct {
	fset *token.FileSet
	pkgs map[string]*ast.Package
}

func (p *Parser) Parse(path string) error {
	pkgs, err := parser.ParseDir(p.fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	p.pkgs = pkgs
	return nil
}

func (p *Parser) GetFunctions() (functions []Function, err error) {
	for _, pkg := range p.pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.FuncDecl:
					name := n.Name.Name
					var args []FunctionArgument
					for _, param := range n.Type.Params.List {
						args = append(args, FunctionArgument{
							Name: param.Names[0].Name,
							Type: param.Type,
						})
					}
					functions = append(functions, Function{Name: name, Arguments: args})
				}
				return true
			})
		}
	}
	return functions, nil
}

func NewParser() *Parser {
	fset := token.NewFileSet()
	return &Parser{
		fset: fset,
	}
}
