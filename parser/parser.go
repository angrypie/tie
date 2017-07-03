package parser

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type Parser struct {
	fset        *token.FileSet
	pkgs        map[string]*ast.Package
	Package     string
	PackagePath string
}

func NewParser() *Parser {
	fset := token.NewFileSet()
	return &Parser{
		fset: fset,
	}
}

func (p *Parser) Parse(pkg string) error {
	p.Package = pkg
	p.PackagePath = fmt.Sprintf("%s/src/%s", build.Default.GOPATH, pkg)
	pkgs, err := parser.ParseDir(p.fset, p.PackagePath, func(info os.FileInfo) bool {
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
