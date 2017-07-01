package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

type Parser struct {
	fset *token.FileSet
	pkgs map[string]*ast.Package
}

func (p *Parser) Parse(path string) error {
	pkgs, err := parser.ParseDir(p.fset, path, func(info os.FileInfo) bool {
		name := info.Name()
		return !info.IsDir() &&
			!strings.HasPrefix(name, ".") &&
			strings.HasSuffix(name, ".go") &&
			!strings.HasSuffix(name, "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return err
	}
	p.pkgs = pkgs
	return nil
}

func (p *Parser) GetFunctions() (functions []Function, err error) {
	for name, pkg := range p.pkgs {
		log.Println("Found package ", name)
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
