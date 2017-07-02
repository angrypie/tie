package parser

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
)

type Parser struct {
	fset *token.FileSet
	pkgs map[string]*ast.Package
	Path string
}

func (p *Parser) Parse(path string) error {
	p.Path = fmt.Sprintf("%s/src/%s", build.Default.GOPATH, path)
	pkgs, err := parser.ParseDir(p.fset, p.Path, func(info os.FileInfo) bool {
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

func (p *Parser) GetFunctions() (functions []*Function, err error) {
	for _, pkg := range p.pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.FuncDecl:
					if function, ok := processFunction(n); ok {
						functions = append(functions, function)
					}
				}
				return true
			})
		}
	}
	return functions, nil
}

func processFunction(n *ast.FuncDecl) (*Function, bool) {
	name := n.Name.Name
	if ok, err := regexp.MatchString("^[A-Z]", name); !ok || err != nil {
		return nil, false
	}
	var args []FunctionArgument
	for _, param := range n.Type.Params.List {
		paramName := param.Names[0].Name
		args = append(args, FunctionArgument{
			Name: paramName,
			Type: param.Type,
		})
	}
	return &Function{Name: name, Arguments: args}, true
}

func NewParser() *Parser {
	fset := token.NewFileSet()
	return &Parser{
		fset: fset,
	}
}
